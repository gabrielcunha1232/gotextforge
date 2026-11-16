import { useSyncExternalStore } from "react";
import { posStore, problemsStore } from "../bridge";
import type { UIState } from "../types";

const LSP_LABEL: Record<string, string> = {
  starting: "gopls…",
  ready: "gopls",
  failed: "gopls!",
};

export function StatusBar({ ui }: { ui: UIState | null }) {
  const pos = useSyncExternalStore(posStore.subscribe, posStore.get);
  if (!ui) {
    return <footer className="status" />;
  }
  const s = ui.Status;
  const mode = ui.Mode.toLowerCase();

  return (
    <footer className="status">
      <span className="mode" style={{ background: `var(--c-mode-${mode})` }}>
        {ui.Mode}
      </span>
      <span>
        {s.File}
        {s.Modified && " ●"}
      </span>
      <span className="click" title="Problemas" onClick={() => problemsStore.set(true)}>
        <span className={s.Errors > 0 ? "err" : ""}>⊙ {s.Errors}</span>
        <span className={s.Warnings > 0 ? "warn" : ""}>△ {s.Warnings}</span>
      </span>
      {pos.pending && <span className="pending">{pos.pending}</span>}
      <span className="push">
        Ln {pos.line}, Col {pos.col}　{pos.progress}
      </span>
      <span>{s.Language === "go" ? "Tab" : "Espaços"}</span>
      <span>UTF-8{s.BOM ? " BOM" : ""}</span>
      <span>{s.LineEnding}</span>
      {s.Language === "go" && <span>Go {s.GoVersion}</span>}
      {s.LSP && <span className={`lsp ${s.LSP}`}>● {LSP_LABEL[s.LSP]}</span>}
    </footer>
  );
}
