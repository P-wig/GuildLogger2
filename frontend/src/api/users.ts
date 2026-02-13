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

export type LoginResponse = {
  message: string;
  ok: boolean;
  user: User;
};

export type RegisterResponse = {
  user: User;
  message: string;
};

export const usersApi = {
  // POST /api/auth/login
  login: (credentials: LoginRequest) =>
    api.post<LoginResponse>("/auth/login", credentials),
  // POST /api/auth/register
  register: (userData: RegisterRequest) =>
    api.post<RegisterResponse>("/auth/register", userData),
};
