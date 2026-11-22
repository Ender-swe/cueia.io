import { AgentProfile } from "./components/agent-profile";
import { ScheduledTasks } from "./components/scheduled-tasks";
import { TaskActivity } from "./components/task-activity";

export function TaskDetails({
  sampleMode,
  hour12,
  onClose,
}: {
  sampleMode: boolean;
  hour12: boolean;
  onClose: () => void;
}) {
  return (
    <>
      <AgentProfile sampleMode={sampleMode} onClose={onClose} />
      <ScheduledTasks sampleMode={sampleMode} hour12={hour12} />
      <TaskActivity sampleMode={sampleMode} hour12={hour12} />
    </>
  );
}
