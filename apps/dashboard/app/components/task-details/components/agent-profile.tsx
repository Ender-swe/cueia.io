import Close from "@react-spectrum/s2/icons/Close";
import { AgentAvatar } from "@cueia/components";
import { Button } from "@cueia/ui";

export function AgentProfile({
  sampleMode,
  onClose,
}: {
  sampleMode: boolean;
  onClose: () => void;
}) {
  return (
    <div className="relative flex flex-col items-center p-[56px_16px_29px] [&_h2]:mt-3.25 [&_h2]:text-[23px] [&_h2]:leading-tight [&_h2]:font-semibold [&_h2]:tracking-[-0.04em] compact:p-[48px_5px_20px]">
      <Button
        className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8 p-0 rounded-[7px] text-muted bg-transparent hover:text-ink hover:bg-[#373633] absolute top-3.25 right-3.25"
        aria-label="Close Tasks"
        onPress={onClose}
      >
        <Close />
      </Button>
      <AgentAvatar large />
      <h2>Cueia</h2>
      <div className="flex items-center gap-1.75 mt-1.25 text-[13px] text-muted [&_i]:inline-block [&_i]:w-1.75 [&_i]:h-1.75 [&_i]:rounded-full compact:text-center compact:gap-1">
        <i className={sampleMode ? "bg-sage" : "bg-muted"} />
        {sampleMode ? "Connected" : "Agent disconnected"}
      </div>
    </div>
  );
}
