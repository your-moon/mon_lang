const API_URL = process.env.NEXT_PUBLIC_API_URL || "";

export interface RunResult {
  stdout: string;
  stderr: string;
  exitCode: number;
  duration: string;
  error: string;
}

export interface Example {
  name: string;
  filename: string;
  code: string;
}

export async function runCode(code: string): Promise<RunResult> {
  const res = await fetch(`${API_URL}/api/run`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
  });

  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }

  return res.json();
}

export async function getExamples(): Promise<Example[]> {
  const res = await fetch(`${API_URL}/api/examples`);

  if (!res.ok) {
    throw new Error(`Server error: ${res.status}`);
  }

  return res.json();
}
