import { api } from "./http";

export type User = {
  _id: string;
  userId: string;
};

export type LoginRequest = {
  userId: string;
  password: string;
};

export type RegisterRequest = {
  userId: string;
  password: string;
};

export type AuthResponse = {
  user: User;
  message: string;
};

export const authApi = {
  login: (credentials: LoginRequest) =>
    api.post<AuthResponse>("/auth/login", credentials),

  register: (userData: RegisterRequest) =>
    api.post<AuthResponse>("/auth/register", userData),
};