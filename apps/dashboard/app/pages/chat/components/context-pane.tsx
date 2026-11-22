import { useState, type ReactNode } from "react";
import { PanelContent, type WorkspaceView } from "@cueia/components";
import { Button } from "@cueia/ui";
import { Link } from "react-router";
import Add from "@react-spectrum/s2/icons/Add";
import ChatIcon from "@react-spectrum/s2/icons/Chat";
import Lock from "@react-spectrum/s2/icons/Lock";
import { ResizableSidebar } from "../../../components/resizable-sidebar";
import { FindPanel } from "../../../components/find-panel";
import { SearchBox } from "../../../components/search-box";
import type { Schema } from "../../../api";
type LeftView = Exclude<WorkspaceView, "tasks">;
const panelNames = {
  chat: "Chats",
  find: "Find",
  files: "Files",
  ideas: "Ideas",
};
export function ContextPane({
  left,
  tasksOpen,
  onClose,
  active,
  sideChats,
  openChat,
  onCreateSideChat,
  data,
  sampleMode,
  files,
}: {
  left: LeftView | null;
  tasksOpen: boolean;
  onClose: () => void;
  active: string;
  sideChats: { id: string; title: string }[];
  openChat: (id: string) => void;
  onCreateSideChat: () => void;
  data: Schema["Workspace"];
  sampleMode: boolean;
  files: ReactNode;
}) {
  const [search, setSearch] = useState("");
  const matches = (value: string) =>
    value.toLowerCase().includes(search.trim().toLowerCase());

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
            onPress={() => {
              setSearch("");
              onCreateSideChat();
            }}
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
        {files}
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
    <ResizableSidebar
      side="context"
      panel={left}
      otherOpen={tasksOpen}
      onEscape={onClose}
    >
      {(Object.keys(panels) as LeftView[]).map((view) => (
        <aside
          key={view}
          id={view + "-panel"}
          hidden={left !== view}
          className="[[hidden]]:hidden h-full overflow-y-auto min-w-0 scrollbar-thin [&_label]:grid [&_label]:gap-2 [&_label]:text-[14px] [&_label_>_input]:w-full [&_label_>_input]:bg-raised [&_label_>_input]:p-2 [&_label_>_input]:[border:1px_solid_#53514b] [&_label_>_input]:rounded-md flex flex-col [border-right:1px_solid_var(--color-line)] bg-panel compact:[border-right:0] compact:[border-bottom:1px_solid_var(--color-line)]"
          aria-label={panelNames[view]}
        >
          <PanelContent visible={left === view}>{panels[view]}</PanelContent>
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
  );
}
