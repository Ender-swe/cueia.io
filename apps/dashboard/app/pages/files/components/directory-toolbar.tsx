export function DirectoryToolbar({
  directory,
  openDirectory,
  onRefresh,
}: {
  directory: string;
  openDirectory: (path: string) => void;
  onRefresh: () => void;
}) {
  return (
    <div className="flex flex-wrap gap-1.5 items-center my-2 text-[13px] [&_button]:py-1 [&_button]:px-2 [&_button]:bg-raised [&_button]:[border:1px_solid_var(--color-line)] [&_button]:rounded-md">
      {directory !== "." && (
        <button
          onClick={() =>
            openDirectory(directory.split("/").slice(0, -1).join("/") || ".")
          }
        >
          Up
        </button>
      )}
      <span className="flex-1">
        {directory === "." ? "Workspace" : directory}
      </span>
      <button onClick={onRefresh}>Refresh</button>
    </div>
  );
}
