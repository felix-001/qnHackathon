package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConfigService struct {
	db *db.MongoDB
}

func NewConfigService(db *db.MongoDB) *ConfigService {
	return &ConfigService{db: db}
}

func (s *ConfigService) List(projectID, environment string) ([]*model.Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}
	if environment != "" {
		filter["environment"] = environment
	}

	cursor, err := s.db.Database.Collection("configs").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []*model.Config
	if err = cursor.All(ctx, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func (s *ConfigService) Get(id string) (*model.Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var config model.Config
	err = s.db.Database.Collection("configs").FindOne(ctx, bson.M{"_id": objID}).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *ConfigService) GetByProjectAndEnv(projectID, environment string) (*model.Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var config model.Config
	err := s.db.Database.Collection("configs").FindOne(ctx, bson.M{
		"projectId":   projectID,
		"environment": environment,
	}).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *ConfigService) Create(config *model.Config, operator string, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	result, err := s.db.Database.Collection("configs").InsertOne(ctx, config)
	if err != nil {
		return err
	}

	config.ID = result.InsertedID.(primitive.ObjectID).Hex()

	history := &model.ConfigHistory{
		ConfigID:    config.ID,
		ProjectID:   config.ProjectID,
		ProjectName: config.ProjectName,
		Environment: config.Environment,
		FileName:    config.FileName,
		OldContent:  "",
		NewContent:  config.Content,
		ChangeType:  "create",
		Reason:      reason,
		Operator:    operator,
		CreatedAt:   time.Now(),
	}

	_, err = s.db.Database.Collection("config_history").InsertOne(ctx, history)
	return err
}

func (s *ConfigService) Update(id string, config *model.Config, operator string, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oldConfig, err := s.Get(id)
	if err != nil {
		return err
	}

	config.UpdatedAt = time.Now()
	config.ID = id

	update := bson.M{
		"$set": bson.M{
			"projectId":   config.ProjectID,
			"projectName": config.ProjectName,
			"environment": config.Environment,
			"fileName":    config.FileName,
			"content":     config.Content,
			"description": config.Description,
			"updatedAt":   config.UpdatedAt,
		},
	}

	_, err = s.db.Database.Collection("configs").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:    id,
		ProjectID:   config.ProjectID,
		ProjectName: config.ProjectName,
		Environment: config.Environment,
		FileName:    config.FileName,
		OldContent:  oldConfig.Content,
		NewContent:  config.Content,
		ChangeType:  "update",
		Reason:      reason,
		Operator:    operator,
		CreatedAt:   time.Now(),
	}

	_, err = s.db.Database.Collection("config_history").InsertOne(ctx, history)
	return err
}

func (s *ConfigService) Delete(id string, operator string, reason string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	config, err := s.Get(id)
	if err != nil {
		return err
	}

	_, err = s.db.Database.Collection("configs").DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:    id,
		ProjectID:   config.ProjectID,
		ProjectName: config.ProjectName,
		Environment: config.Environment,
		FileName:    config.FileName,
		OldContent:  config.Content,
		NewContent:  "",
		ChangeType:  "delete",
		Reason:      reason,
		Operator:    operator,
		CreatedAt:   time.Now(),
	}

	_, err = s.db.Database.Collection("config_history").InsertOne(ctx, history)
	return err
}

