import { useState } from "react";
import Bell from "@react-spectrum/s2/icons/Bell";
import Chevron from "@react-spectrum/s2/icons/ChevronRight";
import {
  Button,
  Heading,
  Modal,
  ModalContainer,
  ModalContent,
} from "@cueia/ui";
import { reminderTime } from "../../../sample-content";

export function ScheduledTasks({
  sampleMode,
  hour12,
}: {
  sampleMode: boolean;
  hour12: boolean;
}) {
  const [allTasks, setAllTasks] = useState(false);
  const time = sampleMode ? reminderTime(hour12) : "";
  return (
    <>
      <section className="my-0 mx-4.75 py-5.5 px-0.5 [border-top:1px_solid_var(--color-line)] [&_h3]:flex [&_h3]:items-center [&_h3]:gap-2 [&_h3]:mb-4 [&_h3]:text-[14px] [&_h3]:font-semibold compact:mx-[9px] compact:py-3.5 @max-[220px]/tasks:[&_h3]:block compact:[&_h3]:text-[13px]">
        <h3>
          Scheduled tasks
          {sampleMode && (
            <span className="text-muted font-medium text-[12px] whitespace-nowrap compact:ml-1">
              {" "}
              (1)
            </span>
          )}
        </h3>
        {sampleMode ? (
          <>
            <div className="flex items-start gap-2.75 [&>svg]:w-4.25 [&>svg]:h-4.25 [&>svg]:mt-0.75 [&>svg]:text-apricot [&_p_span]:mx-0.5 [&_small]:block [&_small]:mt-1.5 [&_small]:text-caption [&_small]:text-muted @max-[220px]/tasks:grid @max-[220px]/tasks:gap-1.5">
              <Bell />
              <div>
                <strong className="block text-interface font-label compact:text-[14px]">
                  Workshop reminder
                </strong>
                <p className="mt-0.5 text-[13px] leading-[1.45] text-muted">
                  Sat, Oct 17 <span>·</span> {time}
                </p>
              </div>
            </div>
            <Button
              className="inline-flex justify-start items-center gap-1 m-[9px_0_0_22px] p-1.25 border-0 text-apricot bg-transparent text-[12px] font-medium [&_svg]:w-3 [&_svg]:h-3 hover:underline compact:whitespace-normal @max-[220px]/tasks:ml-0 compact:text-left"
              onPress={() => setAllTasks(true)}
            >
              See all scheduled tasks
              <Chevron />
            </Button>
          </>
        ) : (
          <p className="text-[13px] text-muted">No scheduled tasks.</p>
        )}
      </section>
      <ModalContainer onDismiss={() => setAllTasks(false)}>
        {allTasks && (
          <Modal size="S" isDismissible>
            <Heading>Scheduled tasks</Heading>
            <ModalContent>
              <div className="font-sans text-ink leading-normal">
                <strong>Workshop reminder</strong>
                <p className="text-muted">
                  Saturday, October 17 at {time} · America/Chicago
                </p>
                <p className="text-muted">
                  30 minutes before the North Pier mapmaking workshop.
                </p>
              </div>
            </ModalContent>
          </Modal>
        )}
      </ModalContainer>
    </>
  );
}
