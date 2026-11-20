import { useEffect, useRef, useState } from "react"
import { AttachmentPreview } from "@cueia/components"
import { api, type Schema } from "../api"

export function Files({ token }: { token: string }) {
  const [directory, setDirectory] = useState(".")
  const [query, setQuery] = useState("")
  const [revision, refresh] = useState(0)
  const [entries, setEntries] = useState<Schema["Entry"][] | null>(null)
  const [error, setError] = useState("")
  const [attachment, setAttachment] = useState<{ filename: string; signedUrl: string } | null>(null)
  const selected = useRef<HTMLButtonElement | null>(null)
  const previewRequest = useRef<AbortController | null>(null)
  const name = query.trim()

  useEffect(() => {
    const controller = new AbortController()
    const path = name ? "/files/search?name=" + encodeURIComponent(name) : "/files?directory=" + encodeURIComponent(directory)
    setEntries(null)
    setError("")
    api<Schema["Entry"][]>(path, token, "GET", undefined, controller.signal)
      .then(result => { if (!controller.signal.aborted) setEntries(result) })
      .catch(err => { if (!controller.signal.aborted) setError(err.message) })
    return () => {
      controller.abort()
      previewRequest.current?.abort()
    }
  }, [token, directory, name, revision])

  function openDirectory(path: string) {
    setQuery("")
    setDirectory(path)
  }

  async function preview(entry: Schema["Entry"], button: HTMLButtonElement) {
    previewRequest.current?.abort()
    const controller = new AbortController()
    previewRequest.current = controller
    selected.current = button
    setError("")
    try {
      const { fileId } = await api<{ fileId: string }>("/documents", token, "POST", { path: entry.path }, controller.signal)
      const link = await api<Schema["Preview"]>("/documents/" + fileId + "/preview-links", token, "POST", undefined, controller.signal)
      if (!controller.signal.aborted) setAttachment({ filename: entry.name, signedUrl: link.url })
    } catch (err) {
      if (!controller.signal.aborted) setError(err instanceof Error ? err.message : "Could not open PDF.")
    }
  }

  return <>
    <label>Search files<input type="search" value={query} onChange={e => setQuery(e.target.value)} /></label>
    <div className="directory-controls">
      {directory !== "." && <button onClick={() => openDirectory(directory.split("/").slice(0, -1).join("/") || ".")}>Up</button>}
      <span>{directory === "." ? "Workspace" : directory}</span>
      <button onClick={() => refresh(n => n + 1)}>Refresh</button>
    </div>
    {error && <p role="alert">{error}</p>}
    {entries === null && !error && <p role="status">Loading files…</p>}
    {entries?.length === 0 && <p>{name ? "No matching files." : "This directory is empty."}</p>}
    <ul className="file-list">
      {entries?.map(entry => <li key={entry.path}>
        {entry.kind === "directory" ? <button onClick={() => openDirectory(entry.path)}>{entry.name}/</button>
          : /\.pdf$/i.test(entry.name) ? <button onClick={e => void preview(entry, e.currentTarget)}>{entry.name}</button>
          : <span>{entry.name}</span>}
        {name && <small>{entry.path}</small>}
      </li>)}
    </ul>
    <AttachmentPreview attachment={attachment} onClose={() => {
      setAttachment(null)
      selected.current?.focus()
    }} />
  </>
}
