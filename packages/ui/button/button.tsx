import { ActionButton, type ActionButtonProps } from "@react-spectrum/s2/ActionButton"
import { cn } from "../util/cn"

export interface ButtonProps extends ActionButtonProps { className?: string }

export function Button({ className, ...props }: ButtonProps) {
  return <ActionButton isQuiet {...props} UNSAFE_className={cn("cue-button", props.UNSAFE_className, className)} />
}
