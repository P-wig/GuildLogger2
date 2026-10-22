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

export const SESSION_EXPIRED_EVENT = "auth:session-expired";

export const DEFAULT_SESSION_EXPIRED_MESSAGE =
  "Your session expired. Please sign in again.";

export type SessionExpiredDetail = { message: string };

/**
 * Signals that the backend rejected our credentials and the session is now dead.
 *
 * This is a DOM event rather than a BroadcastChannel message because neither
 * BroadcastChannel nor the `storage` event fires in the context that triggered it —
 * so the tab that actually received the 401 would be the one tab never told about it,
 * leaving the user staring at a blank protected page. Cross-tab fan-out is handled
 * separately by the provider via broadcastAuthChanged().
 */
export function notifySessionExpired(
  message: string = DEFAULT_SESSION_EXPIRED_MESSAGE,
) {
  window.dispatchEvent(
    new CustomEvent<SessionExpiredDetail>(SESSION_EXPIRED_EVENT, {
      detail: { message },
    }),
  );
}
