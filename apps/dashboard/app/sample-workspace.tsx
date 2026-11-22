import { useState } from "react";
import { useClerk } from "@clerk/react";
import { Chat } from "./pages/chat";
import { sample } from "./sample-content";

export function SampleWorkspace() {
  const [data, setData] = useState(sample);
  const { signOut } = useClerk();
  return (
    <Chat
      data={data}
      sampleMode
      onSignOut={() => signOut({ redirectUrl: "/sign-in" })}
      onPreference={async (timeFormat) => {
        setData((current) => ({
          ...current,
          preferences: { ...current.preferences, timeFormat },
        }));
      }}
    />
  );
}
