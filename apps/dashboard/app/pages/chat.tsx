import { useEffect, useRef, useState } from "react";
import { Link } from "react-router";
import Add from "@react-spectrum/s2/icons/Add";
import ChatIcon from "@react-spectrum/s2/icons/Chat";
import Lock from "@react-spectrum/s2/icons/Lock";
import Close from "@react-spectrum/s2/icons/Close";
import Attach from "@react-spectrum/s2/icons/Attach";
import {
  CalendarEntry,
  Correspondence,
  DraftComposer,
  PanelContent,
  WorkspaceNavigation,
  type WorkspaceView,
} from "@cueia/components";
import {
  Button,
  Heading,
  Modal,
  ModalContainer,
  ModalContent,
} from "@cueia/ui";
import { Files } from "./files";
import { compactWorkspace, ResizableSidebar } from "../resizable-sidebar";
import { TaskDetails } from "../task-details";
import { SampleConversation } from "../sample-content";
import { FindPanel, SampleFiles, SearchBox } from "../workspace-panels";
import type { Schema } from "../api";

type Props = {
  data: Schema["Workspace"];
  onPreference: (format: string) => Promise<void>;
  onSignOut: () => Promise<void>;
} & (
  { sampleMode: true; token?: never } | { sampleMode?: false; token: string }
);

