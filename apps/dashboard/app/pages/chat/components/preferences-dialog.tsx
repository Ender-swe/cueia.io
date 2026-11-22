import { Link } from "react-router";
import {
  Button,
  Heading,
  Modal,
  ModalContainer,
  ModalContent,
} from "@cueia/ui";
export function PreferencesDialog({
  menu,
  onDismiss,
  timeFormat,
  saving,
  save,
  onPreference,
  onSignOut,
  sampleMode,
  error,
}: {
  menu: boolean;
  onDismiss: () => void;
  timeFormat: string;
  saving: boolean;
  save: (action: () => Promise<void>, message: string) => Promise<void>;
  onPreference: (format: string) => Promise<void>;
  onSignOut: () => Promise<void>;
  sampleMode: boolean;
  error: string;
}) {
  return (
    <ModalContainer onDismiss={onDismiss}>
      {menu && (
        <Modal isDismissible size="S">
          <Heading>Preferences</Heading>
          <ModalContent>
            <div className="font-sans text-ink leading-normal grid justify-items-start gap-4 [&_a]:text-apricot">
              <label className="grid gap-2 text-[14px]">
                Time format
                <select
                  className="w-full bg-raised p-2 [border:1px_solid_#53514b] rounded-md"
                  disabled={saving}
                  value={timeFormat}
                  onChange={(e) => {
                    const value = e.target.value;
                    void save(
                      () => onPreference(value),
                      "Could not save preference.",
                    );
                  }}
                >
                  <option value="12h">12 hour</option>
                  <option value="24h">24 hour</option>
                </select>
              </label>
              <Button
                isDisabled={saving}
                onPress={() =>
                  void save(onSignOut, "Could not sign out. Please retry.")
                }
              >
                Sign out
              </Button>
              {sampleMode && (
                <Link to="/workspace">Open private workspace</Link>
              )}
              {error && <p role="alert">{error}</p>}
            </div>
          </ModalContent>
        </Modal>
      )}
    </ModalContainer>
  );
}
