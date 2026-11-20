import { TextField, type TextFieldProps } from "@react-spectrum/s2/TextField";
import { cn } from "../util/cn";

export function Input(props: TextFieldProps) {
  return (
    <TextField
      {...props}
      UNSAFE_className={cn(
        "w-full [font-family:inherit] [&_div:has(>_input)]:bg-ash [&_div:has(>_input)]:[border:1px_solid_#3b3a37] [&_div:has(>_input)]:rounded-lg [&_div:has(>_input)]:min-h-9 [&_div:has(>_input):focus-within]:[outline:1px_solid_var(--color-apricot)] [&_input]:text-ink [&_input]:[font:inherit] [&_input]:text-[14px] [&_input]:[outline:0] [&_input]:min-w-0 [&_input::placeholder]:text-muted [&_input::placeholder]:opacity-[1] min-w-0 max-w-full",
        props.UNSAFE_className,
      )}
    />
  );
}
