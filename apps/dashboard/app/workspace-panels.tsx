import { useState } from "react"
import Search from "@react-spectrum/s2/icons/Search"
import File from "@react-spectrum/s2/icons/FileText"
import { AttachmentPreview } from "@cueia/components"
import { Input } from "@cueia/ui/input"
import type { Schema } from "./api"
import { samplePdf } from "./sample-content"

export function SearchBox({ label, value, onChange }: {
  label: string; value: string; onChange: (value: string) => void
}) {
  return <div className="search-box">
    <Search aria-hidden="true" />
    <Input type="search" aria-label={label} placeholder="Search" value={value} onChange={onChange} />
  </div>
}

export function FindPanel({ data, onSelect }: { data: Schema["Workspace"]; onSelect: (id: string) => void }) {
  const [query, setQuery] = useState("")
  const term = query.trim().toLowerCase()
  const entries = [
    ...data.threads.map(t => ({ id: t.id, type: "Email", title: t.subject, text: t.messages.map(m => m.senderAddress + " " + m.bodyText).join(" ") })),
    ...data.events.map(e => ({ id: e.id, type: "Calendar", title: e.title, text: "" }))
  ].filter(item => (item.title + " " + item.text).toLowerCase().includes(term))
  return <>
    <SearchBox label="Find in workspace" value={query} onChange={setQuery} />
    <h2 className="panel-title">Find</h2>
    {entries.map(entry => <button className="result-row" key={entry.id} onClick={() => onSelect(entry.id)}>
      <small>{entry.type}</small>{entry.title}
    </button>)}
    {!entries.length && <p className="empty-state">No matching records.</p>}
  </>
}

export function SampleFiles() {
  const [query, setQuery] = useState("")
  const [open, setOpen] = useState(false)
  return <>
    <SearchBox label="Search files" value={query} onChange={setQuery} />
    {samplePdf.filename.includes(query.trim().toLowerCase())
      ? <button className="result-row file-row" onClick={() => setOpen(true)}><File />{samplePdf.filename}</button>
      : <p className="empty-state">No matching files.</p>}
    <AttachmentPreview attachment={open ? samplePdf : null} onClose={() => setOpen(false)} />
  </>
}
