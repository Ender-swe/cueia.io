import { useAuth, useClerk, RedirectToSignIn } from "@clerk/react"
import { useEffect, useState } from "react"
import { api, type Schema } from "../api"
import { Chat } from "./chat"

export function WorkspacePage() {
  const {isLoaded, isSignedIn, getToken, sessionId} = useAuth()
  const {signOut} = useClerk()
  const [workspace,setWorkspace] = useState<{sessionId: string; token:string; data:Schema["Workspace"]} | null>(null)
  const [error,setError] = useState(false)
  const [attempt,retry] = useState(0)

  useEffect(() => {
    if (!isSignedIn || !sessionId) return
    const controller = new AbortController()
    const {signal} = controller
    setWorkspace(null)
    setError(false)
    async function open() {
      const clerk = await getToken()
      if (!clerk) throw new Error("Sign in again")
      const session = await api<Schema["Session"]>("/session",clerk,"POST",undefined,signal)
      const data = await api<Schema["Workspace"]>("/workspace",session.token,"GET",undefined,signal)
      if (!signal.aborted) setWorkspace({sessionId: sessionId!, token:session.token,data})
    }
    open().catch(() => {if (!signal.aborted) setError(true)})
    return () => controller.abort()
  },[sessionId,isSignedIn,getToken,attempt])

  if (isLoaded && !isSignedIn) return <RedirectToSignIn redirectUrl={window.location.origin + "/workspace"} />
  if (error) return <main className="entry" role="alert">Could not open your workspace. <button onClick={() => retry(n=>n+1)}>Retry</button></main>
  if (!workspace || workspace.sessionId !== sessionId) return <main className="entry" role="status">Opening your workspace…</main>
  return <Chat key={workspace.sessionId} data={workspace.data} token={workspace.token}
    onPreference={async timeFormat => {
      const preferences = await api<Schema["Preferences"]>("/workspace/preferences",workspace.token,"PATCH",{timeFormat})
      setWorkspace({...workspace,data:{...workspace.data,preferences}})
    }}
    onSignOut={async () => {
      await api("/session",workspace.token,"DELETE")
      setWorkspace(null)
      await signOut({redirectUrl:"/"})
    }} />
}
