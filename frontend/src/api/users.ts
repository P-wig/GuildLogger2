import { api } from './http';

export type User = {
  _id: string;
  userId: string;
  email?: string;
};

export type LoginRequest = {
  userId: string;
  password: string;
};

export type LoginResponse = {
  user: User;
  token?: string;
  message: string;
};

export const usersApi = {
  // POST /api/auth/login
  login: (credentials: LoginRequest) => 
    api.post<LoginResponse>('/auth/login', credentials),
};