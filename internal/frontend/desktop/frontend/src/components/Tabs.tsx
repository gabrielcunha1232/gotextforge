import { useEffect, useRef } from "react";
import { invoke } from "../bridge";
import type { UIState } from "../types";

export function Tabs({ ui }: { ui: UIState | null }) {
  const tabs = ui?.Tabs ?? [];
  const bar = useRef<HTMLDivElement>(null);
  const active = tabs.findIndex((t) => t.Active);

  useEffect(() => {
    bar.current?.querySelector(".tab.active")?.scrollIntoView({ block: "nearest", inline: "nearest" });
  }, [active, tabs.length]);

  return (
    <>
      <div
        className="tabs"
        ref={bar}
        onWheel={(e) => {
          bar.current!.scrollLeft += e.deltaY;
        }}
      >
        {tabs.map((t, i) => (
          <div
            key={`${t.Path}-${i}`}
            className={t.Active ? "tab active" : "tab"}
            title={t.Path || t.Name}
            onClick={() => invoke("SwitchTab", i)}
            onAuxClick={(e) => {
              if (e.button === 1) {
                invoke("CloseTab", i);
              }
            }}
          >
            <span className="name">{t.Name}</span>
            <span
              className={t.Modified ? "close modified" : "close"}
              title="Fechar (botão do meio também fecha)"
              onClick={(e) => {
                e.stopPropagation();
                invoke("CloseTab", i);
              }}
            >
              <span className="dot">●</span>
              <span className="x">×</span>
            </span>
          </div>
        ))}
      </div>
      <div className="crumb">{crumbs(ui)}</div>
    </>
  );
}

function crumbs(ui: UIState | null) {
  const path = ui?.Status.Path;
  if (!ui || !path) {
    return null;
  }
  let rel = path;
  const root = ui.Root.replace(/[\\/]+$/, "");
  if (root && rel.toLowerCase().startsWith(root.toLowerCase())) {
    rel = rel.slice(root.length);
  }
  const parts = rel.split(/[\\/]/).filter(Boolean);
  return parts.map((p, i) => (
    <span key={i} className={i === parts.length - 1 ? "last" : ""}>
      {i > 0 && <span className="sep">›</span>}
      {p}
    </span>
  ));
}
