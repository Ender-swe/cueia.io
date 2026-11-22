import Check from "@react-spectrum/s2/icons/CheckmarkCircle";

export function TaskActivity({
  sampleMode,
  hour12,
}: {
  sampleMode: boolean;
  hour12: boolean;
}) {
  return (
    <section className="my-0 mx-4.75 py-5.5 px-0.5 [border-top:1px_solid_var(--color-line)] [&_h3]:flex [&_h3]:items-center [&_h3]:gap-2 [&_h3]:mb-4 [&_h3]:text-[14px] [&_h3]:font-semibold compact:mx-[9px] compact:py-3.5 @max-[220px]/tasks:[&_h3]:block compact:[&_h3]:text-[13px]">
      <h3>{sampleMode ? "Today" : "Today's activity"}</h3>
      {sampleMode ? (
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
      ) : (
        <p className="text-[13px] text-muted">No activity yet.</p>
      )}
    </section>
  );
}
