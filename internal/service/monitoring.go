package service

import (
	"github.com/felix-001/qnHackathon/internal/model"
	"math/rand"
	"time"
)

type MonitoringService struct{}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{}
}

func (s *MonitoringService) GetRealtime(releaseID, machineID string) (*model.MonitoringMetrics, error) {
	seed := int64(0)
	if machineID != "" {
		for _, c := range machineID {
			seed += int64(c)
		}
		rand.Seed(seed)
	}

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

func (s *MonitoringService) GetTimeSeries(releaseID, machineID string) (*model.MonitoringTimeSeries, error) {
	now := time.Now().Unix()
	dataPoints := 30

	seed := int64(0)
	if machineID != "" {
		for _, c := range machineID {
			seed += int64(c)
		}
		rand.Seed(seed + now)
	}

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
