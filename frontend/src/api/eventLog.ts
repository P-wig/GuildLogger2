import { api } from "./http";
import type { ActiveEvent } from "./guilds";

export type EventLogMember = {
  discordId: string;
  username: string;
  avatarHash: string;
};

export type ValidateTokenResponse =
  | {
      ok: true;
      event: ActiveEvent;
      preSelectedIds: string[];
      members: EventLogMember[];
    }
  | {
      ok: false;
      reason: "expired" | "not_found" | "already_submitted" | "missing_token";
    };

export const eventLogApi = {
  validate: (token: string) =>
    api.get<ValidateTokenResponse>("/event-log/validate", {
      params: { token },
    }),

  submit: (payload: {
    token: string;
    summary: string;
    participantIds: string[];
  }) => api.post<{ ok: boolean }>("/event-log/submit", payload),
};
