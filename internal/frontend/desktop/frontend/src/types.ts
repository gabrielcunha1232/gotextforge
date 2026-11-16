export type UITab = { Name: string; Path: string; Modified: boolean; Active: boolean };

export type UITreeRow = {
  Name: string;
  Path: string;
  Depth: number;
  Dir: boolean;
  Expanded: boolean;
  Empty: boolean;
  Current: boolean;
};

export type UIExplorer = { Open: boolean; Name: string; Rows: UITreeRow[] | null; Selected: number };

export type UITerminal = {
  Open: boolean;
  Command: string;
  Dir: string;
  State: "" | "running" | "ok" | "failed" | "canceled";
  ExitCode: number;
  Millis: number;
  First: number;
  Lines: string[] | null;
  Kinds: string;
  Selected: number;
  Follow: boolean;
};

export type DockerContainer = { ID: string; Name: string; Image: string; State: string; Status: string; Ports: string };
export type DockerService = { Name: string; Service: string; State: string; Status: string; Health: string; ExitCode: number; Ports: string };
export type DockerImage = { ID: string; Repository: string; Tag: string; Size: string };

export type UIDocker = {
  Open: boolean;
  Tab: string;
  Selected: number;
  Containers: DockerContainer[] | null;
  Services: DockerService[] | null;
  Images: DockerImage[] | null;
  Error: string;
};

export type UIProblem = { Path: string; Name: string; Line: number; Col: number; Severity: string; Message: string };

export type UIFinder = {
  Title: string;
  Query: string;
  Items: { Text: string; Positions: number[] | null }[] | null;
  Create: boolean;
  Selected: number;
  Total: number;
  Truncated: boolean;
};

export type UIStatus = {
  File: string;
  Path: string;
  Modified: boolean;
  Errors: number;
  Warnings: number;
  Language: string;
  LSP: "" | "starting" | "ready" | "failed";
  LineEnding: string;
  BOM: boolean;
  GoVersion: string;
};

export type UIState = {
  Mode: string;
  Focus: "editor" | "explorer" | "terminal" | "docker" | "finder" | "command";
  Root: string;
  Tabs: UITab[] | null;
  Explorer: UIExplorer;
  Bottom: "" | "terminal" | "docker";
  Terminal: UITerminal;
  Docker: UIDocker;
  Problems: UIProblem[] | null;
  Finder: UIFinder | null;
  Status: UIStatus;
};

export type Pos = { line: number; col: number; progress: string; pending?: string };
