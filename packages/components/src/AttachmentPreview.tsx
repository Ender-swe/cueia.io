import { Content, Dialog, DialogContainer, Heading } from "@react-spectrum/s2"

export interface AttachmentPreviewProps {
  attachment: { filename: string; signedUrl: string } | null
  onClose: () => void
}

export function AttachmentPreview({ attachment, onClose }: AttachmentPreviewProps) {
  return (
    <DialogContainer onDismiss={onClose}>
      {attachment && (
        <Dialog size="XL" isDismissible>
          <Heading>{attachment.filename}</Heading>
          <Content>
            <iframe
              title={attachment.filename}
              src={attachment.signedUrl}
              referrerPolicy="no-referrer"
              style={{ width: "100%", height: "65vh", border: 0 }}
            />
          </Content>
        </Dialog>
      )}
    </DialogContainer>
  )
}
