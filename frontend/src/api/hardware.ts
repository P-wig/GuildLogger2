import { api } from "./http";

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

export type HardwareCheckoutRequest = {
  projectId: string;
  amount: number;
  userId: string;
};

export type HardwareCheckinRequest = {
  projectId: string;
  amount: number;
  userId: string;
};

export const hardwareApi = {
  list: (params?: { assignedProject?: string }) =>
    api.get<Hardware[]>("/hardware", { params }),

  get: (id: string) => api.get<Hardware>(`/hardware/${id}`),

  create: (data: CreateHardwareRequest) =>
    api.post<Hardware>("/hardware", data),

  update: (id: string, data: UpdateHardwareRequest) =>
    api.patch<Hardware>(`/hardware/${id}`, data),

  delete: (id: string) => api.delete(`/hardware/${id}`),

  checkout: (id: string, data: HardwareCheckoutRequest) =>
    api.post<Hardware>(`/hardware/${id}/checkout`, data),

  checkin: (id: string, data: HardwareCheckinRequest) =>
    api.post<Hardware>(`/hardware/${id}/checkin`, data),
};
