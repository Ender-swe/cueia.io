import { useState } from "react"
import { Chat } from "./pages/chat"
import { sample } from "./sample-content"

export function SampleWorkspace({ canSignIn }: { canSignIn: boolean }) {
  const [data, setData] = useState(sample)
  return <Chat data={data} sampleMode canSignIn={canSignIn} onPreference={async timeFormat => {
    setData(current => ({ ...current, preferences: { ...current.preferences, timeFormat } }))
  }} />
}
