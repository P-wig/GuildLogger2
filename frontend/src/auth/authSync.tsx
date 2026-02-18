// authSync.ts
const CHANNEL = "auth";

export const authChannel =
  "BroadcastChannel" in window ? new BroadcastChannel(CHANNEL) : null;

export function broadcastAuthChanged() {
  if (authChannel) authChannel.postMessage({ type: "AUTH_CHANGED" });
  // storage event fires in other tabs when localStorage is mutated
  localStorage.setItem("auth.lastChangedAt", String(Date.now()));
}
