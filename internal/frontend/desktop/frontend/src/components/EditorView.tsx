import { useEffect, useRef } from "react";
import { attachEditor, clickText, wheelText } from "../bridge";

export function EditorView() {
  const canvas = useRef<HTMLCanvasElement>(null);
  const textarea = useRef<HTMLTextAreaElement>(null);
  const attached = useRef(false);

  useEffect(() => {
    if (attached.current) {
      return;
    }
    attached.current = true;
    attachEditor(canvas.current!, textarea.current!);
  }, []);

  return (
    <section className="editor">
      <canvas
        ref={canvas}
        onMouseDown={(e) => {
          if (e.button === 0) {
            clickText(e.nativeEvent.offsetX, e.nativeEvent.offsetY);
          }
        }}
        onWheel={(e) => wheelText(e.deltaY, e.deltaMode)}
      />
      <textarea
        ref={textarea}
        className="keys"
        autoComplete="off"
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck={false}
      />
    </section>
  );
}
