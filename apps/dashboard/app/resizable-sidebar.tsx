import { useLayoutEffect, useRef, useState, type ReactNode } from "react";
import type { WorkspaceView } from "@cueia/components";

interface Props {
  side: "context" | "tasks";
  panel: WorkspaceView | null;
  otherOpen: boolean;
  onEscape?: () => void;
  children: ReactNode;
}

export const compactWorkspace = "(max-width: 760px)";

interface Space {
  width: number;
  height: number;
  compact: boolean;
}

function sizeRange(side: Props["side"], space: Space, otherOpen: boolean) {
  if (space.compact && side === "context") {
    return {
      min: 96,
      initial: Math.min(224, space.height * 0.3),
      max: Math.min(360, space.height * 0.45),
    };
  }
  if (space.compact) {
    return {
      min: 112,
      initial: Math.min(224, window.innerWidth * 0.32),
      max: Math.min(320, space.width * 0.48),
    };
  }
  // Reserve space for the conversation when either or both panels are visible.
  const max = Math.min(480, (space.width - 280) / (otherOpen ? 2 : 1));
  return side === "context"
    ? { min: 176, initial: Math.min(244, window.innerWidth * 0.17), max }
    : { min: 192, initial: Math.min(336, window.innerWidth * 0.23), max };
}

export function ResizableSidebar({
  side,
  panel,
  otherOpen,
  onEscape,
  children,
}: Props) {
  const container = useRef<HTMLDivElement>(null);
  const drag = useRef<{ coordinate: number; size: number } | null>(null);
  const [space, setSpace] = useState({ width: 0, height: 0, compact: false });
  const [preferred, setPreferred] = useState<{
    compact?: number;
    desktop?: number;
  }>({});

  useLayoutEffect(() => {
    const body = container.current!.closest("[data-workspace-body]")!;
    const media = matchMedia(compactWorkspace);
    const measure = () =>
      setSpace({
        width: body.clientWidth,
        height: body.clientHeight,
        compact: media.matches,
      });

    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(body);
    media.addEventListener("change", measure);
    return () => {
      observer.disconnect();
      media.removeEventListener("change", measure);
    };
  }, []);

  const context = side === "context";
  const stacked = space.compact && context;
  const mode = space.compact ? "compact" : "desktop";
  const range = sizeRange(side, space, otherOpen);
  const max = Math.max(1, Math.floor(range.max));
  const min = Math.min(max, range.min);
  const clamp = (value: number) =>
    Math.max(min, Math.min(max, Math.round(value)));
  const size = clamp(preferred[mode] ?? range.initial);
  const direction = context ? 1 : -1;
  function resize(value: number) {
    setPreferred((previous) => ({ ...previous, [mode]: clamp(value) }));
  }

  return (
    <div
      ref={container}
      hidden={!panel}
      className={
        stacked
          ? "relative flex-none min-h-0 h-(--pane-size)"
          : "relative flex-none min-h-0 w-(--pane-size)"
      }
      onKeyDown={(event) => {
        if (event.key === "Escape") onEscape?.();
      }}
      style={{ "--pane-size": size + "px" } as React.CSSProperties}
    >
      {children}
      {panel && (
        <div
          className={`[--handle-width:10px] absolute z-1 inset-y-0 w-(--handle-width) cursor-col-resize touch-none select-none after:[content:''] after:absolute after:inset-[0_auto_0_50%] after:w-0.5 [&:hover::after]:bg-muted [&:active::after]:bg-muted focus-visible:[outline:none] [&:focus-visible::after]:bg-apricot pointer-coarse:[--handle-width:18px] ${context ? "left-[calc(100%-var(--handle-width)/2)] compact:inset-[auto_0_calc(var(--handle-width)_/_-2)] compact:w-auto compact:h-[var(--handle-width)] compact:cursor-row-resize compact:after:inset-[50%_0_auto] compact:after:w-auto compact:after:h-0.5" : "left-[calc(var(--handle-width)/-2)]"}`}
          role="separator"
          tabIndex={0}
          aria-labelledby={panel + "-panel"}
          aria-controls={panel + "-panel"}
          aria-orientation={stacked ? "horizontal" : "vertical"}
          aria-valuemin={min}
          aria-valuemax={max}
          aria-valuenow={size}
          aria-valuetext={size + " pixels"}
          onPointerDown={(event) => {
            if (!event.isPrimary || event.button !== 0) return;
            event.currentTarget.setPointerCapture(event.pointerId);
            drag.current = {
              coordinate: stacked ? event.clientY : event.clientX,
              size,
            };
            event.preventDefault();
            event.currentTarget.focus();
          }}
          onPointerMove={(event) => {
            if (
              !drag.current ||
              !event.currentTarget.hasPointerCapture(event.pointerId)
            )
              return;
            const coordinate = stacked ? event.clientY : event.clientX;
            resize(
              drag.current.size +
                direction * (coordinate - drag.current.coordinate),
            );
          }}
          onLostPointerCapture={() => {
            drag.current = null;
          }}
          onKeyDown={(event) => {
            const less = stacked ? "ArrowUp" : "ArrowLeft";
            const more = stacked ? "ArrowDown" : "ArrowRight";
            if (event.key === less) resize(size - direction * 16);
            else if (event.key === more) resize(size + direction * 16);
            else if (event.key === "Home") resize(min);
            else if (event.key === "End") resize(max);
            else return;
            event.preventDefault();
          }}
        />
      )}
    </div>
  );
}
