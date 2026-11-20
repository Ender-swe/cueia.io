import { ClerkFailed, SignIn, SignUp, useAuth } from "@clerk/react";
import { CueiaMark } from "@cueia/components";
import { Button } from "@cueia/ui";
import Pause from "@react-spectrum/s2/icons/Pause";
import Play from "@react-spectrum/s2/icons/Play";
import { useState, type ComponentProps, type CSSProperties } from "react";
import { Navigate, useMatch, useSearchParams } from "react-router";
import paperOrbit from "../assets/paper-orbit.txt?raw";
import "../auth.css";

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
    rootBox: "auth-widget",
    cardBox: "auth-card-box",
    card: "auth-card",
    header: "auth-heading",
    headerTitle: "auth-title",
    socialButtonsBlockButton: "auth-social-button",
    formFieldInput: "auth-input",
    formButtonPrimary: "auth-primary",
    footer: "auth-footer",
  },
};

function workspaceDestination(value: string | null) {
  if (!value) return null;
  try {
    const url = new URL(value, window.location.origin);
    if (url.origin === window.location.origin && ["/", "/workspace", "/workspace/"].includes(url.pathname)) {
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
    animationPlayState: paused ? "paused" : "running",
  };

  return (
    <div className="auth-art">
      <pre
        aria-hidden="true"
        style={style}
        onAnimationIteration={event => {
          if (event.animationName === "paper-orbit-turn") setTurnAngle(randomTurnAngle());
        }}
      >{paperOrbit.trimEnd()}</pre>
      <Button
        className="auth-art-control"
        aria-label={paused ? "Play artwork animation" : "Pause artwork animation"}
        onPress={() => setPaused(value => !value)}
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
  if (isSignedIn && atStart) {
    const to = destination ? destination.pathname + destination.search + destination.hash : "/";
    return <Navigate replace to={to} />;
  }
  const returnQuery = destination ? "?" + new URLSearchParams({ redirect_url: destination.href }) : "";
  const fallback = (
    <p className="auth-status" role="status">
      Loading {signingIn ? "sign-in" : "sign-up"}…
    </p>
  );

  return (
    <main className="auth-page">
      <title>
        {signingIn ? "Sign in to Cueia" : "Create your Cueia account"}
      </title>
      <header className="auth-brand">
        <span className="auth-wordmark"><CueiaMark />Cueia</span>
      </header>
      <section
        className="auth-form"
        aria-label={signingIn ? "Sign in" : "Sign up"}
      >
        {signingIn ? (
          <SignIn
            path="/sign-in"
            signUpUrl={"/sign-up" + returnQuery}
            routing="path"
            appearance={appearance}
            fallback={fallback}
          />
        ) : (
          <SignUp
            path="/sign-up"
            signInUrl={"/sign-in" + returnQuery}
            routing="path"
            appearance={appearance}
            fallback={fallback}
          />
        )}
        <ClerkFailed>
          <p className="auth-status" role="alert">
            Sign-in couldn’t load. Refresh the page to try again.
          </p>
        </ClerkFailed>
      </section>
      <PaperOrbit />
    </main>
  );
}
