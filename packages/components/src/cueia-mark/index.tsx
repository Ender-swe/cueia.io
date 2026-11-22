const brandRing =
  "[grid-area:1/1] rounded-[50%] border-[0.0625em] border-solid border-current border-r-transparent rotate-[-18deg]";

export function CueiaMark({ className }: { className?: string } = {}) {
  return (
    <span
      className={`grid flex-none place-items-center text-apricot ${className ?? "size-[33px] text-[32px] compact:size-7"}`}
      aria-hidden="true"
    >
      <i className={`${brandRing} size-[1em]`} />
      <i className={`${brandRing} size-[0.71875em]`} />
      <i className={`${brandRing} size-[0.4375em]`} />
    </span>
  );
}
