import { TextArea as SpectrumTextArea, type TextAreaProps } from "@react-spectrum/s2/TextArea"
import { cn } from "../util/cn"

export function Textarea(props: TextAreaProps) {
  return <SpectrumTextArea {...props} UNSAFE_className={cn("cue-textarea", props.UNSAFE_className)} />
}
