import Attach from "@react-spectrum/s2/icons/Attach"
import Calendar from "@react-spectrum/s2/icons/Calendar"
import Add from "@react-spectrum/s2/icons/Add"
import Microphone from "@react-spectrum/s2/icons/Microphone"
import { Button } from "@cueia/ui/button"
import { Textarea } from "@cueia/ui/textarea"
import { useId, type ReactNode } from "react"

const portrait = new URL("./assets/cueia-orbit.webp", import.meta.url).href

export function AgentAvatar({ large = false }: { large?: boolean }) {
  return <img className={large ? "agent-portrait" : "agent-avatar"} src={portrait} alt="" draggable={false} />
}

export function Turn({ author, stamp, children }: { author: "Cueia" | "You"; stamp?: string; children: ReactNode }) {
  return <article className={author === "You" ? "turn outgoing" : "turn incoming"}>
    <div className="turn-byline">
      {author === "Cueia" && <AgentAvatar />}<strong>{author}</strong>{stamp && <span>{stamp}</span>}
    </div>
    <div className="bubble">{children}</div>
  </article>
}

interface Message {
  id: string
  senderAddress: string
  bodyText: string
  sentAt: string
  attachments: { fileId: string; filename: string }[]
}

export function Correspondence({ subject, messages }: { subject: string; messages: Message[] }) {
  return <article className="mail-thread">
    <h2>{subject}</h2>
    {messages.map(message => <section key={message.id} className="message">
      <div className="mail-byline"><strong>{message.senderAddress}</strong>
        <span>{new Intl.DateTimeFormat("en-US", { dateStyle: "medium" }).format(new Date(message.sentAt))}</span>
      </div>
      <p>{message.bodyText}</p>
      {message.attachments.map(file => <div key={file.fileId} className="attachment-name"><Attach />{file.filename}</div>)}
    </section>)}
  </article>
}

export function CalendarEntry({ title, startsAt, timeZone, hour12 }: {
  title: string; startsAt: string; timeZone: string; hour12: boolean
}) {
  const start = new Date(startsAt)
  const format = (options: Intl.DateTimeFormatOptions) =>
    new Intl.DateTimeFormat("en-US", { ...options, timeZone }).format(start)
  return <article className="calendar-entry">
    <div className="date-tile" aria-hidden="true"><span>{format({ month: "short" })}</span><b>{format({ day: "numeric" })}</b></div>
    <div>
      <h3>{title}</h3>
      <time dateTime={startsAt}>{format({ weekday: "short", month: "short", day: "numeric" })} · {format({ hour: "numeric", minute: "2-digit", hour12, timeZoneName: "short" })}</time>
    </div>
    <Calendar />
  </article>
}

export function DraftComposer({ draft, onChange, onAttach, onMicrophone, note }: {
  draft: string
  onChange: (value: string) => void
  onAttach?: () => void
  onMicrophone?: () => void
  note?: string
}) {
  const id = useId()
  return <div className="draft-composer">
    <div className="composer-line">
      <Button className="composer-tool" aria-label="Add attachment" isDisabled={!onAttach} onPress={onAttach}><Add /></Button>
      <Textarea aria-label="Message" aria-describedby={note ? id : undefined}
        placeholder="Message…" value={draft} onChange={onChange} />
      <Button className="composer-tool mic" aria-label="Microphone" isDisabled={!onMicrophone} onPress={onMicrophone}><Microphone /></Button>
    </div>
    {note && <p id={id} className="composer-status">{note}</p>}
  </div>
}
