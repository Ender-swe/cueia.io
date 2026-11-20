import { useEffect, useRef, useState } from "react";
import { flushSync } from "react-dom";
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
  const composer = useRef<HTMLDivElement>(null);
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
    flushSync(() => {
      setActive(id);
      if (matchMedia(compactWorkspace).matches) setLeft(null);
    });
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

  const visibleSideChats = sideChats.filter((item) => matches(item.title));
  const panels = {
    chat: (
      <>
        <SearchBox label="Search chats" value={search} onChange={setSearch} />
        {matches("Main chat") && (
          <button
            className="flex items-center gap-2.5 w-full min-h-9.5 py-1.75 px-2.75 bg-transparent border-0 rounded-md text-left wrap-anywhere text-[14px] font-label first-of-type:mt-2 [&_svg]:w-4 [&_svg]:h-4 [&_svg]:text-muted aria-[current=page]:bg-thread-selected [&[aria-current=page]_svg]:text-apricot hover:bg-raised"
            aria-current={active === "main" ? "page" : undefined}
            onClick={() => openChat("main")}
          >
            <ChatIcon />
            Main chat
          </button>
        )}
        <div className="flex items-center justify-between pl-2.5 mt-2.25 [&_h2]:text-[13px] [&_h2]:font-label [&_h2]:text-muted">
          <h2>Side chats</h2>
          <Button
            className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8 p-0 rounded-[7px] text-muted bg-transparent hover:text-ink hover:bg-[#373633]"
            aria-label="New side chat"
            onPress={createSideChat}
          >
            <Add />
          </Button>
        </div>
        {visibleSideChats.map((item) => (
          <button
            key={item.id}
            className="flex items-center gap-2.5 w-full min-h-9.5 py-1.75 px-2.75 bg-transparent border-0 rounded-md text-left wrap-anywhere text-[14px] font-label first-of-type:mt-2 [&_svg]:w-4 [&_svg]:h-4 [&_svg]:text-muted aria-[current=page]:bg-thread-selected [&[aria-current=page]_svg]:text-apricot hover:bg-raised"
            aria-current={active === item.id ? "page" : undefined}
            onClick={() => openChat(item.id)}
          >
            <ChatIcon />
            {item.title}
          </button>
        ))}
        {!matches("Main chat") && !visibleSideChats.length && (
          <p className="text-[13px] text-muted">No matching chats.</p>
        )}
      </>
    ),
    files: (
      <>
        <h2 className="py-1 px-2 text-[16px] font-semibold">Files</h2>
        {sampleMode ? <SampleFiles /> : <Files token={token} />}
      </>
    ),
    find: <FindPanel data={data} onSelect={openChat} />,
    ideas: (
      <>
        <h2 className="py-1 px-2 text-[16px] font-semibold">Ideas</h2>
        <p className="text-[13px] text-muted">No ideas yet.</p>
      </>
    ),
  };

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
                className="[[hidden]]:hidden h-full overflow-y-auto min-w-0 scrollbar-thin [&_label]:grid [&_label]:gap-2 [&_label]:text-[14px] [&_label_>_input]:w-full [&_label_>_input]:bg-raised [&_label_>_input]:p-2 [&_label_>_input]:[border:1px_solid_#53514b] [&_label_>_input]:rounded-md flex flex-col [border-right:1px_solid_var(--color-line)] bg-panel compact:[border-right:0] compact:[border-bottom:1px_solid_var(--color-line)]"
                aria-label={panelNames[view]}
              >
                <PanelContent visible={left === view}>
                  {panels[view]}
                </PanelContent>
                <div className="mt-auto flex items-center gap-2 py-4.5 px-5.75 text-muted text-[12px] [&_svg]:w-3.25 [&_svg]:h-3.25 [&_a]:text-inherit [&_a]:no-underline [&_a:hover]:underline compact:hidden">
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
          <section
            className={`flex min-h-0 min-w-0 flex-1 flex-col @container/conversation ${left ? "phone:hidden" : ""}`}
            aria-label={title}
          >
            <header className="flex items-center flex-none gap-2.5 h-15.5 px-7 [border-bottom:1px_solid_var(--color-line)] [&_h1]:text-[17px] [&_h1]:font-semibold [&_h1]:tracking-tight [&_svg]:text-muted [&_svg]:w-4.25 [&_svg]:h-4.25 compact:h-12.5 compact:px-4">
              <ChatIcon />
              <h1>{title}</h1>
            </header>
            <div
              className="min-h-0 flex-1 overflow-y-auto px-[clamp(18px,4.5cqw,40px)] scrollbar-thin wrap-anywhere focus:[outline:none]"
              tabIndex={-1}
              ref={feed}
            >
              <div className="w-[min(100%,720px)] mx-auto py-6.5 flex flex-col gap-[clamp(20px,3vh,31px)] compact:py-4.5 compact:gap-4.5">
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
                  <p className="text-[13px] text-muted text-center">
                    Start a draft in {sideChat.title.toLowerCase()}.
                  </p>
                )}
              </div>
            </div>
            <div className="flex-none p-[8px_clamp(14px,3.5cqw,32px)_max(12px,env(safe-area-inset-bottom))]">
              {draft.file && (
                <div className="flex items-center gap-1.5 text-apricot text-[13px] my-[0_6px] wrap-anywhere w-[min(100%,744px)] mx-auto">
                  <Attach />
                  {draft.file.name}
                  <Button
                    className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8 p-0 rounded-[7px] text-muted bg-transparent hover:text-ink hover:bg-[#373633]"
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
                composerRef={composer}
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
              <div className="font-sans text-ink leading-normal grid justify-items-start gap-4 [&_a]:text-apricot">
                <label className="grid gap-2 text-[14px]">
                  Time format
                  <select
                    className="w-full bg-raised p-2 [border:1px_solid_#53514b] rounded-md"
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
