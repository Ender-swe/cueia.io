import { Button } from "@cueia/ui";
import Pause from "@react-spectrum/s2/icons/Pause";
import Play from "@react-spectrum/s2/icons/Play";
import { useState, type CSSProperties } from "react";
import paperOrbit from "../../../assets/paper-orbit.txt?raw";

function randomTurnAngle() {
  return Math.random() < 0.5 ? "1turn" : "-1turn";
}

export function PaperOrbit() {
  const [paused, setPaused] = useState(false);
  const [turnAngle, setTurnAngle] = useState(randomTurnAngle);
  const style: CSSProperties & { "--turn-angle": string } = {
    "--turn-angle": turnAngle,
  };

  return (
    <div className="col-2 row-span-full @container grid place-items-center overflow-clip bg-surface [border-inline-start:1px_solid_var(--color-line)] compact:hidden">
      <pre
        aria-hidden="true"
        className={`[grid-area:1/1] m-0 text-apricot font-mono text-[min(1.7cqw,1rem)] leading-[1.05] font-normal [font-variant-ligatures:none] select-none animated-art:[background-image:radial-gradient(_ellipse_24%_23%,_var(--color-ink),_var(--color-apricot)_48%,_color-mix(in_srgb,_var(--color-apricot)_64%,_var(--color-surface))_)] animated-art:[background-size:200%_200%] animated-art:[background-clip:text] animated-art:text-transparent animated-art:[animation:paper-orbit-light_8s_linear_infinite,_paper-orbit-turn_16s_ease-in-out_infinite] ${paused ? "[animation-play-state:paused]!" : ""}`}
        style={style}
        onAnimationIteration={(event) => {
          if (event.animationName === "paper-orbit-turn")
            setTurnAngle(randomTurnAngle());
        }}
      >
        {paperOrbit.trimEnd()}
      </pre>
      <Button
        className="hidden [grid-area:1/1] place-self-end m-3 focus-visible:[outline:2px_solid_var(--color-apricot)] focus-visible:outline-offset-2 animated-art:inline-flex"
        aria-label={
          paused ? "Play artwork animation" : "Pause artwork animation"
        }
        onPress={() => setPaused((value) => !value)}
      >
        {paused ? <Play /> : <Pause />}
      </Button>
    </div>
  );
}
