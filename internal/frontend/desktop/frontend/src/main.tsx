import { createRoot } from "react-dom/client";
import { App } from "./App";
import { loadTheme } from "./bridge";
import "./styles.css";

loadTheme().finally(() => {
  createRoot(document.getElementById("root")!).render(<App />);
});
