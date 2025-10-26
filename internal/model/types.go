package model

import (
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type FlexibleVersion string

func (v *FlexibleVersion) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	raw := bson.RawValue{Type: t, Value: data}
	
	switch t {
	case bsontype.String:
		var s string
		if err := raw.Unmarshal(&s); err != nil {
			return err
		}
		*v = FlexibleVersion(s)
		return nil
	case bsontype.Int32:
		var i int32
		if err := raw.Unmarshal(&i); err != nil {
			return err
		}
		*v = FlexibleVersion(fmt.Sprintf("v1.0.%d", i))
		return nil
	case bsontype.Int64:
		var i int64
		if err := raw.Unmarshal(&i); err != nil {
			return err
		}
		*v = FlexibleVersion(fmt.Sprintf("v1.0.%d", i))
		return nil
	default:
		return fmt.Errorf("cannot decode %s into FlexibleVersion", t)
	}
}

func (v FlexibleVersion) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(string(v))
}

func (v FlexibleVersion) String() string {
	return string(v)
}

func (v FlexibleVersion) ToInt() (int, error) {
	s := string(v)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

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
	ID          string          `json:"id" bson:"_id,omitempty"`
	ProjectID   string          `json:"projectId" bson:"projectId"`
	ProjectName string          `json:"projectName" bson:"projectName"`
	Environment string          `json:"environment" bson:"environment"`
	FileName    string          `json:"fileName" bson:"fileName"`
	Content     string          `json:"content" bson:"content"`
	Description string          `json:"description" bson:"description"`
	Version     FlexibleVersion `json:"version" bson:"version"`
	Approver    string          `json:"approver" bson:"approver"`
	CreatedAt   time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt" bson:"updatedAt"`
}

type ConfigHistory struct {
	ID          string          `json:"id" bson:"_id,omitempty"`
	ConfigID    string          `json:"configId" bson:"configId"`
	ProjectID   string          `json:"projectId" bson:"projectId"`
	ProjectName string          `json:"projectName" bson:"projectName"`
	Environment string          `json:"environment" bson:"environment"`
	FileName    string          `json:"fileName" bson:"fileName"`
	OldContent  string          `json:"oldContent" bson:"oldContent"`
	NewContent  string          `json:"newContent" bson:"newContent"`
	ChangeType  string          `json:"changeType" bson:"changeType"`
	Reason      string          `json:"reason" bson:"reason"`
	Operator    string          `json:"operator" bson:"operator"`
	Approver    string          `json:"approver" bson:"approver"`
	GitLabMR    string          `json:"gitlabMR,omitempty" bson:"gitlabMR,omitempty"`
	Version     FlexibleVersion `json:"version" bson:"version"`
	CreatedAt   time.Time       `json:"createdAt" bson:"createdAt"`
}

type GrayReleaseRule struct {
	Dimension string   `json:"dimension" bson:"dimension"`
	Values    []string `json:"values" bson:"values"`
}

type TimeWindow struct {
	StartTime time.Time `json:"startTime" bson:"startTime"`
	EndTime   time.Time `json:"endTime" bson:"endTime"`
}

type GrayReleaseStrategy struct {
	Type       string      `json:"type" bson:"type"`
	Dimension  string      `json:"dimension,omitempty" bson:"dimension,omitempty"`
	Values     []string    `json:"values,omitempty" bson:"values,omitempty"`
	Percentage int         `json:"percentage,omitempty" bson:"percentage,omitempty"`
	HashKey    string      `json:"hashKey,omitempty" bson:"hashKey,omitempty"`
	TimeWindow *TimeWindow `json:"timeWindow,omitempty" bson:"timeWindow,omitempty"`
}

type GrayReleaseConfig struct {
	ID           string                  `json:"id" bson:"_id,omitempty"`
	ConfigID     string                  `json:"configId" bson:"configId"`
	ProjectID    string                  `json:"projectId" bson:"projectId"`
	ProjectName  string                  `json:"projectName" bson:"projectName"`
	Environment  string                  `json:"environment" bson:"environment"`
	Version      string                  `json:"version" bson:"version"`
	Rules        []GrayReleaseRule       `json:"rules" bson:"rules"`
	StrategyType string                  `json:"strategyType,omitempty" bson:"strategyType,omitempty"`
	Strategies   []GrayReleaseStrategy   `json:"strategies,omitempty" bson:"strategies,omitempty"`
	RuleLogic    string                  `json:"ruleLogic,omitempty" bson:"ruleLogic,omitempty"`
	Stage        int                     `json:"stage,omitempty" bson:"stage,omitempty"`
	Status       string                  `json:"status" bson:"status"`
	Operator     string                  `json:"operator" bson:"operator"`
	Description  string                  `json:"description" bson:"description"`
	CreatedAt    time.Time               `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt" bson:"updatedAt"`
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

type Machine struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	ProjectID string    `json:"projectId" bson:"projectId"`
	IP        string    `json:"ip" bson:"ip"`
	Hostname  string    `json:"hostname" bson:"hostname"`
	Status    string    `json:"status" bson:"status"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
