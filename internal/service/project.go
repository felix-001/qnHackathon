package service

import (
	"context"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectService struct {
	collection *mongo.Collection
}

func NewProjectService(mongodb *db.MongoDB) *ProjectService {
	return &ProjectService{
		collection: mongodb.Database.Collection("projects"),
	}
}

func (s *ProjectService) List() []model.Project {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return []model.Project{}
	}
	defer cursor.Close(ctx)

	var projects []model.Project
	if err = cursor.All(ctx, &projects); err != nil {
		return []model.Project{}
	}

	return projects
}

func (s *ProjectService) Create(project *model.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	project.Status = "active"

	_, err := s.collection.InsertOne(ctx, project)
	return err
}

func (s *ProjectService) Update(id string, project *model.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	project.UpdatedAt = time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": project}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *ProjectService) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	_, err := s.collection.DeleteOne(ctx, filter)
	return err
}
