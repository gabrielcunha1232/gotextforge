import { useEffect, useRef } from "react";
import { invoke, sendKey } from "../bridge";
import type { UIFinder } from "../types";

function Highlighted({ text, positions }: { text: string; positions: number[] | null }) {
  const chars = Array.from(text);
  const hit = new Set(positions ?? []);
  const slash = Math.max(text.lastIndexOf("/"), text.lastIndexOf("\\"));
  const nameStart = slash < 0 ? 0 : Array.from(text.slice(0, slash + 1)).length;

  const part = (from: number, to: number) =>
    chars.slice(from, to).map((c, j) => {
      const i = from + j;
      return hit.has(i) ? (
        <mark key={i}>{c}</mark>
      ) : (
        <span key={i}>{c}</span>
      );
    });

  return (
    <>
      <span className="fname">{part(nameStart, chars.length)}</span>
      {nameStart > 0 && <span className="fdir">{part(0, nameStart - 1)}</span>}
    </>
  );
}

export function Finder({ finder }: { finder: UIFinder }) {
  const items = finder.Items ?? [];
  const list = useRef<HTMLDivElement>(null);

  useEffect(() => {
    list.current?.querySelector(".fitem.selected")?.scrollIntoView({ block: "nearest" });
  }, [finder.Selected, items.length]);

  return (
    <div className="finder-backdrop" onClick={() => sendKey({ key: "Escape" })}>
      <div className="finder" onClick={(e) => e.stopPropagation()}>
        <div className="finput">
          <span className="fprompt">›</span>
          <span className="fquery">{finder.Query}</span>
          <span className="fcaret" />
          {finder.Query === "" && <span className="fplaceholder">digite parte do nome… (edtgo acha editor/editor.go)</span>}
          <span className="fcount">
            {items.length}/{finder.Total}
            {finder.Truncated && "+"}
          </span>
        </div>
        <div className="ftitle">{finder.Title}</div>
        <div className="flist" ref={list}>
          {items.map((it, i) => (
            <div
              key={it.Text}
              className={i === finder.Selected ? "fitem selected" : "fitem"}
              onClick={() => invoke("FinderClick", i)}
            >
              <Highlighted text={it.Text} positions={it.Positions} />
            </div>
          ))}
          {finder.Create && (
            <div
              className={finder.Selected === items.length ? "fitem create selected" : "fitem create"}
              onClick={() => invoke("FinderClick", items.length)}
            >
              ＋ criar <b>{finder.Query.trim()}</b>
            </div>
          )}
          {items.length === 0 && !finder.Create && <div className="fempty">Nada encontrado.</div>}
        </div>
        <div className="ffoot">
          <kbd>↑↓</kbd> navegar <kbd>Enter</kbd> abrir <kbd>Esc</kbd> fechar <kbd>Ctrl u</kbd> limpar
        </div>
      </div>
    </div>
  );
}
