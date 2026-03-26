"use client";

import { useState, useEffect, useCallback } from "react";
import dynamic from "next/dynamic";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Play, Loader2 } from "lucide-react";
import { runCode, getExamples, type Example, type RunResult } from "@/lib/api";
import OutputPanel from "@/components/output-panel";
import { ThemeToggle } from "@/components/theme-toggle";

const PlaygroundEditor = dynamic(
  () => import("@/components/playground-editor"),
  {
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        Засварлагч ачааллаж байна...
      </div>
    ),
  }
);

const DEFAULT_CODE = `// Mon Lang Playground
// Ctrl+Enter дарж ажиллуулна уу

функц үндсэн() -> тоо {
    мөр_хэвлэх("Сайн байна уу, Дэлхий!\\n");
    буц 0;
}
`;

export default function Home() {
  const [code, setCode] = useState(DEFAULT_CODE);
  const [output, setOutput] = useState<RunResult | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [examples, setExamples] = useState<Example[]>([]);
  const [selectedExample, setSelectedExample] = useState<string>("");
  const [executionTime, setExecutionTime] = useState<string>("");

  useEffect(() => {
    getExamples()
      .then((data) => {
        setExamples(data);
        // Load example from URL hash if present
        const hash = window.location.hash.slice(1);
        if (hash) {
          const example = data.find((e: Example) => e.filename === hash);
          if (example) {
            setCode(example.code);
            setSelectedExample(example.filename);
          }
        }
      })
      .catch(() => {});
  }, []);

  const handleRun = useCallback(async () => {
    if (isRunning) return;
    setIsRunning(true);
    setOutput(null);
    setExecutionTime("");

    try {
      const result = await runCode(code);
      setOutput(result);
      setExecutionTime(result.duration || "");
    } catch (err) {
      setOutput({
        stdout: "",
        stderr: "",
        exitCode: -1,
        duration: "",
        error:
          err instanceof Error
            ? err.message
            : "Серверт холбогдож чадсангүй",
      });
    } finally {
      setIsRunning(false);
    }
  }, [code, isRunning]);

  const handleExampleChange = (value: string | null) => {
    if (!value) return;
    setSelectedExample(value);
    window.location.hash = value;
    const example = examples.find((e) => e.filename === value);
    if (example) {
      setCode(example.code);
      setOutput(null);
      setExecutionTime("");
    }
  };

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <header className="flex items-center justify-between px-4 py-2 border-b border-border bg-card shrink-0">
        <div className="flex items-center gap-3">
          <h1 className="text-base font-semibold tracking-tight">
            Mon Lang Playground
          </h1>
          <Badge variant="secondary" className="text-xs hidden sm:inline-flex">
            v0.1
          </Badge>
        </div>
        <ThemeToggle />
      </header>

      {/* Toolbar */}
      <div className="flex items-center gap-3 px-4 py-1.5 border-b border-border bg-card/50 shrink-0">
        <Button
          onClick={handleRun}
          disabled={isRunning}
          size="sm"
          className="bg-emerald-600 hover:bg-emerald-700 text-white border-0 gap-1.5"
        >
          {isRunning ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
          ) : (
            <Play className="h-3.5 w-3.5" />
          )}
          Ажиллуулах
          <kbd className="hidden sm:inline-flex ml-1 text-[10px] opacity-70 bg-white/10 px-1 py-0.5 rounded">
            Ctrl+Enter
          </kbd>
        </Button>
        {examples.length > 0 && (
          <>
            <Separator orientation="vertical" className="h-4" />
            <Select value={selectedExample} onValueChange={handleExampleChange}>
              <SelectTrigger size="sm" className="w-[160px] text-xs">
                <SelectValue placeholder="Жишээ сонгох..." />
              </SelectTrigger>
              <SelectContent>
                {examples.map((example) => (
                  <SelectItem key={example.filename} value={example.filename}>
                    {example.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </>
        )}
        <Separator orientation="vertical" className="h-4" />
        {executionTime && (
          <span className="text-xs text-muted-foreground">
            {executionTime}
          </span>
        )}
        {output && !isRunning && (
          <Badge
            variant={
              output.exitCode === 0 && !output.error
                ? "secondary"
                : "destructive"
            }
            className="text-xs"
          >
            {output.exitCode === 0 && !output.error
              ? "Амжилттай"
              : `Алдаа (${output.exitCode})`}
          </Badge>
        )}
      </div>

      {/* Main content: editor + output */}
      <div className="flex flex-1 min-h-0 flex-col md:flex-row">
        {/* Editor */}
        <div className="flex-1 min-h-[300px] md:min-h-0 border-b md:border-b-0 md:border-r border-border">
          <PlaygroundEditor
            code={code}
            onChange={setCode}
            onRun={handleRun}
          />
        </div>

        {/* Output */}
        <div className="flex-1 min-h-[200px] md:min-h-0 md:max-w-[50%] dark:bg-[#1a1a20] bg-[#f0f0ec]">
          <OutputPanel
            stdout={output?.stdout || ""}
            stderr={output?.stderr || ""}
            error={output?.error || ""}
            isRunning={isRunning}
          />
        </div>
      </div>

      {/* Footer */}
      <footer className="flex items-center justify-between px-4 py-1.5 border-t border-border bg-card text-xs text-muted-foreground shrink-0">
        <span>Mon Lang - Монгол програмчлалын хэл</span>
        <a
          href="https://github.com/munkherdene-codes/mon_lang"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1 hover:text-foreground transition-colors"
        >
          <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
          GitHub
        </a>
      </footer>
    </div>
  );
}
