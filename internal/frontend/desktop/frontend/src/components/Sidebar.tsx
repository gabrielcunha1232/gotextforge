import { useEffect, useRef } from "react";
import { invoke } from "../bridge";
import type { UIState } from "../types";

type Icon = { text: string; cls: string };

function fileIcon(name: string): Icon {
  const lower = name.toLowerCase();
  if (lower.endsWith("_test.go")) return { text: "go", cls: "ic-test" };
  if (lower.endsWith(".go")) return { text: "go", cls: "ic-go" };
  if (lower === "go.mod" || lower === "go.sum" || lower === "go.work") return { text: "◈", cls: "ic-mod" };
  if (lower.endsWith(".md")) return { text: "M↓", cls: "ic-md" };
  if (lower === "dockerfile" || lower.startsWith("compose.") || lower.startsWith("docker-compose."))
    return { text: "◳", cls: "ic-docker" };
  if (/\.(json|ya?ml|toml)$/.test(lower)) return { text: "{}", cls: "ic-data" };
  if (/\.(tsx?|jsx?)$/.test(lower)) return { text: lower.includes(".ts") ? "TS" : "JS", cls: "ic-script" };
  if (/\.(exe|dll|so|dylib)$/.test(lower)) return { text: "▣", cls: "ic-bin" };
  if (lower.startsWith(".")) return { text: "·", cls: "ic-hidden" };
  return { text: "≡", cls: "ic-file" };
}

export function Sidebar({ ui }: { ui: UIState }) {
  const x = ui.Explorer;
  const rows = x.Rows ?? [];
  const focused = ui.Focus === "explorer";
  const list = useRef<HTMLDivElement>(null);

  useEffect(() => {
    list.current?.querySelector(".row.selected")?.scrollIntoView({ block: "nearest" });
  }, [x.Selected, rows.length]);

  return (
    <aside className="side">
      <div className="cap">
        Explorador
        <button className="cap-btn" title="Abrir pasta… (Ctrl Shift O)" onClick={() => invoke("OpenFolder")}>
          Abrir pasta…
        </button>
      </div>
      {rows.length === 0 ? (
        <div className="hint">Pasta vazia</div>
      ) : (
        <>
          <div className="proj" title={ui.Root}>
            ⌄　{x.Name}
          </div>
          <div className="tree" ref={list}>
            {rows.map((r, i) => {
              const cls = ["row"];
              if (r.Dir) cls.push("dir");
              if (r.Current) cls.push("current");
              if (i === x.Selected) cls.push("selected", focused ? "sel-focus" : "sel");
              const icon = r.Dir ? null : fileIcon(r.Name);
              return (
                <div
                  key={r.Path}
                  className={cls.join(" ")}
                  style={{ paddingLeft: 6 + r.Depth * 14 }}
                  title={r.Path}
                  onClick={() => invoke("ExplorerClick", r.Path)}
                >
                  <span className="twisty">{r.Dir ? (r.Expanded ? "▾" : "▸") : ""}</span>
                  {icon && <span className={`icon ${icon.cls}`}>{icon.text}</span>}
                  <span className="name">{r.Name}</span>
                  {r.Dir && r.Expanded && r.Empty && <span className="muted"> (vazia)</span>}
                </div>
              );
            })}
          </div>
        </>
      )}
      {focused && (
        <div className="side-keys">
          <kbd>a</kbd> criar <kbd>r</kbd> renomear <kbd>d</kbd> apagar <kbd>H</kbd> ocultos
        </div>
      )}
    </aside>
  );
}
