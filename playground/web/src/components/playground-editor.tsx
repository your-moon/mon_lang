"use client";

import Editor, { type OnMount } from "@monaco-editor/react";
import { useCallback, useEffect, useRef } from "react";
import { useTheme } from "next-themes";

interface PlaygroundEditorProps {
  code: string;
  onChange: (value: string) => void;
  onRun: () => void;
}

function registerMonLanguage(monaco: Parameters<OnMount>[1]) {
  monaco.languages.register({ id: "mon" });

  monaco.languages.setMonarchTokensProvider("mon", {
    keywords: [
      "функц",
      "зарла",
      "хэрэв",
      "бол",
      "эсвэл",
      "давтах",
      "буц",
      "хоосон",
      "тоо",
      "тоо64",
      "мөр",
      "шинэ",
      "зогс",
      "үргэлжлүүл",
      "тунх",
      "extern",
      "ашигла",
      "үнэн",
      "худал",
      "бүтэц",
      "массив",
    ],
    operators: [
      "=",
      ">",
      "<",
      "!",
      "==",
      "!=",
      "<=",
      ">=",
      "&&",
      "||",
      "+",
      "-",
      "*",
      "/",
      "%",
    ],
    symbols: /[=><!~?:&|+\-*/^%]+/,
    escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4})/,

    tokenizer: {
      root: [
        [
          /[a-zA-Z\u0400-\u04FF_][\w\u0400-\u04FF]*/,
          {
            cases: {
              "@keywords": "keyword",
              "@default": "identifier",
            },
          },
        ],
        { include: "@whitespace" },
        [/[{}()[\]]/, "@brackets"],
        [
          /@symbols/,
          {
            cases: {
              "@operators": "operator",
              "@default": "",
            },
          },
        ],
        [/\d*\.\d+([eE][-+]?\d+)?/, "number.float"],
        [/\d+/, "number"],
        [/"([^"\\]|\\.)*$/, "string.invalid"],
        [/"/, { token: "string.quote", bracket: "@open", next: "@string" }],
      ],
      string: [
        [/[^\\"]+/, "string"],
        [/@escapes/, "string.escape"],
        [/\\./, "string.escape.invalid"],
        [/"/, { token: "string.quote", bracket: "@close", next: "@pop" }],
      ],
      whitespace: [
        [/[ \t\r\n]+/, "white"],
        [/\/\/.*$/, "comment"],
      ],
    },
  });

  monaco.editor.defineTheme("mon-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "c586c0", fontStyle: "bold" },
      { token: "comment", foreground: "6a9955" },
      { token: "string", foreground: "ce9178" },
      { token: "number", foreground: "b5cea8" },
      { token: "operator", foreground: "d4d4d4" },
      { token: "identifier", foreground: "9cdcfe" },
    ],
    colors: {
      "editor.background": "#1e1e24",
      "editor.foreground": "#d4d4d4",
      "editorLineNumber.foreground": "#555560",
      "editorLineNumber.activeForeground": "#b0b0b8",
      "editor.selectionBackground": "#264f78",
      "editor.lineHighlightBackground": "#25252c",
    },
  });

  monaco.editor.defineTheme("mon-light", {
    base: "vs",
    inherit: true,
    rules: [
      { token: "keyword", foreground: "7c3aed", fontStyle: "bold" },
      { token: "comment", foreground: "6b7280" },
      { token: "string", foreground: "b45309" },
      { token: "number", foreground: "0d9488" },
      { token: "operator", foreground: "374151" },
      { token: "identifier", foreground: "1d4ed8" },
    ],
    colors: {
      "editor.background": "#f5f5f0",
      "editor.foreground": "#2e2e32",
      "editorLineNumber.foreground": "#a0a0a8",
      "editorLineNumber.activeForeground": "#505058",
      "editor.selectionBackground": "#d0dff0",
      "editor.lineHighlightBackground": "#eeeeea",
    },
  });
}

export default function PlaygroundEditor({
  code,
  onChange,
  onRun,
}: PlaygroundEditorProps) {
  const editorRef = useRef<Parameters<OnMount>[0] | null>(null);
  const onRunRef = useRef(onRun);
  const { resolvedTheme } = useTheme();

  useEffect(() => {
    onRunRef.current = onRun;
  }, [onRun]);

  useEffect(() => {
    if (editorRef.current) {
      const monaco = (window as any).monaco;
      if (monaco) {
        monaco.editor.setTheme(resolvedTheme === "dark" ? "mon-dark" : "mon-light");
      }
    }
  }, [resolvedTheme]);

  const handleMount: OnMount = useCallback(
    (editor, monaco) => {
      editorRef.current = editor;
      registerMonLanguage(monaco);
      const initialTheme = resolvedTheme === "dark" ? "mon-dark" : "mon-light";
      editor.updateOptions({ theme: initialTheme });
      monaco.editor.setTheme(initialTheme);

      editor.addAction({
        id: "run-code",
        label: "Run Code",
        keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
        run: () => {
          onRunRef.current();
        },
      });

      editor.focus();
    },
    [resolvedTheme]
  );

  return (
    <Editor
      height="100%"
      defaultLanguage="mon"
      language="mon"
      theme={resolvedTheme === "dark" ? "mon-dark" : "mon-light"}
      value={code}
      onChange={(value) => onChange(value || "")}
      onMount={handleMount}
      loading={
        <div className="flex items-center justify-center h-full text-muted-foreground">
          Засварлагч ачааллаж байна...
        </div>
      }
      options={{
        fontFamily: "var(--font-mono), 'JetBrains Mono', monospace",
        fontSize: 14,
        lineHeight: 1.6,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        padding: { top: 16, bottom: 16 },
        renderLineHighlight: "line",
        cursorBlinking: "smooth",
        smoothScrolling: true,
        wordWrap: "on",
        automaticLayout: true,
        tabSize: 4,
        bracketPairColorization: { enabled: true },
        unicodeHighlight: {
          ambiguousCharacters: false,
          invisibleCharacters: false,
          nonBasicASCII: false,
        },
      }}
    />
  );
}
