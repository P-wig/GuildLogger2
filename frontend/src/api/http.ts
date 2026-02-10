import axios from 'axios';
import type { AxiosError, AxiosResponse } from 'axios';

// Create axios instance with base configuration
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
    // Handle request errors
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor - runs after every response
api.interceptors.response.use(
  (response: AxiosResponse) => {
    // Log successful responses in development
    if (import.meta.env.DEV) {
      console.log(`API Response: ${response.config.method?.toUpperCase()} ${response.config.url}`, response.data);
    }
    
    return response;
  },
  (error: AxiosError) => {
    // Handle response errors
    if (import.meta.env.DEV) {
      console.error(`API Error: ${error.config?.method?.toUpperCase()} ${error.config?.url}`, error.response?.data);
    }
    
    // Handle specific error cases - check if response exists first
    if (error.response) {
      const status = error.response.status;
      
      if (status === 401) {
        // Unauthorized - clear token and redirect to login
        localStorage.removeItem('authToken');
        // You could dispatch a logout action here
        // window.location.href = '/auth';
      }
      
      if (status === 403) {
        // Forbidden - user doesn't have permission
        console.error('Access forbidden');
      }
      
      if (status >= 500) {
        // Server error
        console.error('Server error occurred');
      }
    } else {
      // Network error or request timeout
      console.error('Network error or timeout occurred');
    }
    
    return Promise.reject(error);
  }
);

// Helper functions for common request types
export const httpClient = {
  get: <T>(url: string, config = {}) => api.get<T>(url, config),
  post: <T>(url: string, data?: any, config = {}) => api.post<T>(url, data, config),
  put: <T>(url: string, data?: any, config = {}) => api.put<T>(url, data, config),
  patch: <T>(url: string, data?: any, config = {}) => api.patch<T>(url, data, config),
  delete: <T>(url: string, config = {}) => api.delete<T>(url, config),
};

// Export types for error handling
export type ApiError = {
  message: string;
  status?: number;
  data?: any;
};

// Helper function to extract error message
export const getErrorMessage = (error: AxiosError): string => {
  if (error.response?.data && typeof error.response.data === 'object') {
    const data = error.response.data as any;
    return data.message || data.error || 'An error occurred';
  }
  
  return error.message || 'Network error occurred';
};

export default api;