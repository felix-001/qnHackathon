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

func (s *MonitoringService) GetRealtime(releaseID string) (*model.MonitoringMetrics, error) {
	return &model.MonitoringMetrics{
		RequestRate: float64(rand.Intn(2000)),
		ErrorRate:   rand.Float64() * 0.01,
		LatencyP50:  float64(rand.Intn(100)),
		LatencyP95:  float64(rand.Intn(300)),
		LatencyP99:  float64(rand.Intn(500)),
		CPUUsage:    float64(rand.Intn(100)),
		MemoryUsage: float64(rand.Intn(100)),
		FDCount:     float64(100 + rand.Intn(900)),
		ConnCount:   float64(50 + rand.Intn(950)),
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
	}

	return &model.MonitoringTimeSeries{
		RequestRate: requestRate,
		ErrorRate:   errorRate,
		LatencyP50:  latencyP50,
		LatencyP99:  latencyP99,
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		FDCount:     fdCount,
		ConnCount:   connCount,
	}, nil
}
