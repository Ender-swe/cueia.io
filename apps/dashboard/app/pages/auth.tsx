import { ClerkFailed, SignIn, SignUp, useAuth } from "@clerk/react";
import { CueiaMark } from "@cueia/components";
import { Button } from "@cueia/ui";
import Pause from "@react-spectrum/s2/icons/Pause";
import Play from "@react-spectrum/s2/icons/Play";
import { useState, type ComponentProps, type CSSProperties } from "react";
import { Navigate, useMatch, useSearchParams } from "react-router";
import paperOrbit from "../assets/paper-orbit.txt?raw";

export const authLocalization = {
  signIn: {
    start: {
      title: "Sign in to Cueia",
      subtitle: "Continue to your personal workspace.",
      titleCombined: "Sign in to Cueia",
      subtitleCombined: "Continue to your personal workspace.",
    },
  },
  signUp: {
    start: {
      title: "Create your Cueia account",
      subtitle: "Choose how you’ll sign in to your workspace.",
    },
  },
};

const appearance: ComponentProps<typeof SignIn>["appearance"] = {
  options: { logoPlacement: "none", socialButtonsVariant: "blockButton" },
  captcha: { theme: "dark" },
  variables: {
    colorBackground: "#1b1b1a",
    colorForeground: "#f0ede8",
    colorMuted: "#292927",
    colorMutedForeground: "#aaa7a1",
    colorInput: "#242423",
    colorInputForeground: "#f0ede8",
    colorNeutral: "#f0ede8",
    colorPrimary: "#e9aa83",
    colorPrimaryForeground: "#191918",
    colorRing: "#e9aa83",
    colorDanger: "#f4a398",
    fontFamily: '"DM Sans Variable", system-ui, sans-serif',
    fontSize: "0.9375rem",
    borderRadius: "0.5rem",
  },
  elements: {
    rootBox: "w-full max-w-full",
    cardBox:
      "w-full max-w-full overflow-visible bg-transparent border-0 rounded-none shadow-none",
    card: "bg-transparent border-0 rounded-none shadow-none p-0 gap-7 compact-auth:gap-3",
    header: "items-start text-left",
    headerTitle: "text-[23px] leading-[1.25] tracking-[-0.04em] font-semibold",
    socialButtonsBlockButton:
      "min-h-11 rounded-lg bg-ash text-ink [border:1px_solid_var(--color-line)] shadow-none hover:bg-thread-selected",
    formFieldInput: "min-h-11 rounded-lg",
    formButtonPrimary: "min-h-11 rounded-lg",
    footer:
      "bg-transparent border-0 rounded-none shadow-none p-[24px_0_0] [&_.cl-footerAction]:px-0 [&_.cl-footerActionLink]:whitespace-nowrap compact-auth:pt-3",
  },
};

function workspaceDestination(value: string | null) {
  if (!value) return null;
  try {
    const url = new URL(value, window.location.origin);
    if (
      url.origin === window.location.origin &&
      ["/", "/workspace", "/workspace/"].includes(url.pathname)
    ) {
      return url;
    }
  } catch {
    // Malformed return links use the same fallback as direct visits.
  }
  return null;
}

function randomTurnAngle() {
  return Math.random() < 0.5 ? "1turn" : "-1turn";
}

function PaperOrbit() {
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

export function AuthPage({ mode }: { mode: "sign-in" | "sign-up" }) {
  const signingIn = mode === "sign-in";
  const { isSignedIn } = useAuth();
  const atStart = useMatch("/" + mode);
  const [params] = useSearchParams();
  const destination = workspaceDestination(params.get("redirect_url"));
  const returnTo = destination
    ? destination.pathname + destination.search + destination.hash
    : "/";
  if (isSignedIn && atStart) {
    return <Navigate replace to={returnTo} />;
  }
  const returnQuery = destination
    ? "?" + new URLSearchParams({ redirect_url: destination.href })
    : "";
  const fallback = (
    <p className="m-0 text-muted" role="status">
      Loading {signingIn ? "sign-in" : "sign-up"}…
    </p>
  );

  return (
    <main className="grid grid-cols-[minmax(0,1fr)_min(50%,75svh)] grid-rows-[minmax(3rem,4fr)_auto_minmax(0,1fr)] min-h-svh bg-canvas text-ink font-sans text-interface font-normal compact:grid-cols-[minmax(0,_1fr)]">
      <title>
        {signingIn ? "Sign in to Cueia" : "Create your Cueia account"}
      </title>
      <header className="@container-size grid place-items-center">
        <span className="select-none inline-flex items-center gap-[0.28em] text-[clamp(1.5rem,min(16cqw,68cqh),5.5rem)] leading-none font-semibold tracking-[-0.04em]">
          <CueiaMark className="text-[1.15em] size-[1em]" />
          Cueia
        </span>
      </header>
      <section
        className="row-2 self-center w-[min(368px,calc(100%-64px))] my-2 mx-auto min-w-0 [&_:is(button,a,input):focus-visible]:[outline:2px_solid_var(--color-apricot)] [&_:is(button,a,input):focus-visible]:outline-offset-2 compact:w-[min(368px,_calc(100%_-_48px))] compact-auth:[&_.cl-form]:gap-3 compact-auth:[&_.cl-main]:gap-2 compact-auth:[&_.cl-form_:has(>_.cl-formFieldRow)]:gap-y-2 compact-auth:[&_.cl-formField_:has(>_.cl-formFieldLabelRow)]:gap-y-1 compact-auth:[&_.cl-footerAction]:py-2 compact-auth:[&_.cl-footerItem]:py-2"
        aria-label={signingIn ? "Sign in" : "Sign up"}
      >
        {signingIn ? (
          <SignIn
            path="/sign-in"
            signUpUrl={"/sign-up" + returnQuery}
            forceRedirectUrl={returnTo}
            routing="path"
            appearance={appearance}
            fallback={fallback}
          />
        ) : (
          <SignUp
            path="/sign-up"
            signInUrl={"/sign-in" + returnQuery}
            forceRedirectUrl={returnTo}
            routing="path"
            appearance={appearance}
            fallback={fallback}
          />
        )}
        <ClerkFailed>
          <p className="m-0 text-muted" role="alert">
            Sign-in couldn’t load. Refresh the page to try again.
          </p>
        </ClerkFailed>
      </section>
      <PaperOrbit />
    </main>
  );
}
