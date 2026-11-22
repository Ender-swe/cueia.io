const portrait = new URL("../assets/cueia-orbit.webp", import.meta.url).href;

export function AgentAvatar({ large = false }: { large?: boolean }) {
  return (
    <img
      className={
        large
          ? "size-29.5 rounded-full compact:w-[min(80%,_80px)] compact:h-auto"
          : "block size-[21px] rounded-full"
      }
      src={portrait}
      alt=""
      draggable={false}
    />
  );
}
