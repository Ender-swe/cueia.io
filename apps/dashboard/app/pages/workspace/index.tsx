import { useAuth, useClerk } from "@clerk/react";
import { useEffect, useState } from "react";
import { api, type Schema } from "../../api";
import { Chat } from "../chat";
import { WorkspaceStatus } from "./components/workspace-status";

export function WorkspacePage() {
  const { isSignedIn, getToken, sessionId } = useAuth();
  const { signOut } = useClerk();
  const [workspace, setWorkspace] = useState<{
    sessionId: string;
    token: string;
    data: Schema["Workspace"];
  } | null>(null);
  const [error, setError] = useState(false);
  const [attempt, retry] = useState(0);

  useEffect(() => {
    if (!isSignedIn || !sessionId) return;
    const controller = new AbortController();
    const { signal } = controller;
    setWorkspace(null);
    setError(false);
    async function open() {
      const clerk = await getToken();
      if (!clerk) throw new Error("Sign in again");
      const session = await api<Schema["Session"]>(
        "/session",
        clerk,
        "POST",
        undefined,
        signal,
      );
      const data = await api<Schema["Workspace"]>(
        "/workspace",
        session.token,
        "GET",
        undefined,
        signal,
      );
      if (!signal.aborted)
        setWorkspace({ sessionId: sessionId!, token: session.token, data });
    }
    open().catch(() => {
      if (!signal.aborted) setError(true);
    });
    return () => controller.abort();
  }, [sessionId, isSignedIn, getToken, attempt]);

  if (error) return <WorkspaceStatus onRetry={() => retry((n) => n + 1)} />;
  if (!workspace || workspace.sessionId !== sessionId)
    return <WorkspaceStatus />;
  return (
    <Chat
      key={workspace.sessionId}
      data={workspace.data}
      token={workspace.token}
      onPreference={async (timeFormat) => {
        const preferences = await api<Schema["Preferences"]>(
          "/workspace/preferences",
          workspace.token,
          "PATCH",
          { timeFormat },
        );
        setWorkspace({
          ...workspace,
          data: { ...workspace.data, preferences },
        });
      }}
      onSignOut={async () => {
        await api("/session", workspace.token, "DELETE");
        setWorkspace(null);
        await signOut({ redirectUrl: "/sign-in" });
      }}
    />
  );
}
