import axios from 'axios';
import type { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { notifySessionExpired } from '../auth/authSync';

// Paths where a 401 means "this one-time link is bad", not "your login died".
// The request interceptor attaches the session JWT to every request, so without this
// a logged-in user opening a stale event-log link would be signed out of the whole app.
const NON_SESSION_401_PATHS = ['/event-log/'];

export const api = axios.create({
  baseURL: '/api', // Uses Vite proxy from vite.config.ts to route to backend
  timeout: 10000, // 10 second timeout
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - runs before every request
api.interceptors.request.use(
  (config) => {
    // Add authentication token if available
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // Log requests in development
    if (import.meta.env.DEV) {
      console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`, config.data);
    }

    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor - runs after every response
api.interceptors.response.use(
  (response: AxiosResponse) => {
    if (import.meta.env.DEV) {
      console.log(`API Response: ${response.config.method?.toUpperCase()} ${response.config.url}`, response.data);
    }

    return response;
  },
  (error: AxiosError) => {
    if (import.meta.env.DEV) {
      console.error(`API Error: ${error.config?.method?.toUpperCase()} ${error.config?.url}`, error.response?.data);
    }

    if (error.response) {
      const status = error.response.status;

      if (status === 401) {
        const url = error.config?.url ?? '';
        const isSessionAuth = !NON_SESSION_401_PATHS.some((p) => url.startsWith(p));

        if (isSessionAuth) {
          // Clear both keys — triggers AuthProvider storage listener which sets user to null
          localStorage.removeItem('authToken');
          localStorage.removeItem('session_token');

          // `reauth` marks a dead Discord authorization, which needs a different
          // explanation than an ordinary expired login.
          const body = error.response.data as { error?: string; reauth?: boolean } | undefined;
          notifySessionExpired(
            body?.reauth && body.error ? body.error : undefined,
          );
        }
      }

      if (status === 403) {
        console.error('Access forbidden');
      }

      if (status >= 500) {
        console.error('Server error occurred');
      }
    } else {
      console.error('Network error or timeout occurred');
    }

    return Promise.reject(error);
  }
);

// Typed helper functions — B is the request body type, T is the response type
export const httpClient = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    api.get<T>(url, config),
  post: <T, B = unknown>(url: string, data?: B, config?: AxiosRequestConfig) =>
    api.post<T>(url, data, config),
  put: <T, B = unknown>(url: string, data?: B, config?: AxiosRequestConfig) =>
    api.put<T>(url, data, config),
  patch: <T, B = unknown>(url: string, data?: B, config?: AxiosRequestConfig) =>
    api.patch<T>(url, data, config),
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    api.delete<T>(url, config),
};

export type ApiError = {
  message: string;
  status?: number;
  data?: unknown;
};

export const getErrorMessage = (error: AxiosError): string => {
  if (error.response?.data && typeof error.response.data === 'object') {
    const data = error.response.data as Record<string, unknown>;
    const msg = data.message ?? data.error;
    if (typeof msg === 'string') return msg;
  }

  return error.message || 'Network error occurred';
};

export default api;