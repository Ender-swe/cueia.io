export function cn(...names: (string | false | null | undefined)[]) {
  return names.filter(Boolean).join(" ");
}
