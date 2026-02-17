import { api } from "./http";

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

export const projectsApi = {
  list: (params?: { ownerUserId?: string; assignedUser?: string }) =>
    api.get<Project[]>("/projects", { params }),

  get: (projectId: string) => api.get<Project>(`/projects/${projectId}`),

  create: (data: CreateProjectRequest) => api.post<Project>("/projects", data),

  update: (projectId: string, data: UpdateProjectRequest) =>
    api.patch<Project>(`/projects/${projectId}`, data),

  join: (projectId: string, userId: string) =>
    api.post<Project>(`/projects/${projectId}/join`, { userId }),

  leave: (projectId: string, userId: string) =>
    api.post<Project>(`/projects/${projectId}/leave`, { userId }),

  delete: (projectId: string, userId: string) =>
    api.delete(`/projects/${projectId}`, { params: { userId } }),
};
