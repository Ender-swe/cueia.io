import Chat from "@react-spectrum/s2/icons/Chat";
import Search from "@react-spectrum/s2/icons/Search";
import Folder from "@react-spectrum/s2/icons/Folder";
import Lightbulb from "@react-spectrum/s2/icons/Lightbulb";
import Tasks from "@react-spectrum/s2/icons/CheckBox";
import Menu from "@react-spectrum/s2/icons/MenuHamburger";
import { Button } from "@cueia/ui/button";
import { CueiaMark } from "../cueia-mark";

export type WorkspaceView = "chat" | "find" | "files" | "ideas" | "tasks";

const destinations = [
  { view: "chat", label: "Chat", Icon: Chat },
  { view: "find", label: "Find", Icon: Search },
  { view: "files", label: "Files", Icon: Folder },
  { view: "ideas", label: "Ideas", Icon: Lightbulb },
  { view: "tasks", label: "Tasks", Icon: Tasks },
] as const;

interface NavigationProps {
  expanded: Partial<Record<WorkspaceView, boolean>>;
  onSelect: (view: WorkspaceView) => void;
  menuOpen: boolean;
  onMenu: () => void;
}

const navigationButton =
  "flex flex-col items-center justify-center gap-1 shrink-0 w-full h-14 py-[5px] px-0 border-0 rounded-lg bg-transparent text-muted [font-family:inherit] text-caption leading-[1.2] font-medium [&_svg]:w-[19px] [&_svg]:h-[19px] aria-expanded:bg-nav-selected aria-expanded:text-ink [&#nav-chat[aria-expanded=true]]:text-apricot hover:bg-raised hover:text-ink compact:h-[47px] short-screen:h-10 short-screen:gap-0.5";

export function WorkspaceNavigation({
  expanded,
  onSelect,
  menuOpen,
  onMenu,
}: NavigationProps) {
  return (
    <nav
      className="flex items-center flex-col [flex:0_0_72px] p-[22px_8px_12px] gap-2 [border-right:1px_solid_var(--color-line)] bg-rail overflow-auto scrollbar-none compact:basis-13.5 compact:p-[16px_4px_6px] short-screen:py-1"
      aria-label="Workspace"
    >
      <div role="img" aria-label="Cueia">
        <CueiaMark />
      </div>
      <div className="grid gap-1.5 my-auto w-full compact:gap-0.5 short-screen:gap-0">
        {destinations.map(({ view, label, Icon }) => (
          <Button
            key={view}
            id={"nav-" + view}
            className={navigationButton}
            aria-label={label}
            aria-controls={view + "-panel"}
            aria-expanded={!!expanded[view]}
            onPress={() => onSelect(view)}
          >
            <Icon />
            <span>{label}</span>
          </Button>
        ))}
      </div>
      <Button
        id="workspace-menu"
        className={navigationButton}
        aria-label="Menu"
        aria-haspopup="dialog"
        aria-expanded={menuOpen}
        onPress={onMenu}
      >
        <Menu />
        <span>Menu</span>
      </Button>
    </nav>
  );
}
