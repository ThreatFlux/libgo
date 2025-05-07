import axios, { AxiosError, AxiosRequestConfig, isAxiosError } from 'axios';
import { LoginRequest, LoginResponse } from '@/types/api';

// Create axios instance
const apiClient = axios.create({
  // When using Vite dev server, baseURL can be empty as the proxy will handle it
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: false,
});

// Add token to request header
apiClient.interceptors.request.use((config) => {
  // Always get fresh token from localStorage
  const token = localStorage.getItem('token');
  if (token && config.headers) {
    config.headers['Authorization'] = `Bearer ${token}`;
  }
  return config;
});

// Handle token expiration
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const login = async (data: LoginRequest): Promise<LoginResponse> => {
  const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', data);
  
  // Save token to localStorage
  localStorage.setItem('token', response.data.token);
  
  return response.data;
};

export const logout = (): void => {
  localStorage.removeItem('token');
  window.location.href = '/login';
};

// Generic API functions
export const get = async <T>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  try {
    console.log(`Making GET request to ${url}`);
    const response = await apiClient.get<T>(url, config);
    console.log(`GET ${url} response (status ${response.status}):`, response.data);
    
    // Log full response structure for debugging
    console.log('Response headers:', response.headers);
    console.log('Response structure:', JSON.stringify(response.data, null, 2));
    
    return response.data;
  } catch (error) {
    console.error(`GET ${url} failed:`, error);
    if (isAxiosError(error) && error.response) {
      console.error('Error response:', error.response.status, error.response.data);
    }
    throw error;
  }
};

export const post = async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
  try {
    const response = await apiClient.post<T>(url, data, config);
    return response.data;
  } catch (error) {
    console.error(`POST ${url} failed:`, error);
    throw error;
  }
};

export const put = async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
  try {
    const response = await apiClient.put<T>(url, data, config);
    return response.data;
  } catch (error) {
    console.error(`PUT ${url} failed:`, error);
    throw error;
  }
};

export const del = async <T>(url: string, config?: AxiosRequestConfig): Promise<T> => {
  try {
    const response = await apiClient.delete<T>(url, config);
    return response.data;
  } catch (error) {
    console.error(`DELETE ${url} failed:`, error);
    throw error;
  }
};

export default apiClient;