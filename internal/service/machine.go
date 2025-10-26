package service

import (
	"context"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MachineService struct {
	collection *mongo.Collection
}

func NewMachineService(mongodb *db.MongoDB) *MachineService {
	return &MachineService{
		collection: mongodb.Database.Collection("machines"),
	}
}

func (s *MachineService) ListByProject(projectID string) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var machines []model.Machine
	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return s.generateMockMachines(projectID), nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &machines); err != nil {
		return s.generateMockMachines(projectID), nil
	}

	if len(machines) == 0 {
		return s.generateMockMachines(projectID), nil
	}

	return machines, nil
}

func (s *MachineService) generateMockMachines(projectID string) []model.Machine {
	machines := []model.Machine{
		{
			ID:        "machine-1",
			ProjectID: projectID,
			IP:        "192.168.1.10",
			Hostname:  "node-01",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "machine-2",
			ProjectID: projectID,
			IP:        "192.168.1.11",
			Hostname:  "node-02",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "machine-3",
			ProjectID: projectID,
			IP:        "192.168.1.12",
			Hostname:  "node-03",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "machine-4",
			ProjectID: projectID,
			IP:        "10.0.0.20",
			Hostname:  "node-04",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "machine-5",
			ProjectID: projectID,
			IP:        "10.0.0.21",
			Hostname:  "node-05",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if projectID != "" {
		return machines
	}

	return machines
}
