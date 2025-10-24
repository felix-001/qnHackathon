package service

import (
	"github.com/felix-001/qnHackathon/internal/model"
	"math/rand"
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
	}, nil
}
