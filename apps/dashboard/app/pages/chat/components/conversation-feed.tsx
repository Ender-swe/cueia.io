import type { RefObject } from "react";
import { CalendarEntry, Correspondence } from "@cueia/components";
import { SampleConversation } from "../../../components/sample-conversation";
import type { Schema } from "../../../api";
export function ConversationFeed({
  feed,
  active,
  sampleMode,
  hour12,
  data,
  thread,
  event,
  sideChat,
}: {
  feed: RefObject<HTMLDivElement | null>;
  active: string;
  sampleMode: boolean;
  hour12: boolean;
  data: Schema["Workspace"];
  thread?: Schema["Workspace"]["threads"][number];
  event?: Schema["Workspace"]["events"][number];
  sideChat?: { title: string };
}) {
  return (
    <div
      className="min-h-0 flex-1 overflow-y-auto px-[clamp(18px,4.5cqw,40px)] scrollbar-thin wrap-anywhere focus:[outline:none]"
      tabIndex={-1}
      ref={feed}
    >
      <div className="w-[min(100%,720px)] mx-auto py-6.5 flex flex-col gap-[clamp(20px,3vh,31px)] compact:py-4.5 compact:gap-4.5">
        {active === "main" &&
          (sampleMode ? (
            <SampleConversation hour12={hour12} />
          ) : (
            <>
              {data.threads.map((item) => (
                <Correspondence key={item.id} {...item} />
              ))}
              {data.events.map((item) => (
                <CalendarEntry key={item.id} {...item} hour12={hour12} />
              ))}
            </>
          ))}
        {thread && <Correspondence {...thread} />}
        {event && <CalendarEntry {...event} hour12={hour12} />}
        {sideChat && (
          <p className="text-[13px] text-muted text-center">
            Start a draft in {sideChat.title.toLowerCase()}.
          </p>
        )}
      </div>
    </div>
  );
}
