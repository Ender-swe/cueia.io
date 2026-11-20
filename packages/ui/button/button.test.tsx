import { act } from "react"
import { createRoot } from "react-dom/client"
import { expect, it, vi } from "vitest"
import { Button } from "./button"

it("activates a named action and respects its disabled state", async () => {
  const host = document.body.appendChild(document.createElement("div"))
  const root = createRoot(host)
  const onPress = vi.fn()
  try {
    await act(() => root.render(<Button onPress={onPress}>Menu</Button>))
    const button = host.querySelector("button")!
    expect(button.textContent).toContain("Menu")
    await act(() => button.click())
    await act(() => root.render(<Button isDisabled onPress={onPress}>Menu</Button>))
    expect(button.disabled).toBe(true)
    await act(() => button.click())
    expect(onPress).toHaveBeenCalledTimes(1)
  } finally {
    await act(() => root.unmount())
    host.remove()
  }
})
