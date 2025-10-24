package model

import "time"

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Description    string    `json:"description"`
	Owner          string    `json:"owner"`
	RepositoryURL  string    `json:"repositoryUrl"`
	BuildTool      string    `json:"buildTool"`
	DeploymentType string    `json:"deploymentType"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Application struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	Description    string    `json:"description"`
	HealthCheckURL string    `json:"healthCheckUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Release struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"projectId"`
	ApplicationID string     `json:"applicationId"`
	Version       string     `json:"version"`
	Environment   string     `json:"environment"`
	Strategy      string     `json:"strategy"`
	Status        string     `json:"status"`
	Description   string     `json:"description"`
	Scheduler     string     `json:"scheduler"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ReleaseStage struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

type MonitoringMetrics struct {
	RequestRate float64 `json:"requestRate"`
	ErrorRate   float64 `json:"errorRate"`
	LatencyP50  float64 `json:"latencyP50"`
	LatencyP95  float64 `json:"latencyP95"`
	LatencyP99  float64 `json:"latencyP99"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
