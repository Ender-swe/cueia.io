import Add from "@react-spectrum/s2/icons/Add";
import Microphone from "@react-spectrum/s2/icons/Microphone";
import { Button } from "@cueia/ui/button";
import { Textarea } from "@cueia/ui/textarea";
import { useId } from "react";

export function DraftComposer({
  draft,
  composerRef,
  onChange,
  onAttach,
  onMicrophone,
  note,
}: {
  draft: string;
  composerRef?: React.Ref<HTMLDivElement>;
  onChange: (value: string) => void;
  onAttach?: () => void;
  onMicrophone?: () => void;
  note?: string;
}) {
  const id = useId();
  return (
    <div ref={composerRef} className="w-[min(100%,_744px)] mx-auto">
      <div className="flex items-center gap-1.5 min-h-13 py-1.5 px-2.5 [border:1px_solid_#484641] rounded-[10px] bg-composer focus-within:border-apricot compact:p-1 compact:gap-0.5">
        <Button
          className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8.5 p-0 rounded-[7px] text-ink bg-transparent [&:enabled:hover]:text-ink [&:enabled:hover]:bg-[#373633] disabled:text-muted disabled:opacity-[0.65] compact:w-7"
          aria-label="Add attachment"
          isDisabled={!onAttach}
          onPress={onAttach}
        >
          <Add />
        </Button>
        <Textarea
          aria-label="Message"
          aria-describedby={note ? id : undefined}
          placeholder="Message…"
          value={draft}
          onChange={onChange}
        />
        <Button
          className="inline-flex items-center justify-center flex-none border-0 min-w-0 size-8.5 p-0 rounded-[7px] text-ink bg-transparent [&:enabled:hover]:text-ink [&:enabled:hover]:bg-[#373633] disabled:text-muted disabled:opacity-[0.65] compact:w-7"
          aria-label="Microphone"
          isDisabled={!onMicrophone}
          onPress={onMicrophone}
        >
          <Microphone />
        </Button>
      </div>
      {note && (
        <p id={id} className="pt-1.5 text-center text-muted text-caption">
          {note}
        </p>
      )}
    </div>
  );
}
