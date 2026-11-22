import { CalendarEntry, Turn } from "@cueia/components";
import { sample, reminderTime } from "../../sample-content";

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
        <CalendarEntry {...sample.events[0]} hour12={hour12} />
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
