import axios from 'axios';
import type { Project, Application, Release, MonitoringMetrics } from '@/types';

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export const projectAPI = {
  list: () => api.get<any, { data: Project[] }>('/projects'),
  create: (data: Partial<Project>) => api.post('/projects', data),
  update: (id: string, data: Partial<Project>) => api.put(`/projects/${id}`, data),
  delete: (id: string) => api.delete(`/projects/${id}`),
};

export const applicationAPI = {
  list: (projectId: string) => api.get<any, { data: Application[] }>(`/projects/${projectId}/applications`),
  create: (data: Partial<Application>) => api.post('/applications', data),
  update: (id: string, data: Partial<Application>) => api.put(`/applications/${id}`, data),
  delete: (id: string) => api.delete(`/applications/${id}`),
};

export const releaseAPI = {
  list: () => api.get<any, { data: Release[] }>('/releases'),
  create: (data: Partial<Release>) => api.post('/releases', data),
  get: (id: string) => api.get<any, { data: Release }>(`/releases/${id}`),
  rollback: (id: string, targetVersion: string, reason: string) => 
    api.post(`/releases/${id}/rollback`, { targetVersion, reason }),
};

export const monitoringAPI = {
  getRealtime: (releaseId: string) => 
    api.get<any, { data: { metrics: MonitoringMetrics } }>(`/monitoring/realtime?releaseId=${releaseId}`),
};

export default api;
