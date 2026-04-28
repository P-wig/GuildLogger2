/**
 * Cross-tab authentication synchronization helpers.
 *
 * Why this exists:
 * - A user can log in or out in one tab while other tabs are still open.
 * - We need all tabs to refresh auth state immediately.
 *
 * How it works:
 * - `authChannel` uses the browser BroadcastChannel API on the "auth" channel.
 * - `broadcastAuthChanged()` emits an "AUTH_CHANGED" message to listening tabs.
 * - It also writes `__auth_changed_at__` in localStorage so tabs listening to
 *   the `storage` event are notified as a fallback/secondary signal.
 */

const CHANNEL = "auth";

export const authChannel =
  "BroadcastChannel" in window ? new BroadcastChannel(CHANNEL) : null;

export function broadcastAuthChanged() {
  if (authChannel) authChannel.postMessage({ type: "AUTH_CHANGED" });
  // storage event fires in other tabs when localStorage is mutated
  localStorage.setItem("__auth_changed_at__", String(Date.now()));
}
