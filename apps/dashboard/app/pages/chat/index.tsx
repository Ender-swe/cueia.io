import { useEffect, useRef, useState } from "react";
import { flushSync } from "react-dom";
import ChatIcon from "@react-spectrum/s2/icons/Chat";
import { WorkspaceNavigation, type WorkspaceView } from "@cueia/components";
import { ChatComposer, type Draft } from "./components/chat-composer";
import { ContextPane } from "./components/context-pane";
import { ConversationFeed } from "./components/conversation-feed";
import { PreferencesDialog } from "./components/preferences-dialog";
import { SampleFiles } from "../../components/sample-files";
import { Files } from "../files";
import { compactWorkspace, ResizableSidebar } from "../../components/resizable-sidebar";
import { TaskDetails } from "../../components/task-details";
import type { Schema } from "../../api";

type Props = {
  data: Schema["Workspace"];
  onPreference: (format: string) => Promise<void>;
  onSignOut: () => Promise<void>;
} & (
  { sampleMode: true; token?: never } | { sampleMode?: false; token: string }
);

type LeftView = Exclude<WorkspaceView, "tasks">;
const newDraft: Draft = { text: "", file: null };

export function Chat({
  data,
  token,
  sampleMode,
  onPreference,
  onSignOut,
}: Props) {
  const [left, setLeft] = useState<LeftView | null>(() =>
    matchMedia(compactWorkspace).matches ? null : "chat",
  );
  const [tasksOpen, setTasksOpen] = useState(
    () => !matchMedia(compactWorkspace).matches,
  );
  const [menu, setMenu] = useState(false);
  const [active, setActive] = useState("main");
  const [sideChats, setSideChats] = useState<{ id: string; title: string }[]>(
    [],
  );
  const [drafts, setDrafts] = useState<Record<string, typeof newDraft>>({});
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const composer = useRef<HTMLDivElement>(null);
  const feed = useRef<HTMLDivElement>(null);
  const hour12 = data.preferences.timeFormat !== "24h";
  const draft = drafts[active] ?? newDraft;
  const sideChat = sideChats.find((item) => item.id === active);
  const thread = data.threads.find((item) => item.id === active);
  const event = data.events.find((item) => item.id === active);
  const title =
    sideChat?.title ??
    (thread ? "Email thread" : event ? "Calendar event" : "Main chat");
  useEffect(() => {
    const media = matchMedia(compactWorkspace);
    const resized = () => {
      setLeft(media.matches ? null : "chat");
      if (media.matches) setTasksOpen(false);
    };
    media.addEventListener("change", resized);
    return () => media.removeEventListener("change", resized);
  }, []);

  useEffect(() => {
    feed.current?.scrollTo(0, 0);
  }, [active]);

  function select(view: WorkspaceView) {
    const compact = matchMedia(compactWorkspace).matches;
    if (view === "tasks") {
      setTasksOpen((value) => !value);
      if (compact) setLeft(null);
    } else {
      setLeft((current) => (current === view ? null : view));
      if (compact) setTasksOpen(false);
    }
  }

  function closeTasks() {
    setTasksOpen(false);
    document.getElementById("nav-tasks")?.focus();
  }

  function openChat(id: string) {
    flushSync(() => {
      setActive(id);
      if (matchMedia(compactWorkspace).matches) setLeft(null);
    });
    feed.current?.focus();
  }

  function createSideChat() {
    const id = crypto.randomUUID();
    setSideChats((items) => [
      ...items,
      { id, title: "Side chat " + (items.length + 1) },
    ]);
    openChat(id);
    composer.current?.querySelector("textarea")?.focus();
  }

  function updateDraft(change: Partial<typeof newDraft>) {
    setDrafts((current) => ({
      ...current,
      [active]: { ...(current[active] ?? newDraft), ...change },
    }));
  }

  async function save(action: () => Promise<void>, message: string) {
    setSaving(true);
    setError("");
    try {
      await action();
    } catch {
      setError(message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="font-sans text-ink flex isolate h-dvh min-h-0 bg-canvas text-interface [font-synthesis:none] [&_:where(svg)]:shrink-0 [&_:where(svg)]:size-4.5">
      <WorkspaceNavigation
        expanded={{ ...(left ? { [left]: true } : {}), tasks: tasksOpen }}
        onSelect={select}
        menuOpen={menu}
        onMenu={() => {
          setMenu(true);
          setError("");
        }}
      />
      <div data-workspace-body className="flex min-h-0 min-w-0 flex-1">
        <div
          className={`flex min-h-0 min-w-0 flex-1 compact:flex-col ${tasksOpen ? "phone:hidden" : ""}`}
        >
          <ContextPane
            left={left}
            tasksOpen={tasksOpen}
            onClose={() => {
              if (left) document.getElementById("nav-" + left)?.focus();
              setLeft(null);
            }}
            active={active}
            sideChats={sideChats}
            openChat={openChat}
            onCreateSideChat={createSideChat}
            data={data}
            sampleMode={!!sampleMode}
            files={sampleMode ? <SampleFiles /> : <Files token={token} />}
          />
          <section
            className={`flex min-h-0 min-w-0 flex-1 flex-col @container/conversation ${left ? "phone:hidden" : ""}`}
            aria-label={title}
          >
            <header className="flex items-center flex-none gap-2.5 h-15.5 px-7 [border-bottom:1px_solid_var(--color-line)] [&_h1]:text-[17px] [&_h1]:font-semibold [&_h1]:tracking-tight [&_svg]:text-muted [&_svg]:w-4.25 [&_svg]:h-4.25 compact:h-12.5 compact:px-4">
              <ChatIcon />
              <h1>{title}</h1>
            </header>
            <ConversationFeed
              feed={feed}
              active={active}
              sampleMode={!!sampleMode}
              hour12={hour12}
              data={data}
              thread={thread}
              event={event}
              sideChat={sideChat}
            />
            <ChatComposer
              composer={composer}
              draft={draft}
              updateDraft={updateDraft}
              sampleMode={!!sampleMode}
            />
          </section>
        </div>
        <ResizableSidebar
          side="tasks"
          panel={tasksOpen ? "tasks" : null}
          otherOpen={!!left}
          onEscape={closeTasks}
        >
          <aside
            id="tasks-panel"
            hidden={!tasksOpen}
            className="[[hidden]]:hidden h-full overflow-y-auto min-w-0 scrollbar-thin [&_label]:grid [&_label]:gap-2 [&_label]:text-[14px] [&_label_>_input]:w-full [&_label_>_input]:bg-raised [&_label_>_input]:p-2 [&_label_>_input]:[border:1px_solid_#53514b] [&_label_>_input]:rounded-md [border-left:1px_solid_var(--color-line)] bg-panel @container/tasks compact:wrap-anywhere"
            aria-label="Tasks"
          >
            <TaskDetails
              sampleMode={!!sampleMode}
              hour12={hour12}
              onClose={closeTasks}
            />
          </aside>
        </ResizableSidebar>
      </div>
      <PreferencesDialog
        menu={menu}
        timeFormat={data.preferences.timeFormat ?? "12h"}
        saving={saving}
        save={save}
        onPreference={onPreference}
        onSignOut={onSignOut}
        sampleMode={!!sampleMode}
        error={error}
        onDismiss={() => {
          setMenu(false);
          document.getElementById("workspace-menu")?.focus();
        }}
      />
    </main>
  );
}
