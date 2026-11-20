import { Heading, Modal, ModalContainer, ModalContent } from "@cueia/ui";

export interface AttachmentPreviewProps {
  attachment: { filename: string; signedUrl: string } | null;
  onClose: () => void;
}

export function AttachmentPreview({
  attachment,
  onClose,
}: AttachmentPreviewProps) {
  return (
    <ModalContainer onDismiss={onClose}>
      {attachment && (
        <Modal size="XL" isDismissible>
          <Heading>{attachment.filename}</Heading>
          <ModalContent>
            <iframe
              title={attachment.filename}
              src={attachment.signedUrl}
              referrerPolicy="no-referrer"
              style={{ width: "100%", height: "65vh", border: 0 }}
            />
          </ModalContent>
        </Modal>
      )}
    </ModalContainer>
  );
}
