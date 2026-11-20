import { TextField, type TextFieldProps } from "@react-spectrum/s2/TextField"
import { cn } from "../util/cn"

export function Input(props: TextFieldProps) {
  return <TextField {...props} UNSAFE_className={cn("cue-input", props.UNSAFE_className)} />
}
