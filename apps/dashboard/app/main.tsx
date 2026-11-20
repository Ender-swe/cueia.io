import { createRoot } from "react-dom/client"
import { ClerkProvider, RedirectToSignIn, useAuth } from "@clerk/react"
import { Provider } from "@react-spectrum/s2/Provider"
import { BrowserRouter, Link, Outlet, Route, Routes, useHref, useMatch, useNavigate } from "react-router"
import { WorkspacePage } from "./pages/workspace"
import { SampleWorkspace } from "./sample-workspace"
import "@fontsource-variable/dm-sans"
import "./workspace.css"

function RequireSignIn() {
  const { isLoaded, isSignedIn, sessionId } = useAuth()
  const atHome = useMatch("/")

  if (!isLoaded) return <main className="entry" role="status">Checking sign-in…</main>
  if (!isSignedIn || !sessionId) {
    if (atHome) return <main className="entry">
      <h1>Sign in to Cueia</h1>
      <Link to="/workspace">Open sample workspace</Link>
    </main>
    return <RedirectToSignIn redirectUrl={window.location.origin + "/workspace"} />
  }
  return <Outlet key={sessionId} />
}

function Dashboard() {
  const navigate = useNavigate()
  const key = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY
  return <Provider colorScheme="dark" router={{ navigate, useHref }}>
    {key ? <ClerkProvider publishableKey={key}>
      <Routes>
        <Route element={<RequireSignIn />}>
          <Route path="/" element={<SampleWorkspace />} />
          <Route path="/workspace" element={<WorkspacePage />} />
        </Route>
      </Routes>
    </ClerkProvider> : <main className="entry" role="alert">Set VITE_CLERK_PUBLISHABLE_KEY to enable sign-in.</main>}
  </Provider>
}

createRoot(document.getElementById("root")!).render(<BrowserRouter><Dashboard /></BrowserRouter>)