type LeftView = Exclude<WorkspaceView, "tasks">;
const panelNames = {
  chat: "Chats",
  find: "Find",
  files: "Files",
  ideas: "Ideas",
};
const newDraft = { text: "", file: null as File | null };

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
  const [search, setSearch] = useState("");
  const [active, setActive] = useState("main");
  const [sideChats, setSideChats] = useState<{ id: string; title: string }[]>(
    [],
  );
  const [drafts, setDrafts] = useState<Record<string, typeof newDraft>>({});
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const feed = useRef<HTMLDivElement>(null);
  const fileInput = useRef<HTMLInputElement>(null);
  const hour12 = data.preferences.timeFormat !== "24h";
  const draft = drafts[active] ?? newDraft;
  const sideChat = sideChats.find((item) => item.id === active);
  const thread = data.threads.find((item) => item.id === active);
  const event = data.events.find((item) => item.id === active);
  const title =
    sideChat?.title ??
    (thread ? "Email thread" : event ? "Calendar event" : "Main chat");
  const matches = (value: string) =>
    value.toLowerCase().includes(search.trim().toLowerCase());

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
    setActive(id);
    if (matchMedia(compactWorkspace).matches) setLeft(null);
    feed.current?.focus();
  }

  function createSideChat() {
    const id = crypto.randomUUID();
    setSearch("");
    setSideChats((items) => [
      ...items,
      { id, title: "Side chat " + (items.length + 1) },
    ]);
    openChat(id);
    document
      .querySelector<HTMLTextAreaElement>(".draft-composer textarea")
      ?.focus();
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

  const visibleSideChats = sideChats.filter((item) => matches(item.title));
  const panels = {
    chat: (
      <>
        <SearchBox label="Search chats" value={search} onChange={setSearch} />
        {matches("Main chat") && (
          <button
            className="thread-link"
            aria-current={active === "main" ? "page" : undefined}
            onClick={() => openChat("main")}
          >
            <ChatIcon />
            Main chat
          </button>
        )}
        <div className="side-chats-heading">
          <h2>Side chats</h2>
          <Button
            className="icon-button"
            aria-label="New side chat"
            onPress={createSideChat}
          >
            <Add />
          </Button>
        </div>
        {visibleSideChats.map((item) => (
          <button
            key={item.id}
            className="thread-link"
            aria-current={active === item.id ? "page" : undefined}
            onClick={() => openChat(item.id)}
          >
            <ChatIcon />
            {item.title}
          </button>
        ))}
        {!matches("Main chat") && !visibleSideChats.length && (
          <p className="empty-state">No matching chats.</p>
        )}
      </>
    ),
    files: (
      <>
        <h2 className="panel-title">Files</h2>
        {sampleMode ? <SampleFiles /> : <Files token={token} />}
      </>
    ),
    find: <FindPanel data={data} onSelect={openChat} />,
    ideas: (
      <>
        <h2 className="panel-title">Ideas</h2>
        <p className="empty-state">No ideas yet.</p>
      </>
    ),
  };

  return (
    <main className="workspace">
      <WorkspaceNavigation
        expanded={{ ...(left ? { [left]: true } : {}), tasks: tasksOpen }}
        onSelect={select}
        menuOpen={menu}
        onMenu={() => {
          setMenu(true);
          setError("");
        }}
      />
      <div className="workspace-body">
        <div className="chat-area">
          <ResizableSidebar
            side="context"
            panel={left}
            otherOpen={tasksOpen}
            onEscape={() => {
              if (left) document.getElementById("nav-" + left)?.focus();
              setLeft(null);
            }}
          >
            {(Object.keys(panels) as LeftView[]).map((view) => (
              <aside
                key={view}
                id={view + "-panel"}
                hidden={left !== view}
                className="workspace-view context-panel"
                aria-label={panelNames[view]}
              >
                <PanelContent visible={left === view}>
                  {panels[view]}
                </PanelContent>
                <div className="workspace-identity">
                  <Lock />
                  {sampleMode ? (
                    <Link to="/workspace" aria-label="Open sample workspace">
                      Sample workspace
                    </Link>
                  ) : (
                    "Private workspace"
                  )}
                </div>
              </aside>
            ))}
          </ResizableSidebar>
          <section className="conversation" aria-label={title}>
            <header className="conversation-heading">
              <ChatIcon />
              <h1>{title}</h1>
            </header>
            <div className="feed" tabIndex={-1} ref={feed}>
              <div className="reading-column">
                {active === "main" &&
                  (sampleMode ? (
                    <SampleConversation hour12={hour12} />
                  ) : (
                    <>
                      {data.threads.map((item) => (
                        <Correspondence key={item.id} {...item} />
                      ))}
                      {data.events.map((item) => (
                        <CalendarEntry
                          key={item.id}
                          {...item}
                          hour12={hour12}
                        />
                      ))}
                    </>
                  ))}
                {thread && <Correspondence {...thread} />}
                {event && <CalendarEntry {...event} hour12={hour12} />}
                {sideChat && (
                  <p className="empty-state side-chat-empty">
                    Start a draft in {sideChat.title.toLowerCase()}.
                  </p>
                )}
              </div>
            </div>
            <div className="composer-region">
              {draft.file && (
                <div className="draft-file">
                  <Attach />
                  {draft.file.name}
                  <Button
                    className="icon-button"
                    aria-label="Remove attachment"
                    onPress={() => updateDraft({ file: null })}
                  >
                    <Close />
                  </Button>
                </div>
              )}
              <input
                ref={fileInput}
                type="file"
                hidden
                onChange={(e) => {
                  if (e.target.files?.[0])
                    updateDraft({ file: e.target.files[0] });
                  e.target.value = "";
                }}
              />
              <DraftComposer
                draft={draft.text}
                onChange={(text) => updateDraft({ text })}
                onAttach={() => fileInput.current?.click()}
                note={
                  sampleMode
                    ? "Sample conversation · Drafts stay in this tab"
                    : "Draft only · Sending is unavailable until the agent connects."
                }
              />
            </div>
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
            className="workspace-view tasks-panel"
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
      <ModalContainer
        onDismiss={() => {
          setMenu(false);
          document.getElementById("workspace-menu")?.focus();
        }}
      >
        {menu && (
          <Modal isDismissible size="S">
            <Heading>Preferences</Heading>
            <ModalContent>
              <div className="workspace-preferences">
                <label>
                  Time format
                  <select
                    disabled={saving}
                    value={data.preferences.timeFormat ?? "12h"}
                    onChange={(e) => {
                      const value = e.target.value;
                      void save(
                        () => onPreference(value),
                        "Could not save preference.",
                      );
                    }}
                  >
                    <option value="12h">12 hour</option>
                    <option value="24h">24 hour</option>
                  </select>
                </label>
                <Button
                  isDisabled={saving}
                  onPress={() =>
                    void save(onSignOut, "Could not sign out. Please retry.")
                  }
                >
                  Sign out
                </Button>
                {sampleMode && (
                  <Link to="/workspace">Open private workspace</Link>
                )}
                {error && <p role="alert">{error}</p>}
              </div>
            </ModalContent>
          </Modal>
        )}
      </ModalContainer>
    </main>
  );
}
