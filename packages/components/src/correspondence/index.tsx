import Attach from "@react-spectrum/s2/icons/Attach";

interface Message {
  id: string;
  senderAddress: string;
  bodyText: string;
  sentAt: string;
  attachments: { fileId: string; filename: string }[];
}

export function Correspondence({
  subject,
  messages,
}: {
  subject: string;
  messages: Message[];
}) {
  return (
    <article className="bg-ash rounded-[3px_10px_10px] p-4.5 [&_h2]:font-semibold [&_h2]:text-[18px] [&_h2]:leading-[1.35]">
      <h2>{subject}</h2>
      {messages.map((message) => (
        <section key={message.id} className="pt-3.5">
          <div className="flex flex-wrap gap-[4px_10px] text-[13px] text-muted">
            <strong className="text-ink">{message.senderAddress}</strong>
            <span>
              {new Intl.DateTimeFormat("en-US", { dateStyle: "medium" }).format(
                new Date(message.sentAt),
              )}
            </span>
          </div>
          <p className="mt-2 whitespace-pre-wrap">{message.bodyText}</p>
          {message.attachments.map((file) => (
            <div
              key={file.fileId}
              className="flex items-center gap-1.5 text-apricot text-[13px] my-[10px_0] wrap-anywhere"
            >
              <Attach />
              {file.filename}
            </div>
          ))}
        </section>
      ))}
    </article>
  );
}
