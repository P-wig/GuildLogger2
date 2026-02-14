import { api } from "./http";

// ── Types matching backend schemas ──────────────────────────────

export type Project = {
  _id: string;
  projectId: string;
  projectName: string;
  description: string;
  ownerUserId: string;
  assignedUsers: string[];
  assignedHardware: { hardwareId: string; amount: number }[];
};

export type CreateProjectRequest = {
  projectId: string;
  projectName: string;
  description: string;
  ownerUserId: string;
};

export type UpdateProjectRequest = {
  projectName?: string;
  description?: string;
};

// ── API functions ───────────────────────────────────────────────

export const projectsApi = {
  /** List projects. Optionally filter by owner or assigned user. */
  list: (params?: { ownerUserId?: string; assignedUser?: string }) =>
    api.get<Project[]>("/projects", { params }),

  /** Get a single project by its Mongo _id. */
  get: (id: string) => api.get<Project>(`/projects/${id}`),

  /** Create a new project. */
  create: (data: CreateProjectRequest) => api.post<Project>("/projects", data),

  /** Partially update a project by Mongo _id. */
  update: (id: string, data: UpdateProjectRequest) =>
    api.patch<Project>(`/projects/${id}`, data),

  /** Join a project. */
  join: (id: string, userId: string) =>
    api.post<Project>(`/projects/${id}/join`, { userId }),

  /** Leave a project. */
  leave: (id: string, userId: string) =>
    api.post<{ ok: boolean }>(`/projects/${id}/leave`, { userId }),

  /** Delete a project by Mongo _id. */
  delete: (id: string) => api.delete(`/projects/${id}`),
};
