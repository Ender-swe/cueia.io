import { SampleTaskActivity } from "../../../sample/task-activity";

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
        <SampleTaskActivity hour12={hour12} />
      ) : (
        <p className="text-[13px] text-muted">No activity yet.</p>
      )}
    </section>
  );
}
