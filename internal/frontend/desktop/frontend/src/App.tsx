import { useSyncExternalStore } from "react";
import { uiStore } from "./bridge";
import { TopBar } from "./components/TopBar";
import { Rail } from "./components/Rail";
import { Sidebar } from "./components/Sidebar";
import { Tabs } from "./components/Tabs";
import { EditorView } from "./components/EditorView";
import { BottomPanel } from "./components/BottomPanel";
import { StatusBar } from "./components/StatusBar";

export function App() {
  const ui = useSyncExternalStore(uiStore.subscribe, uiStore.get);

  return (
    <div className="shell">
      <TopBar ui={ui} />
      <div className={ui?.Explorer.Open ? "layout" : "layout no-side"}>
        <Rail ui={ui} />
        {ui?.Explorer.Open && <Sidebar ui={ui} />}
        <main className="center">
          <Tabs ui={ui} />
          <EditorView />
          {ui && ui.Bottom !== "" && <BottomPanel ui={ui} />}
        </main>
      </div>
      <StatusBar ui={ui} />
    </div>
  );
}
