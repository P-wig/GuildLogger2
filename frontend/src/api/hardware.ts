import { api } from "./http";

// ── Types matching backend schemas ──────────────────────────────

export type Hardware = {
  _id: string;
  hardwareName: string;
  available: number;
  capacity: number;
  assignedProjects: string[];
};

export type CreateHardwareRequest = {
  hardwareName: string;
  capacity: number;
};

export type UpdateHardwareRequest = {
  hardwareName?: string;
  capacity?: number;
};

// ── API functions ───────────────────────────────────────────────

export const hardwareApi = {
  /** List all hardware sets. Optionally filter by assigned project. */
  list: (params?: { assignedProject?: string }) =>
    api.get<Hardware[]>("/hardware", { params }),

  /** Get a single hardware set by its Mongo _id. */
  get: (id: string) => api.get<Hardware>(`/hardware/${id}`),

  /** Create a new hardware set. */
  create: (data: CreateHardwareRequest) =>
    api.post<Hardware>("/hardware", data),

  /** Partially update a hardware set by Mongo _id. */
  update: (id: string, data: UpdateHardwareRequest) =>
    api.patch<Hardware>(`/hardware/${id}`, data),

  /** Delete a hardware set by Mongo _id. */
  delete: (id: string) => api.delete(`/hardware/${id}`),
};
