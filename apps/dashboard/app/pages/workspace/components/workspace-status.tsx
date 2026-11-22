export function WorkspaceStatus({ onRetry }: { onRetry?: () => void }) {
  return (
    <main
      className="font-sans text-ink leading-normal [&_button]:py-1 [&_button]:px-2 [&_button]:bg-raised [&_button]:[border:1px_solid_var(--color-line)] [&_button]:rounded-md [&_a]:text-apricot min-h-dvh grid place-content-center gap-3 p-6"
      role={onRetry ? "alert" : "status"}
    >
      {onRetry ? (
        <>
          Could not open your workspace.{" "}
          <button onClick={onRetry}>Retry</button>
        </>
      ) : (
        "Opening your workspace…"
      )}
    </main>
  );
}
