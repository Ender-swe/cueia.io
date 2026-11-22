import { useState } from "react";
import File from "@react-spectrum/s2/icons/FileText";
import { AttachmentPreview } from "@cueia/components";
import { samplePdf } from "../content";
import { SearchBox } from "../../components/search-box";

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
