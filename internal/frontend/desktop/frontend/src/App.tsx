import { useSyncExternalStore } from "react";
import { uiStore } from "./bridge";
import { TopBar } from "./components/TopBar";
import { Rail } from "./components/Rail";
import { Sidebar } from "./components/Sidebar";
import { Tabs } from "./components/Tabs";
import { EditorView } from "./components/EditorView";
import { BottomPanel, useBottomVisible } from "./components/BottomPanel";
import { StatusBar } from "./components/StatusBar";

export function App() {
  const ui = useSyncExternalStore(uiStore.subscribe, uiStore.get);
  const showBottom = useBottomVisible(ui);

  return (
    <div className="shell">
      <TopBar ui={ui} />
      <div className={ui?.Explorer.Open ? "layout" : "layout no-side"}>
        <Rail ui={ui} />
        {ui?.Explorer.Open && <Sidebar ui={ui} />}
        <main className="center">
          <Tabs ui={ui} />
          <EditorView />
          {ui && showBottom && <BottomPanel ui={ui} />}
        </main>
      </div>
      <StatusBar ui={ui} />
    </div>
  );
}
