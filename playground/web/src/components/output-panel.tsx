"use client";

import { AlertCircle, Terminal } from "lucide-react";

interface OutputPanelProps {
  stdout: string;
  stderr: string;
  error: string;
  isRunning: boolean;
}

export default function OutputPanel({
  stdout,
  stderr,
  error,
  isRunning,
}: OutputPanelProps) {
  const hasOutput = stdout || stderr || error;

  if (isRunning) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground gap-2 dark:bg-[#0a0a0a] bg-gray-50">
        <svg
          className="animate-spin h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <span>Ажиллаж байна...</span>
      </div>
    );
  }

  if (!hasOutput) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2 dark:bg-[#0a0a0a] bg-gray-50">
        <Terminal className="h-8 w-8 opacity-40" />
        <p className="text-sm">
          &quot;Ажиллуулах&quot; товчийг дарж үр дүнг харна уу
        </p>
        <p className="text-xs opacity-60">Ctrl+Enter</p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto p-4 font-mono text-sm leading-relaxed dark:bg-[#0a0a0a] bg-gray-50">
      {error && (
        <div className="flex items-start gap-2 text-red-500 dark:text-red-400 mb-2">
          <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
          <pre className="whitespace-pre-wrap break-words">{error}</pre>
        </div>
      )}
      {stderr && (
        <pre className="text-red-500 dark:text-red-400 whitespace-pre-wrap break-words">
          {stderr}
        </pre>
      )}
      {stdout && (
        <pre className="text-foreground whitespace-pre-wrap break-words">
          {stdout}
        </pre>
      )}
    </div>
  );
}
