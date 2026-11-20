import { useState } from "react";
import Search from "@react-spectrum/s2/icons/Search";
import File from "@react-spectrum/s2/icons/FileText";
import { AttachmentPreview } from "@cueia/components";
import { Input } from "@cueia/ui/input";
import type { Schema } from "./api";
import { samplePdf } from "./sample-content";

export function SearchBox({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="relative [&>svg]:absolute [&>svg]:z-1 [&>svg]:left-2.75 [&>svg]:top-2.5 [&>svg]:w-4 [&>svg]:h-4 [&>svg]:pointer-events-none [&>svg]:text-muted [&_input]:ps-6.75">
      <Search aria-hidden="true" />
      <Input
        type="search"
        aria-label={label}
        placeholder="Search"
        value={value}
        onChange={onChange}
      />
    </div>
  );
}

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

export function SampleFiles() {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  return (
    <>
      <SearchBox label="Search files" value={query} onChange={setQuery} />
      {samplePdf.filename.includes(query.trim().toLowerCase()) ? (
        <button
          className="flex items-start gap-0.75 w-full min-h-9.5 py-1.75 px-2.75 bg-transparent border-0 rounded-md text-left wrap-anywhere hover:bg-raised flex-row text-[14px] [&_small]:text-[13px] [&_small]:text-muted"
          onClick={() => setOpen(true)}
        >
          <File />
          {samplePdf.filename}
        </button>
      ) : (
        <p className="text-[13px] text-muted">No matching files.</p>
      )}
      <AttachmentPreview
        attachment={open ? samplePdf : null}
        onClose={() => setOpen(false)}
      />
    </>
  );
}
