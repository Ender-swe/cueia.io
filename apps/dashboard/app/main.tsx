import { createRoot } from "react-dom/client"
import { ClerkProvider } from "@clerk/react"
import { Provider } from "@react-spectrum/s2/Provider"
import { BrowserRouter, Routes, Route, Link, useHref, useNavigate } from "react-router"
import { WorkspacePage } from "./pages/workspace"
import "./workspace.css"

function Dashboard() {
  const navigate = useNavigate()
  const key = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY
  return <Provider colorScheme="dark" router={{navigate, useHref}}>
    {key ? <ClerkProvider publishableKey={key}>
      <Routes>
        <Route path="/" element={<main className="entry"><h1>Cueia</h1><Link className="action" to="/workspace">Open sample workspace</Link></main>} />
        <Route path="/workspace" element={<WorkspacePage />} />
      </Routes>
    </ClerkProvider> : <main className="entry"><h1>Cueia</h1><p>Set VITE_CLERK_PUBLISHABLE_KEY to open the workspace.</p></main>}
  </Provider>
}

createRoot(document.getElementById("root")!).render(<BrowserRouter><Dashboard /></BrowserRouter>)
