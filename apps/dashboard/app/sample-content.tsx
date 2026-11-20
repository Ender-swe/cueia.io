import { CalendarEntry, Turn } from "@cueia/components";
import type { Schema } from "./api";
import workshop from "../../server/internal/workspace/sample/workshop.json";
import registrationUrl from "../../server/internal/workspace/sample/north-pier-registration.pdf?url";

export const samplePdf = {
  filename: workshop.filename,
  signedUrl: registrationUrl,
};

const event = {
  id: "sample-event",
  title: workshop.title,
  startsAt: workshop.startsAt,
  endsAt: workshop.endsAt,
  timeZone: workshop.timeZone,
};

export const sample: Schema["Workspace"] = {
  id: "sample-workspace",
  preferences: { timeFormat: "24h" },
  threads: [
    {
      id: "sample-thread",
      subject: workshop.subject,
      messages: [
        {
          id: "sample-invitation",
          senderAddress: workshop.senderAddress,
          sentAt: workshop.sentAt,
          bodyText: workshop.bodyText,
          attachments: [
            { fileId: "sample-registration", filename: workshop.filename },
          ],
        },
      ],
    },
  ],
  events: [event],
};

export function reminderTime(hour12: boolean) {
  const due = new Date(Date.parse(event.startsAt) - 30 * 60_000);
  return new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    minute: "2-digit",
    timeZone: event.timeZone,
    hour12,
  }).format(due);
}

export function SampleConversation({ hour12 }: { hour12: boolean }) {
  const stamp = (minute: string) =>
    hour12 ? `9:${minute} AM` : `09:${minute}`;
  return (
    <>
      <div className="flex items-center justify-center gap-3.5 text-muted text-[12px] leading-5 before:[content:''] before:h-px before:bg-[#32312e] before:flex-1 after:[content:''] after:h-px after:bg-[#32312e] after:flex-1">
        Monday, October 12
      </div>
      <Turn author="You" stamp={stamp("41")}>
        What’s on my calendar for Saturday?
      </Turn>
      <Turn author="Cueia" stamp={stamp("41")}>
        <p>
          You’re going to the North Pier mapmaking workshop. Mara asked you to
          bring a pencil and meet at the long table in Pier Hall.
        </p>
        <CalendarEntry {...event} hour12={hour12} />
      </Turn>
      <Turn author="You" stamp={stamp("42")}>
        Remind me 30 minutes before Saturday’s workshop.
      </Turn>
      <Turn author="Cueia" stamp={stamp("42")}>
        Reminder set for Saturday at <strong>{reminderTime(hour12)}</strong>.
      </Turn>
    </>
  );
}
