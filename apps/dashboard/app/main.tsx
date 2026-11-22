import { createRoot } from "react-dom/client";
import {
  ClerkFailed,
  ClerkLoading,
  ClerkProvider,
  useAuth,
} from "@clerk/react";
import { Provider } from "@react-spectrum/s2/Provider";
import {
  BrowserRouter,
  Link,
  Navigate,
  Outlet,
  Route,
  Routes,
  useHref,
  useMatch,
  useNavigate,
} from "react-router";
import { AuthPage, authLocalization } from "./pages/auth";
import { WorkspacePage } from "./pages/workspace";
import { SampleWorkspace } from "./sample-workspace";
import "@fontsource-variable/dm-sans";
import "./global.css";

function RequireSignIn() {
  const atLaunch = useMatch("/");
  const { isLoaded, isSignedIn, sessionId } = useAuth();
  if (!isLoaded)
    return (
      <main className="font-sans text-ink leading-normal [&_button]:py-1 [&_button]:px-2 [&_button]:bg-raised [&_button]:[border:1px_solid_var(--color-line)] [&_button]:rounded-md [&_a]:text-apricot min-h-dvh grid place-content-center gap-3 p-6">
        <ClerkLoading>
          <p role="status">Checking sign-in…</p>
        </ClerkLoading>
        <ClerkFailed>
          <p role="alert">
            Sign-in couldn’t load. Refresh the page to try again.
          </p>
        </ClerkFailed>
      </main>
    );
  if (!isSignedIn || !sessionId) {
    if (atLaunch)
      return (
        <main className="min-h-dvh grid place-content-center bg-canvas p-6 font-sans text-interface">
          <Link className="text-apricot" to="/workspace">
            Open sample workspace
          </Link>
        </main>
      );
    const search = new URLSearchParams({
      redirect_url: window.location.href,
    }).toString();
    return <Navigate replace to={{ pathname: "/sign-in", search }} />;
  }
  return <Outlet key={sessionId} />;
}

function Dashboard() {
  const navigate = useNavigate();
  const key = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;
  return (
    <Provider colorScheme="dark" router={{ navigate, useHref }}>
      {key ? (
        <ClerkProvider
          publishableKey={key}
          routerPush={(to) => navigate(to)}
          routerReplace={(to) => navigate(to, { replace: true })}
          signInUrl="/sign-in"
          signUpUrl="/sign-up"
          signInFallbackRedirectUrl="/"
          signUpFallbackRedirectUrl="/"
          localization={authLocalization}
        >
          <Routes>
            <Route path="/sign-in/*" element={<AuthPage mode="sign-in" />} />
            <Route path="/sign-up/*" element={<AuthPage mode="sign-up" />} />
            <Route element={<RequireSignIn />}>
              <Route path="/" element={<SampleWorkspace />} />
              <Route path="/workspace" element={<WorkspacePage />} />
            </Route>
          </Routes>
        </ClerkProvider>
      ) : (
        <main
          className="font-sans text-ink leading-normal [&_button]:py-1 [&_button]:px-2 [&_button]:bg-raised [&_button]:[border:1px_solid_var(--color-line)] [&_button]:rounded-md [&_a]:text-apricot min-h-dvh grid place-content-center gap-3 p-6"
          role="alert"
        >
          Set VITE_CLERK_PUBLISHABLE_KEY to enable sign-in.
        </main>
      )}
    </Provider>
  );
}

createRoot(document.getElementById("root")!).render(
  <BrowserRouter>
    <Dashboard />
  </BrowserRouter>,
);
