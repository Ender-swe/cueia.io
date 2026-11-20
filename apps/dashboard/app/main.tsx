import { createRoot } from "react-dom/client"
import { ClerkProvider } from "@clerk/react"
import { Provider } from "@react-spectrum/s2/Provider"
import { BrowserRouter, Routes, Route, useHref, useNavigate } from "react-router"
import { WorkspacePage } from "./pages/workspace"
import { SampleWorkspace } from "./sample-workspace"
import "@fontsource-variable/dm-sans"
import "./workspace.css"

function Dashboard() {
  const navigate = useNavigate()
  const key = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY
  const routes = <Routes>
    <Route path="/" element={<SampleWorkspace canSignIn={!!key} />} />
    <Route path="/workspace" element={key ? <WorkspacePage /> : <main className="entry"><h1>Cueia</h1><p>Set VITE_CLERK_PUBLISHABLE_KEY to open the workspace.</p><a href="/">View sample workspace</a></main>} />
  </Routes>
  return <Provider colorScheme="dark" router={{ navigate, useHref }}>
    {key ? <ClerkProvider publishableKey={key}>{routes}</ClerkProvider> : routes}
  </Provider>
}

createRoot(document.getElementById("root")!).render(<BrowserRouter><Dashboard /></BrowserRouter>)
