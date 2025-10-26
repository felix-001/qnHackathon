package service

import (
	"context"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
)

type MonitoringService struct{
	db *db.MongoDB
}

func NewMonitoringService(db *db.MongoDB) *MonitoringService {
	return &MonitoringService{db: db}
}

func (s *MonitoringService) GetRealtime(releaseID string) (*model.MonitoringMetrics, error) {
	return &model.MonitoringMetrics{
		RequestRate:      float64(rand.Intn(2000)),
		ErrorRate:        rand.Float64() * 0.01,
		LatencyP50:       float64(rand.Intn(100)),
		LatencyP95:       float64(rand.Intn(300)),
		LatencyP99:       float64(rand.Intn(500)),
		CPUUsage:         float64(rand.Intn(100)),
		MemoryUsage:      float64(rand.Intn(100)),
		FDCount:          float64(100 + rand.Intn(900)),
		ConnCount:        float64(50 + rand.Intn(950)),
		PacketLossRate:   rand.Float64() * 0.05,
		DiskUsage:        float64(40 + rand.Intn(50)),
		SystemLoad:       rand.Float64() * 8.0,
		NetworkBandwidth: float64(10 + rand.Intn(90)),
	}, nil
}

func (s *MonitoringService) GetNodesByProject(projectID string) ([]*model.NodeMonitoringMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}

	cursor, err := s.db.Database.Collection("device_gray_status").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var devices []*model.DeviceGrayStatus
	if err = cursor.All(ctx, &devices); err != nil {
		return nil, err
	}

	nodes := make([]*model.NodeMonitoringMetrics, 0, len(devices))
	for _, device := range devices {
		metrics := &model.NodeMonitoringMetrics{
			NodeID:           device.NodeID,
			NodeName:         device.NodeName,
			ProjectID:        device.ProjectID,
			ProjectName:      device.ProjectName,
			RequestRate:      float64(rand.Intn(2000)),
			ErrorRate:        rand.Float64() * 0.01,
			LatencyP50:       float64(rand.Intn(100)),
			LatencyP95:       float64(rand.Intn(300)),
			LatencyP99:       float64(rand.Intn(500)),
			CPUUsage:         float64(rand.Intn(100)),
			MemoryUsage:      float64(rand.Intn(100)),
			FDCount:          float64(100 + rand.Intn(900)),
			ConnCount:        float64(50 + rand.Intn(950)),
			PacketLossRate:   rand.Float64() * 0.05,
			DiskUsage:        float64(40 + rand.Intn(50)),
			SystemLoad:       rand.Float64() * 8.0,
			NetworkBandwidth: float64(10 + rand.Intn(90)),
			Status:           device.Status,
		}
		nodes = append(nodes, metrics)
	}

	return nodes, nil
}

func (s *MonitoringService) GetNodeMetrics(nodeID string) (*model.NodeMonitoringMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var device model.DeviceGrayStatus
	err := s.db.Database.Collection("device_gray_status").FindOne(ctx, bson.M{"nodeId": nodeID}).Decode(&device)
	if err != nil {
		return nil, err
	}

	return &model.NodeMonitoringMetrics{
		NodeID:           device.NodeID,
		NodeName:         device.NodeName,
		ProjectID:        device.ProjectID,
		ProjectName:      device.ProjectName,
		RequestRate:      float64(rand.Intn(2000)),
		ErrorRate:        rand.Float64() * 0.01,
		LatencyP50:       float64(rand.Intn(100)),
		LatencyP95:       float64(rand.Intn(300)),
		LatencyP99:       float64(rand.Intn(500)),
		CPUUsage:         float64(rand.Intn(100)),
		MemoryUsage:      float64(rand.Intn(100)),
		FDCount:          float64(100 + rand.Intn(900)),
		ConnCount:        float64(50 + rand.Intn(950)),
		PacketLossRate:   rand.Float64() * 0.05,
		DiskUsage:        float64(40 + rand.Intn(50)),
		SystemLoad:       rand.Float64() * 8.0,
		NetworkBandwidth: float64(10 + rand.Intn(90)),
		Status:           device.Status,
	}, nil
}

func (s *MonitoringService) GetTimeSeries(releaseID string) (*model.MonitoringTimeSeries, error) {
	now := time.Now().Unix()
	dataPoints := 30

	requestRate := make([]model.MetricDataPoint, dataPoints)
	errorRate := make([]model.MetricDataPoint, dataPoints)
	latencyP50 := make([]model.MetricDataPoint, dataPoints)
	latencyP99 := make([]model.MetricDataPoint, dataPoints)
	cpuUsage := make([]model.MetricDataPoint, dataPoints)
	memoryUsage := make([]model.MetricDataPoint, dataPoints)
	fdCount := make([]model.MetricDataPoint, dataPoints)
	connCount := make([]model.MetricDataPoint, dataPoints)

	packetLossRate := make([]model.MetricDataPoint, dataPoints)
	diskUsage := make([]model.MetricDataPoint, dataPoints)
	systemLoad := make([]model.MetricDataPoint, dataPoints)
	networkBandwidth := make([]model.MetricDataPoint, dataPoints)

	for i := 0; i < dataPoints; i++ {
		timestamp := now - int64((dataPoints-i-1)*10)
		requestRate[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(1500 + rand.Intn(500)),
		}
		errorRate[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     rand.Float64() * 0.01,
		}
		latencyP50[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(50 + rand.Intn(50)),
		}
		latencyP99[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(200 + rand.Intn(300)),
		}
		cpuUsage[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(30 + rand.Intn(50)),
		}
		memoryUsage[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(40 + rand.Intn(40)),
		}
		fdCount[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(100 + rand.Intn(900)),
		}
		connCount[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(50 + rand.Intn(950)),
		}
		packetLossRate[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     rand.Float64() * 0.05,
		}
		diskUsage[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(40 + rand.Intn(50)),
		}
		systemLoad[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     rand.Float64() * 8.0,
		}
		networkBandwidth[i] = model.MetricDataPoint{
			Timestamp: timestamp,
			Value:     float64(10 + rand.Intn(90)),
		}
	}

	return &model.MonitoringTimeSeries{
		RequestRate:      requestRate,
		ErrorRate:        errorRate,
		LatencyP50:       latencyP50,
		LatencyP99:       latencyP99,
		CPUUsage:         cpuUsage,
		MemoryUsage:      memoryUsage,
		FDCount:          fdCount,
		ConnCount:        connCount,
		PacketLossRate:   packetLossRate,
		DiskUsage:        diskUsage,
		SystemLoad:       systemLoad,
		NetworkBandwidth: networkBandwidth,
	}, nil
}
