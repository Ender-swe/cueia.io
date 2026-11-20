import Attach from "@react-spectrum/s2/icons/Attach";
import Calendar from "@react-spectrum/s2/icons/Calendar";
import Add from "@react-spectrum/s2/icons/Add";
import Microphone from "@react-spectrum/s2/icons/Microphone";
import { Button } from "@cueia/ui/button";
import { Textarea } from "@cueia/ui/textarea";
import { useId, type ReactNode } from "react";

const portrait = new URL("./assets/cueia-orbit.webp", import.meta.url).href;

export function AgentAvatar({ large = false }: { large?: boolean }) {
  return (
    <img
      className={
        large
          ? "size-29.5 rounded-full compact:w-[min(80%,_80px)] compact:h-auto"
          : "block size-[21px] rounded-full"
      }
      src={portrait}
      alt=""
      draggable={false}
    />
  );
}

export function Turn({
  author,
  stamp,
  children,
}: {
  author: "Cueia" | "You";
  stamp?: string;
  children: ReactNode;
}) {
  const outgoing = author === "You";
  return (
    <article
      className={`flex flex-col gap-[7px] max-w-[min(90%,580px)] narrow-chat:max-w-[96%] ${outgoing ? "self-end items-end" : "items-start"}`}
    >
      <div
        className={`flex items-center gap-2 min-h-5.5 ${outgoing ? "flex-row-reverse" : ""}`}
      >
        {!outgoing && <AgentAvatar />}
        <strong className="text-[13px] font-semibold">{author}</strong>
        {stamp && (
          <span className="text-caption tabular-nums text-muted">{stamp}</span>
        )}
      </div>
      <div
        data-message-bubble
        className={`w-fit max-w-full rounded-[11px] px-[17px] py-[13px] text-message [&_strong]:font-semibold compact:px-[13px] compact:py-[11px] ${outgoing ? "rounded-tr-[3px] bg-umber text-cream" : "rounded-tl-[3px] bg-ash"}`}
      >
        {children}
      </div>
    </article>
  );
}

interface Message {
  id: string;
  senderAddress: string;
  bodyText: string;
  sentAt: string;
  attachments: { fileId: string; filename: string }[];
}

export function Correspondence({
  subject,
  messages,
}: {
  subject: string;
  messages: Message[];
}) {
  return (
    <article className="bg-ash rounded-[3px_10px_10px] p-4.5 [&_h2]:font-semibold [&_h2]:text-[18px] [&_h2]:leading-[1.35]">
      <h2>{subject}</h2>
      {messages.map((message) => (
        <section key={message.id} className="pt-3.5">
          <div className="flex flex-wrap gap-[4px_10px] text-[13px] text-muted">
            <strong className="text-ink">{message.senderAddress}</strong>
            <span>
              {new Intl.DateTimeFormat("en-US", { dateStyle: "medium" }).format(
                new Date(message.sentAt),
              )}
            </span>
          </div>
          <p className="mt-2 whitespace-pre-wrap">{message.bodyText}</p>
          {message.attachments.map((file) => (
            <div
              key={file.fileId}
              className="flex items-center gap-1.5 text-apricot text-[13px] my-[10px_0] wrap-anywhere"
            >
              <Attach />
              {file.filename}
            </div>
          ))}
        </section>
      ))}
    </article>
  );
}

export function CalendarEntry({
  title,
  startsAt,
  timeZone,
  hour12,
}: {
  title: string;
  startsAt: string;
  timeZone: string;
  hour12: boolean;
}) {
  const start = new Date(startsAt);
  const format = (options: Intl.DateTimeFormatOptions) =>
    new Intl.DateTimeFormat("en-US", { ...options, timeZone }).format(start);
  return (
    <article className="flex items-center gap-[13px] py-3 px-3.5 [border:1px_solid_#474540] rounded-lg bg-[#252523] in-[[data-message-bubble]]:mt-[13px] [&_h3]:text-[14px] [&_h3]:leading-[1.35] [&_h3]:font-label [&>svg]:w-4 [&>svg]:text-muted [&>svg]:ml-auto narrow-chat:gap-[9px] narrow-chat:p-[9px] narrow-chat:[&>svg]:hidden">
      <div
        className="flex-none flex flex-col items-center justify-center w-[39px] h-[43px] rounded-sm bg-[#3b302a] text-[#e9b99a] leading-none"
        aria-hidden="true"
      >
        <span className="text-[10px] font-semibold">
          {format({ month: "short" })}
        </span>
        <b className="text-[21px] font-label mt-[3px] tabular-nums">
          {format({ day: "numeric" })}
        </b>
      </div>
      <div>
        <h3>{title}</h3>
        <time
          className="block mt-[3px] text-[12px] text-muted"
          dateTime={startsAt}
        >
          {format({ weekday: "short", month: "short", day: "numeric" })} ·{" "}
          {format({
            hour: "numeric",
            minute: "2-digit",
            hour12,
            timeZoneName: "short",
          })}
        </time>
      </div>
      <Calendar />
    </article>
  );
}

export function DraftComposer({
  draft,
  composerRef,
  onChange,
  onAttach,
  onMicrophone,
  note,
}: {
  draft: string;
  composerRef?: React.Ref<HTMLDivElement>;
  onChange: (value: string) => void;
  onAttach?: () => void;
  onMicrophone?: () => void;
  note?: string;
}) {
  const id = useId();
  return (
    <div ref={composerRef} className="w-[min(100%,_744px)] mx-auto">
      <div className="flex items-center gap-1.5 min-h-13 py-1.5 px-2.5 [border:1px_solid_#484641] rounded-[10px] bg-composer focus-within:border-apricot compact:p-1 compact:gap-0.5">
        <Button
          className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8.5 p-0 rounded-[7px] text-ink bg-transparent [&:enabled:hover]:text-ink [&:enabled:hover]:bg-[#373633] disabled:text-muted disabled:opacity-[0.65] compact:w-7"
          aria-label="Add attachment"
          isDisabled={!onAttach}
          onPress={onAttach}
        >
          <Add />
        </Button>
        <Textarea
          aria-label="Message"
          aria-describedby={note ? id : undefined}
          placeholder="Message…"
          value={draft}
          onChange={onChange}
        />
        <Button
          className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8.5 p-0 rounded-[7px] text-ink bg-transparent [&:enabled:hover]:text-ink [&:enabled:hover]:bg-[#373633] disabled:text-muted disabled:opacity-[0.65] compact:w-7"
          aria-label="Microphone"
          isDisabled={!onMicrophone}
          onPress={onMicrophone}
        >
          <Microphone />
        </Button>
      </div>
      {note && (
        <p id={id} className="pt-1.5 text-center text-muted text-caption">
          {note}
        </p>
      )}
    </div>
  );
}
