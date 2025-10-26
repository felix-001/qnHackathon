package model

import "time"

type Project struct {
	ID             string    `json:"id" bson:"_id,omitempty"`
	Name           string    `json:"name" bson:"name"`
	Code           string    `json:"code" bson:"code"`
	Description    string    `json:"description" bson:"description"`
	Owner          string    `json:"owner" bson:"owner"`
	RepositoryURL  string    `json:"repositoryUrl" bson:"repositoryUrl"`
	GithubURL      string    `json:"githubUrl" bson:"githubUrl"`
	BuildTool      string    `json:"buildTool" bson:"buildTool"`
	DeploymentType string    `json:"deploymentType" bson:"deploymentType"`
	Status         string    `json:"status" bson:"status"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Application struct {
	ID             string    `json:"id" bson:"_id,omitempty"`
	ProjectID      string    `json:"projectId" bson:"projectId"`
	Name           string    `json:"name" bson:"name"`
	Code           string    `json:"code" bson:"code"`
	Description    string    `json:"description" bson:"description"`
	HealthCheckURL string    `json:"healthCheckUrl" bson:"healthCheckUrl"`
	CreatedAt      time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt" bson:"updatedAt"`
}

type Release struct {
	ID            string     `json:"id" bson:"_id,omitempty"`
	ProjectID     string     `json:"projectId" bson:"projectId"`
	ProjectName   string     `json:"projectName" bson:"projectName"`
	ApplicationID string     `json:"applicationId" bson:"applicationId"`
	Version       string     `json:"version" bson:"version"`
	Environment   string     `json:"environment" bson:"environment"`
	Strategy      string     `json:"strategy" bson:"strategy"`
	Status        string     `json:"status" bson:"status"`
	Description   string     `json:"description" bson:"description"`
	Scheduler     string     `json:"scheduler" bson:"scheduler"`
	GitlabPRURL   string     `json:"gitlabPrUrl" bson:"gitlabPrUrl"`
	TarFileName   string     `json:"tarFileName" bson:"tarFileName"`
	StartedAt     *time.Time `json:"startedAt,omitempty" bson:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt" bson:"createdAt"`
}

type ReleaseStage struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

type MonitoringMetrics struct {
	RequestRate     float64 `json:"requestRate"`
	ErrorRate       float64 `json:"errorRate"`
	LatencyP50      float64 `json:"latencyP50"`
	LatencyP95      float64 `json:"latencyP95"`
	LatencyP99      float64 `json:"latencyP99"`
	CPUUsage        float64 `json:"cpuUsage"`
	MemoryUsage     float64 `json:"memoryUsage"`
	FDCount         float64 `json:"fdCount"`
	ConnCount       float64 `json:"connCount"`
	PacketLossRate  float64 `json:"packetLossRate"`
	DiskUsage       float64 `json:"diskUsage"`
	SystemLoad      float64 `json:"systemLoad"`
	NetworkBandwidth float64 `json:"networkBandwidth"`
}

type MetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type MonitoringTimeSeries struct {
	RequestRate      []MetricDataPoint `json:"requestRate"`
	ErrorRate        []MetricDataPoint `json:"errorRate"`
	LatencyP50       []MetricDataPoint `json:"latencyP50"`
	LatencyP99       []MetricDataPoint `json:"latencyP99"`
	CPUUsage         []MetricDataPoint `json:"cpuUsage"`
	MemoryUsage      []MetricDataPoint `json:"memoryUsage"`
	FDCount          []MetricDataPoint `json:"fdCount"`
	ConnCount        []MetricDataPoint `json:"connCount"`
	PacketLossRate   []MetricDataPoint `json:"packetLossRate"`
	DiskUsage        []MetricDataPoint `json:"diskUsage"`
	SystemLoad       []MetricDataPoint `json:"systemLoad"`
	NetworkBandwidth []MetricDataPoint `json:"networkBandwidth"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Config struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	ProjectID   string    `json:"projectId" bson:"projectId"`
	ProjectName string    `json:"projectName" bson:"projectName"`
	Environment string    `json:"environment" bson:"environment"`
	FileName    string    `json:"fileName" bson:"fileName"`
	Content     string    `json:"content" bson:"content"`
	Description string    `json:"description" bson:"description"`
	Version     int       `json:"version" bson:"version"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt"`
}

type ConfigHistory struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	ConfigID    string    `json:"configId" bson:"configId"`
	ProjectID   string    `json:"projectId" bson:"projectId"`
	ProjectName string    `json:"projectName" bson:"projectName"`
	Environment string    `json:"environment" bson:"environment"`
	FileName    string    `json:"fileName" bson:"fileName"`
	OldContent  string    `json:"oldContent" bson:"oldContent"`
	NewContent  string    `json:"newContent" bson:"newContent"`
	ChangeType  string    `json:"changeType" bson:"changeType"`
	Reason      string    `json:"reason" bson:"reason"`
	Operator    string    `json:"operator" bson:"operator"`
	GitLabMR    string    `json:"gitlabMR,omitempty" bson:"gitlabMR,omitempty"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

type GrayReleaseRule struct {
	Dimension string   `json:"dimension" bson:"dimension"`
	Values    []string `json:"values" bson:"values"`
}

type GrayReleaseConfig struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	ConfigID    string             `json:"configId" bson:"configId"`
	ProjectID   string             `json:"projectId" bson:"projectId"`
	ProjectName string             `json:"projectName" bson:"projectName"`
	Environment string             `json:"environment" bson:"environment"`
	Version     string             `json:"version" bson:"version"`
	Rules       []GrayReleaseRule  `json:"rules" bson:"rules"`
	Status      string             `json:"status" bson:"status"`
	Operator    string             `json:"operator" bson:"operator"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type DeviceGrayStatus struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	NodeID       string    `json:"nodeId" bson:"nodeId"`
	NodeName     string    `json:"nodeName" bson:"nodeName"`
	ProjectID    string    `json:"projectId" bson:"projectId"`
	ProjectName  string    `json:"projectName" bson:"projectName"`
	Environment  string    `json:"environment" bson:"environment"`
	CurrentVersion string  `json:"currentVersion" bson:"currentVersion"`
	ISP          string    `json:"isp" bson:"isp"`
	Region       string    `json:"region" bson:"region"`
	Province     string    `json:"province" bson:"province"`
	DataCenter   string    `json:"dataCenter" bson:"dataCenter"`
	Status       string    `json:"status" bson:"status"`
	UpdatedAt    time.Time `json:"updatedAt" bson:"updatedAt"`
}

type GrayReleaseStats struct {
	Version      string            `json:"version"`
	DeviceCount  int               `json:"deviceCount"`
	ByDimension  map[string]int    `json:"byDimension"`
}
