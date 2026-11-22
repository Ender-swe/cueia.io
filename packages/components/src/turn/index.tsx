import type { ReactNode } from "react";
import { AgentAvatar } from "../agent-avatar";

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
