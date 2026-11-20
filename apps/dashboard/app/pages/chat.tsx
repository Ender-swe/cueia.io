import { useState } from "react"
import type { Schema } from "../api"

type Props = { data:Schema["Workspace"]; token:string; onPreference:(format:string)=>Promise<void>; onSignOut:()=>Promise<void> }

export function Chat({data,onPreference,onSignOut}:Props) {
  const [menu,setMenu]=useState(false)
  const [error,setError]=useState("")
  return <main className="shell">
    <nav aria-label="Workspace"><strong>Cueia</strong><button aria-expanded={menu} onClick={()=>setMenu(!menu)}>Menu</button></nav>
    {menu && <aside className="menu"><h2>Preferences</h2><label>Time format <select value={data.preferences.timeFormat ?? "12h"} onChange={e=>onPreference(e.target.value).catch(()=>setError("Could not save preference."))}><option value="12h">12 hour</option><option value="24h">24 hour</option></select></label>
      <button onClick={()=>onSignOut().catch(()=>setError("Could not sign out. Please retry."))}>Sign out</button></aside>}
    <section className="conversation"><header><h1>Main chat</h1></header>
      <div className="feed">
        {data.threads.map(t=><article key={t.id}><h2>{t.subject}</h2>{t.messages.map(m=><section key={m.id}><strong>{m.senderAddress}</strong><p>{m.bodyText}</p></section>)}</article>)}
        {data.events.map(e=><article key={e.id}><h2>{e.title}</h2><time dateTime={e.startsAt}>{new Intl.DateTimeFormat("en-US",{dateStyle:"full",timeStyle:"short",timeZone:e.timeZone,hour12:data.preferences.timeFormat!=="24h"}).format(new Date(e.startsAt))}</time></article>)}
        {error && <p role="alert">{error}</p>}
      </div>
    </section>
  </main>
}
