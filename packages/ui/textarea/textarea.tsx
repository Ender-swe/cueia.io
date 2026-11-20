import {
  TextArea as SpectrumTextArea,
  type TextAreaProps,
} from "@react-spectrum/s2/TextArea";
import { cn } from "../util/cn";

export function Textarea(props: TextAreaProps) {
  return (
    <SpectrumTextArea
      {...props}
      UNSAFE_className={cn(
        "[&_textarea::placeholder]:text-muted [&_textarea::placeholder]:opacity-[1] block [font-family:inherit] text-interface flex-1 min-w-0 [&_div:has(>_textarea)]:flex [&_div:has(>_textarea)]:bg-transparent [&_div:has(>_textarea)]:border-0 [&_div:has(>_textarea)]:[outline:0] [&_div:has(>_textarea)]:min-h-0 [&_div:has(>_textarea)]:shadow-none [&_textarea]:w-full [&_textarea]:h-6 [&_textarea]:min-h-6 [&_textarea]:max-h-[16dvh] [&_textarea]:overflow-auto [&_textarea]:resize-none [&_textarea]:p-0 [&_textarea]:[font:inherit] [&_textarea]:leading-[24px] [&_textarea]:text-ink [&_textarea]:bg-transparent [&_textarea]:border-0 [&_textarea]:[outline:0] supports-[field-sizing:content]:[&_textarea]:[field-sizing:content] supports-[field-sizing:content]:[&_textarea]:h-auto! max-w-full",
        props.UNSAFE_className,
      )}
    />
  );
}
