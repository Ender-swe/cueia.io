import { useState } from "react";
import Bell from "@react-spectrum/s2/icons/Bell";
import Check from "@react-spectrum/s2/icons/CheckmarkCircle";
import Close from "@react-spectrum/s2/icons/Close";
import Chevron from "@react-spectrum/s2/icons/ChevronRight";
import { AgentAvatar } from "@cueia/components";
import {
  Button,
  Heading,
  Modal,
  ModalContainer,
  ModalContent,
} from "@cueia/ui";
import { reminderTime } from "./sample-content";

export function TaskDetails({
  sampleMode,
  hour12,
  onClose,
}: {
  sampleMode: boolean;
  hour12: boolean;
  onClose: () => void;
}) {
  const [allTasks, setAllTasks] = useState(false);
  const time = sampleMode ? reminderTime(hour12) : "";
  return (
    <>
      <div className="agent-profile">
        <Button
          className="icon-button close-tasks"
          aria-label="Close Tasks"
          onPress={onClose}
        >
          <Close />
        </Button>
        <AgentAvatar large />
        <h2>Cueia</h2>
        <div className="connection-state">
          <i className={sampleMode ? "connected" : ""} />
          {sampleMode ? "Connected" : "Agent disconnected"}
        </div>
      </div>
      <section className="inspector-section">
        <h3>
          Scheduled tasks
          {sampleMode && <span className="task-count"> (1)</span>}
        </h3>
        {sampleMode ? (
          <>
            <div className="task-row">
              <Bell />
              <div>
                <strong>Workshop reminder</strong>
                <p>
                  Sat, Oct 17 <span>·</span> {time}
                </p>
              </div>
            </div>
            <Button className="all-tasks" onPress={() => setAllTasks(true)}>
              See all scheduled tasks
              <Chevron />
            </Button>
          </>
        ) : (
          <p className="empty-state">No scheduled tasks.</p>
        )}
      </section>
      <section className="inspector-section">
        <h3>{sampleMode ? "Today" : "Today's activity"}</h3>
        {sampleMode ? (
          <div className="task-row activity">
            <Check />
            <div>
              <strong>Scheduled reminder</strong>
              <p>North Pier mapmaking workshop</p>
              <small>{hour12 ? "9:42 AM" : "09:42"}</small>
            </div>
          </div>
        ) : (
          <p className="empty-state">No activity yet.</p>
        )}
      </section>
      <ModalContainer onDismiss={() => setAllTasks(false)}>
        {allTasks && (
          <Modal size="S" isDismissible>
            <Heading>Scheduled tasks</Heading>
            <ModalContent>
              <div className="task-summary">
                <strong>Workshop reminder</strong>
                <p>Saturday, October 17 at {time} · America/Chicago</p>
                <p>30 minutes before the North Pier mapmaking workshop.</p>
              </div>
            </ModalContent>
          </Modal>
        )}
      </ModalContainer>
    </>
  );
}
