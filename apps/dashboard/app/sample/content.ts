import type { Schema } from "../api";
import workshop from "../../../server/internal/workspace/sample/workshop.json";
import registrationUrl from "../../../server/internal/workspace/sample/north-pier-registration.pdf?url";

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
