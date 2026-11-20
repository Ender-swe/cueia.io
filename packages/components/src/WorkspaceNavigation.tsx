import Chat from "@react-spectrum/s2/icons/Chat"
import Search from "@react-spectrum/s2/icons/Search"
import Folder from "@react-spectrum/s2/icons/Folder"
import Lightbulb from "@react-spectrum/s2/icons/Lightbulb"
import Tasks from "@react-spectrum/s2/icons/CheckBox"
import Menu from "@react-spectrum/s2/icons/MenuHamburger"
import { Button } from "@cueia/ui/button"
import { motion, useReducedMotion } from "motion/react"
import type { ReactNode } from "react"

export type WorkspaceView = "chat" | "find" | "files" | "ideas" | "tasks"

const destinations = [
  { view: "chat", label: "Chat", Icon: Chat },
  { view: "find", label: "Find", Icon: Search },
  { view: "files", label: "Files", Icon: Folder },
  { view: "ideas", label: "Ideas", Icon: Lightbulb },
  { view: "tasks", label: "Tasks", Icon: Tasks }
] as const

interface NavigationProps {
  expanded: Partial<Record<WorkspaceView, boolean>>
  onSelect: (view: WorkspaceView) => void
  menuOpen: boolean
  onMenu: () => void
}

export function CueiaMark() {
  return <span className="brand-symbol" aria-hidden="true"><i /><i /><i /></span>
}

export function WorkspaceNavigation({ expanded, onSelect, menuOpen, onMenu }: NavigationProps) {
  return <nav className="workspace-nav" aria-label="Workspace">
    <div role="img" aria-label="Cueia"><CueiaMark /></div>
    <div className="nav-destinations">
      {destinations.map(({ view, label, Icon }) =>
        <Button key={view} id={"nav-" + view} className="nav-control"
          aria-label={label} aria-controls={view + "-panel"} aria-expanded={!!expanded[view]}
          onPress={() => onSelect(view)}><Icon /><span>{label}</span></Button>
      )}
    </div>
    <Button id="workspace-menu" className="nav-control" aria-label="Menu"
      aria-haspopup="dialog" aria-expanded={menuOpen} onPress={onMenu}>
      <Menu /><span>Menu</span>
    </Button>
  </nav>
}

export function PanelContent({ children, visible = true }: { children: ReactNode; visible?: boolean }) {
  const reduced = useReducedMotion()
  return <motion.div initial={false} animate={{ opacity: visible ? 1 : 0 }}
    transition={{ duration: reduced ? 0 : 0.14 }} className="panel-content">
    {children}
  </motion.div>
}
