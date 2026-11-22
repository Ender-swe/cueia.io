import Check from "@react-spectrum/s2/icons/CheckmarkCircle";

export function SampleTaskActivity({ hour12 }: { hour12: boolean }) {
  return (
    <div className="flex items-start gap-2.75 [&>svg]:w-4.25 [&>svg]:h-4.25 [&>svg]:mt-0.75 [&>svg]:text-muted [&_p_span]:mx-0.5 @max-[220px]/tasks:grid @max-[220px]/tasks:gap-1.5">
      <Check />
      <div>
        <strong className="block text-interface font-label compact:text-[14px]">
          Scheduled reminder
        </strong>
        <p className="mt-0.5 text-[13px] leading-[1.45] text-muted">
          North Pier mapmaking workshop
        </p>
        <small className="block mt-1.5 text-caption text-muted">
          {hour12 ? "9:42 AM" : "09:42"}
        </small>
      </div>
    </div>
  );
}
