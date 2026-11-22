import { useRef, type RefObject } from "react";
import { DraftComposer } from "@cueia/components";
import { Button } from "@cueia/ui";
import Close from "@react-spectrum/s2/icons/Close";
import Attach from "@react-spectrum/s2/icons/Attach";
export type Draft = { text: string; file: File | null };
export function ChatComposer({
  composer,
  draft,
  updateDraft,
  sampleMode,
}: {
  composer: RefObject<HTMLDivElement | null>;
  draft: Draft;
  updateDraft: (change: Partial<Draft>) => void;
  sampleMode: boolean;
}) {
  const fileInput = useRef<HTMLInputElement>(null);
  return (
    <div className="flex-none p-[8px_clamp(14px,3.5cqw,32px)_max(12px,env(safe-area-inset-bottom))]">
      {draft.file && (
        <div className="flex items-center gap-1.5 text-apricot text-[13px] my-[0_6px] wrap-anywhere w-[min(100%,744px)] mx-auto">
          <Attach />
          {draft.file.name}
          <Button
            className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8 p-0 rounded-[7px] text-muted bg-transparent hover:text-ink hover:bg-[#373633]"
            aria-label="Remove attachment"
            onPress={() => updateDraft({ file: null })}
          >
            <Close />
          </Button>
        </div>
      )}
      <input
        ref={fileInput}
        type="file"
        hidden
        onChange={(e) => {
          if (e.target.files?.[0]) updateDraft({ file: e.target.files[0] });
          e.target.value = "";
        }}
      />
      <DraftComposer
        composerRef={composer}
        draft={draft.text}
        onChange={(text) => updateDraft({ text })}
        onAttach={() => fileInput.current?.click()}
        note={
          sampleMode
            ? "Sample conversation · Drafts stay in this tab"
            : "Draft only · Sending is unavailable until the agent connects."
        }
      />
    </div>
  );
}
