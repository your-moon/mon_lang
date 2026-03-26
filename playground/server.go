package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed all:static
var staticFiles embed.FS

//go:embed examples/*
var exampleFiles embed.FS

const (
	maxCodeSize   = 64 * 1024 // 64KB
	maxOutputSize = 64 * 1024 // 64KB
	execTimeout   = 5 * time.Second
	rateLimit     = 30 // requests per minute per IP
)

type Server struct {
	compiler string
	stdlib   string
	mux      *http.ServeMux
	limiter  *RateLimiter
}

type RunRequest struct {
	Code string `json:"code"`
}

type RunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

type Example struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Code     string `json:"code"`
}

func NewServer(compiler, stdlib string) *Server {
	s := &Server{
		compiler: compiler,
		stdlib:   stdlib,
		mux:      http.NewServeMux(),
		limiter:  NewRateLimiter(rateLimit, time.Minute),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS for development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// API routes
	s.mux.HandleFunc("/api/run", s.handleRun)
	s.mux.HandleFunc("/api/examples", s.handleExamples)

	// Static files (Next.js export)
	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("static files error: %v", err)
	}
	s.mux.Handle("/", http.FileServer(http.FS(staticSub)))
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limit
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}
	if !s.limiter.Allow(ip) {
		writeJSON(w, http.StatusTooManyRequests, RunResponse{
			Error: "Хэт олон хүсэлт. Түр хүлээнэ үү.",
		})
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, maxCodeSize+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "read error"})
		return
	}
	if len(body) > maxCodeSize {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "code too large (max 64KB)"})
		return
	}

	var req RunRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "invalid JSON"})
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		writeJSON(w, http.StatusBadRequest, RunResponse{Error: "empty code"})
		return
	}

	resp := s.runCode(req.Code)
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) runCode(code string) RunResponse {
	start := time.Now()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "monplay-*")
	if err != nil {
		return RunResponse{Error: fmt.Sprintf("temp dir error: %v", err)}
	}
	defer os.RemoveAll(tmpDir)

	// Symlink stdlib
	if err := os.Symlink(s.stdlib, filepath.Join(tmpDir, "stdlib")); err != nil {
		return RunResponse{Error: fmt.Sprintf("stdlib symlink error: %v", err)}
	}

	// Symlink stdlib .mn files so imports like ашигла "math.mn" work
	entries, _ := os.ReadDir(s.stdlib)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".mn") && e.Name() != "prelude.mn" {
			os.Symlink(filepath.Join(s.stdlib, e.Name()), filepath.Join(tmpDir, e.Name()))
		}
	}

	// Write source
	srcFile := filepath.Join(tmpDir, "program.mn")
	if err := os.WriteFile(srcFile, []byte(code), 0644); err != nil {
		return RunResponse{Error: fmt.Sprintf("write error: %v", err)}
	}

	// Compile and run
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.compiler, "gen", "program.mn", "--run")
	cmd.Dir = tmpDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &stdout, limit: maxOutputSize}
	cmd.Stderr = &limitedWriter{buf: &stderr, limit: maxOutputSize}

	err = cmd.Run()
	duration := time.Since(start)

	resp := RunResponse{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: fmt.Sprintf("%.1fms", float64(duration.Milliseconds())),
	}

	if ctx.Err() == context.DeadlineExceeded {
		resp.Error = "Хугацаа хэтэрсэн (5 секунд)"
		resp.ExitCode = -1
	} else if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			resp.Error = err.Error()
			resp.ExitCode = -1
		}
	}

	return resp
}

func (s *Server) handleExamples(w http.ResponseWriter, r *http.Request) {
	entries, err := exampleFiles.ReadDir("examples")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read examples"})
		return
	}

	var examples []Example
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mn") {
			continue
		}
		data, err := exampleFiles.ReadFile("examples/" + entry.Name())
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".mn")
		name = strings.ReplaceAll(name, "_", " ")
		// Capitalize first letter
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + name[1:]
		}
		examples = append(examples, Example{
			Name:     name,
			Filename: entry.Name(),
			Code:     string(data),
		})
	}

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Filename < examples[j].Filename
	})

	writeJSON(w, http.StatusOK, examples)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// limitedWriter caps output at a byte limit
type limitedWriter struct {
	buf   *bytes.Buffer
	limit int64
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.limit - int64(lw.buf.Len())
	if remaining <= 0 {
		return len(p), nil // discard silently
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	return lw.buf.Write(p)
}

// RateLimiter tracks requests per IP
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old entries
	times := rl.requests[ip]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}
