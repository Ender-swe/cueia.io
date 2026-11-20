import { useState } from "react"
import { Files } from "./files"
import type { Schema } from "../api"

type Props = { data:Schema["Workspace"]; token:string; onPreference:(format:string)=>Promise<void>; onSignOut:()=>Promise<void> }

export function Chat({data,token,onPreference,onSignOut}:Props) {
  const [panel,setPanel] = useState<"chat" | "files" | null>(() => window.matchMedia("(max-width: 760px)").matches ? null : "chat")
  const [draft,setDraft] = useState("")
  const [menu,setMenu]=useState(false)
  const [error,setError]=useState("")
  return <main className="shell">
    <nav aria-label="Workspace"><strong>Cueia</strong>
      <button aria-expanded={panel === "chat"} onClick={() => setPanel("chat")}>Chat</button>
      <button aria-expanded={panel === "files"} onClick={() => setPanel("files")}>Files</button>
      <button aria-expanded={menu} onClick={()=>setMenu(!menu)}>Menu</button></nav>
    {menu && <aside className="menu"><h2>Preferences</h2><label>Time format <select value={data.preferences.timeFormat ?? "12h"} onChange={e=>onPreference(e.target.value).catch(()=>setError("Could not save preference."))}><option value="12h">12 hour</option><option value="24h">24 hour</option></select></label>
      <button onClick={()=>onSignOut().catch(()=>setError("Could not sign out. Please retry."))}>Sign out</button></aside>}
    {panel && <aside className="sidebar" aria-label={panel === "files" ? "Files" : "Chats"}>
      <div className="panel-heading"><h2>{panel === "files" ? "Files" : "Chats"}</h2><button aria-label="Close sidebar" onClick={() => setPanel(null)}>×</button></div>
      {panel === "files" ? <Files token={token} /> : <>
        <button aria-current="page" onClick={() => setPanel(null)}>Main chat</button>
        <h2>Side chats</h2><p>No side chats yet.</p>
      </>}
    </aside>}
    <section className="conversation"><header><h1>Main chat</h1></header>
      <div className="feed">
        {data.threads.map(t=><article key={t.id}><h2>{t.subject}</h2>{t.messages.map(m=><section key={m.id}><strong>{m.senderAddress}</strong><p>{m.bodyText}</p></section>)}</article>)}
        {data.events.map(e=><article key={e.id}><h2>{e.title}</h2><time dateTime={e.startsAt}>{new Intl.DateTimeFormat("en-US",{dateStyle:"full",timeStyle:"short",timeZone:e.timeZone,hour12:data.preferences.timeFormat!=="24h"}).format(new Date(e.startsAt))}</time></article>)}
        {error && <p role="alert">{error}</p>}
      </div>
      <div className="composer"><label htmlFor="draft">Message</label>
        <textarea id="draft" rows={2} value={draft} placeholder="Write a message…" onChange={e => setDraft(e.target.value)} />
      </div>
    </section>
  </main>
}
