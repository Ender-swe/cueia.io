import { act } from "react";
import { createRoot } from "react-dom/client";
import { expect, it } from "vitest";
import { Input } from "./input";

it("associates the search field with its label and renders controlled updates", async () => {
  const host = document.body.appendChild(document.createElement("div"));
  const root = createRoot(host);
  try {
    for (const value of ["Main", "Side"]) {
      await act(() =>
        root.render(
          <Input
            type="search"
            label="Search chats"
            value={value}
            onChange={() => {}}
          />,
        ),
      );
      const input = host.querySelector("input")!;
      expect(input.labels?.[0].textContent).toBe("Search chats");
      expect(input.value).toBe(value);
    }
  } finally {
    await act(() => root.unmount());
    host.remove();
  }
});
