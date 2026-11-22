import type { Schema } from "../../../api";
export function FileResults({
  entries,
  error,
  searching,
  openDirectory,
  preview,
}: {
  entries: Schema["Entry"][] | null;
  error: string;
  searching: boolean;
  openDirectory: (path: string) => void;
  preview: (entry: Schema["Entry"], button: HTMLButtonElement) => void;
}) {
  return (
    <>
      {error && <p role="alert">{error}</p>}
      {entries === null && !error && <p role="status">Loading files…</p>}
      {entries?.length === 0 && (
        <p>{searching ? "No matching files." : "This directory is empty."}</p>
      )}
      <ul className="[list-style:none] p-0 text-[14px] [&_button]:max-w-full [&_button]:py-1 [&_button]:px-0 [&_button]:text-left [&_button]:bg-transparent [&_button]:border-0 [&_button]:wrap-anywhere [&_button:hover]:underline">
        {entries?.map((entry) => (
          <li
            className="[border-top:1px_solid_var(--color-line)] py-2 wrap-anywhere"
            key={entry.path}
          >
            {entry.kind === "directory" ? (
              <button onClick={() => openDirectory(entry.path)}>
                {entry.name}/
              </button>
            ) : /\.pdf$/i.test(entry.name) ? (
              <button onClick={(e) => void preview(entry, e.currentTarget)}>
                {entry.name}
              </button>
            ) : (
              <span>{entry.name}</span>
            )}
            {searching && (
              <small className="block text-muted">{entry.path}</small>
            )}
          </li>
        ))}
      </ul>
    </>
  );
}
