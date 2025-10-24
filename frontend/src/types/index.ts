export interface Project {
  id: string;
  name: string;
  code: string;
  description: string;
  owner: string;
  repositoryUrl: string;
  buildTool: string;
  deploymentType: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface Application {
  id: string;
  projectId: string;
  name: string;
  code: string;
  description: string;
  healthCheckUrl: string;
  createdAt: string;
  updatedAt: string;
}

export interface Release {
  id: string;
  projectId: string;
  applicationId: string;
  version: string;
  environment: string;
  strategy: string;
  status: string;
  description: string;
  scheduler: string;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
}

export interface ReleaseStage {
  name: string;
  status: string;
  startTime?: string;
  endTime?: string;
}

export interface MonitoringMetrics {
  requestRate: number;
  errorRate: number;
  latencyP50: number;
  latencyP95: number;
  latencyP99: number;
}
