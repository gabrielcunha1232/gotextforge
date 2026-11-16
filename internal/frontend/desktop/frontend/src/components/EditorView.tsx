import { useEffect, useRef } from "react";
import { attachEditor } from "../bridge";

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
      <canvas ref={canvas} />
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
