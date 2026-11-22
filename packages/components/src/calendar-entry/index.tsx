import Calendar from "@react-spectrum/s2/icons/Calendar";

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
