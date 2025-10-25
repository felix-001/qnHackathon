package service

import (
	"context"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReleaseService struct {
	collection *mongo.Collection
}

func NewReleaseService(mongodb *db.MongoDB) *ReleaseService {
	return &ReleaseService{
		collection: mongodb.Database.Collection("releases"),
	}
}

func (s *ReleaseService) List() []model.Release {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return []model.Release{}
	}
	defer cursor.Close(ctx)

	var releases []model.Release
	if err = cursor.All(ctx, &releases); err != nil {
		return []model.Release{}
	}

	return releases
}

func (s *ReleaseService) Create(release *model.Release) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release.CreatedAt = time.Now()
	release.Status = "pending_approval"

	_, err := s.collection.InsertOne(ctx, release)
	return err
}

func (s *ReleaseService) Get(id string) (*model.Release, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var release model.Release
	filter := bson.M{"_id": id}
	err := s.collection.FindOne(ctx, filter).Decode(&release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func (s *ReleaseService) Rollback(id string, targetVersion string, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "rolled_back"}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *ReleaseService) UpdateGitlabPR(id string, gitlabPRURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"gitlabPRURL": gitlabPRURL}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *ReleaseService) ApproveReview(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "approved"}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *ReleaseService) StartDeploy(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"status":    "deploying",
		"startedAt": now,
	}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *ReleaseService) CompleteDeploy(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"status":      "completed",
		"completedAt": now,
	}}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}
