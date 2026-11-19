import { createRoot } from "react-dom/client";
import { App } from "./App";
import { loadTheme } from "./bridge";
import "@fontsource/vt323/latin-400.css";
import "@fontsource/space-mono/latin-400.css";
import "@fontsource/space-mono/latin-700.css";
import "./styles.css";

loadTheme().finally(() => {
  createRoot(document.getElementById("root")!).render(<App />);
});