func (s *ConfigService) GetHistory(configID string) ([]*model.ConfigHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.Database.Collection("config_history").Find(
		ctx,
		bson.M{"configId": configID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []*model.ConfigHistory
	if err = cursor.All(ctx, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *ConfigService) GetHistoryByProject(projectID, environment string) ([]*model.ConfigHistory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"projectId": projectID}
	if environment != "" {
		filter["environment"] = environment
	}

	cursor, err := s.db.Database.Collection("config_history").Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []*model.ConfigHistory
	if err = cursor.All(ctx, &history); err != nil {
		return nil, err
	}

	return history, nil
}

func (s *ConfigService) CompareHistory(id1, id2 string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID1, err := primitive.ObjectIDFromHex(id1)
	if err != nil {
		return nil, err
	}

	objID2, err := primitive.ObjectIDFromHex(id2)
	if err != nil {
		return nil, err
	}

	var history1, history2 model.ConfigHistory
	err = s.db.Database.Collection("config_history").FindOne(ctx, bson.M{"_id": objID1}).Decode(&history1)
	if err != nil {
		return nil, err
	}

	err = s.db.Database.Collection("config_history").FindOne(ctx, bson.M{"_id": objID2}).Decode(&history2)
	if err != nil {
		return nil, err
	}

	var content1, content2 map[string]interface{}
	if history1.NewContent != "" {
		json.Unmarshal([]byte(history1.NewContent), &content1)
	}
	if history2.NewContent != "" {
		json.Unmarshal([]byte(history2.NewContent), &content2)
	}

	result := map[string]interface{}{
		"history1": history1,
		"history2": history2,
		"diff": map[string]interface{}{
			"content1": content1,
			"content2": content2,
		},
	}

	return result, nil
}

func (s *ConfigService) GetVersionStats(projectID, environment string) ([]*model.ConfigVersionStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"projectId":   projectID,
				"environment": environment,
			},
		},
		{
			"$group": bson.M{
				"_id":   "$version",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"version":     "$_id",
				"deviceCount": "$count",
				"_id":         0,
			},
		},
		{
			"$sort": bson.M{"version": 1},
		},
	}

	cursor, err := s.db.Database.Collection("config_deployments").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []*model.ConfigVersionStats
	if err = cursor.All(ctx, &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *ConfigService) GetVersionInconsistencies(projectID, environment string) ([]*model.VersionInconsistency, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"projectId":   projectID,
				"environment": environment,
			},
		},
		{
			"$addFields": bson.M{
				"majorVersion": bson.M{
					"$arrayElemAt": []interface{}{
						bson.M{"$split": []interface{}{"$version", "."}},
						0,
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$majorVersion",
				"count": bson.M{"$sum": 1},
				"nodes": bson.M{"$push": "$nodeId"},
			},
		},
		{
			"$match": bson.M{
				"count": bson.M{"$gt": 1},
			},
		},
		{
			"$project": bson.M{
				"majorVersion": "$_id",
				"deviceCount":  "$count",
				"nodeIds":      "$nodes",
				"_id":          0,
			},
		},
	}

	cursor, err := s.db.Database.Collection("config_deployments").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var inconsistencies []*model.VersionInconsistency
	if err = cursor.All(ctx, &inconsistencies); err != nil {
		return nil, err
	}

	return inconsistencies, nil
}

func (s *ConfigService) CreateCanaryRelease(canary *model.CanaryRelease) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	canary.CreatedAt = time.Now()
	canary.Status = "pending"

	result, err := s.db.Database.Collection("canary_releases").InsertOne(ctx, canary)
	if err != nil {
		return err
	}

	canary.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (s *ConfigService) ExecuteCanaryRelease(canaryID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(canaryID)
	if err != nil {
		return err
	}

	var canary model.CanaryRelease
	err = s.db.Database.Collection("canary_releases").FindOne(ctx, bson.M{"_id": objID}).Decode(&canary)
	if err != nil {
		return err
	}

	filter := bson.M{}
	switch canary.TargetGroup {
	case "operator":
		filter["operator"] = canary.TargetValue
	case "region":
		filter["region"] = canary.TargetValue
	case "province":
		filter["province"] = canary.TargetValue
	case "dataCenter":
		filter["dataCenter"] = canary.TargetValue
	default:
		return fmt.Errorf("invalid target group: %s", canary.TargetGroup)
	}

	cursor, err := s.db.Database.Collection("device_nodes").Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var nodes []*model.DeviceNode
	if err = cursor.All(ctx, &nodes); err != nil {
		return err
	}

	for _, node := range nodes {
		deployment := &model.ConfigDeployment{
			ConfigID:    canary.ConfigID,
			ProjectID:   canary.ProjectID,
			Environment: canary.Environment,
			Version:     canary.Version,
			NodeID:      node.NodeID,
			Status:      "deployed",
			DeployedAt:  time.Now(),
		}
		_, err = s.db.Database.Collection("config_deployments").InsertOne(ctx, deployment)
		if err != nil {
			return err
		}
	}

	completedAt := time.Now()
	_, err = s.db.Database.Collection("canary_releases").UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"status":      "completed",
				"completedAt": completedAt,
			},
		},
	)

	return err
}

func (s *ConfigService) ListCanaryReleases(projectID, environment string) ([]*model.CanaryRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}
	if environment != "" {
		filter["environment"] = environment
	}

	cursor, err := s.db.Database.Collection("canary_releases").Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var releases []*model.CanaryRelease
	if err = cursor.All(ctx, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

