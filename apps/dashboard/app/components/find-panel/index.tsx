import { useState } from "react";
import type { Schema } from "../../api";
import { SearchBox } from "../search-box";

export function FindPanel({
  data,
  onSelect,
}: {
  data: Schema["Workspace"];
  onSelect: (id: string) => void;
}) {
  const [query, setQuery] = useState("");
  const term = query.trim().toLowerCase();
  const entries = [
    ...data.threads.map((t) => ({
      id: t.id,
      type: "Email",
      title: t.subject,
      text: t.messages.map((m) => m.senderAddress + " " + m.bodyText).join(" "),
    })),
    ...data.events.map((e) => ({
      id: e.id,
      type: "Calendar",
      title: e.title,
      text: "",
    })),
  ].filter((item) =>
    (item.title + " " + item.text).toLowerCase().includes(term),
  );
  return (
    <>
      <SearchBox label="Find in workspace" value={query} onChange={setQuery} />
      <h2 className="py-1 px-2 text-[16px] font-semibold">Find</h2>
      {entries.map((entry) => (
        <button
          className="flex items-start gap-0.75 w-full min-h-9.5 py-1.75 px-2.75 bg-transparent border-0 rounded-md text-left wrap-anywhere hover:bg-raised flex-col text-[14px]"
          key={entry.id}
          onClick={() => onSelect(entry.id)}
        >
          <small className="text-[13px] text-muted">{entry.type}</small>
          {entry.title}
        </button>
      ))}
      {!entries.length && (
        <p className="text-[13px] text-muted">No matching records.</p>
      )}
    </>
  );
}
