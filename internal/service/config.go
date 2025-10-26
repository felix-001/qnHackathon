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
		ConfigID:   config.ID,
		ProjectID:  config.ProjectID,
		Key:        config.Key,
		OldValue:   "",
		NewValue:   config.Value,
		ChangeType: "create",
		Reason:     reason,
		Operator:   operator,
		CreatedAt:  time.Now(),
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
			"key":         config.Key,
			"value":       config.Value,
			"environment": config.Environment,
			"description": config.Description,
			"grayConfig":  config.GrayConfig,
			"updatedAt":   config.UpdatedAt,
		},
	}

	_, err = s.db.Database.Collection("configs").UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	history := &model.ConfigHistory{
		ConfigID:   id,
		ProjectID:  config.ProjectID,
		Key:        config.Key,
		OldValue:   oldConfig.Value,
		NewValue:   config.Value,
		ChangeType: "update",
		Reason:     reason,
		Operator:   operator,
		CreatedAt:  time.Now(),
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
		ConfigID:   id,
		ProjectID:  config.ProjectID,
		Key:        config.Key,
		OldValue:   config.Value,
		NewValue:   "",
		ChangeType: "delete",
		Reason:     reason,
		Operator:   operator,
		CreatedAt:  time.Now(),
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

	result := map[string]interface{}{
		"history1": history1,
		"history2": history2,
		"diff": map[string]interface{}{
			"key":       history1.Key == history2.Key,
			"oldValue":  history1.OldValue,
			"newValue1": history1.NewValue,
			"newValue2": history2.NewValue,
		},
	}

	return result, nil
}

func (s *ConfigService) SubmitToGitLab(config *model.Config, gitlabMgr *GitLabMgr, operator string, reason string) (string, error) {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("configs/%s/%s_%s.json", config.Environment, config.ProjectID, config.Key)
	branchName := fmt.Sprintf("config-update-%s-%d", config.Key, time.Now().Unix())
	commitMessage := fmt.Sprintf("Update config: %s\n\nReason: %s\nOperator: %s", config.Key, reason, operator)

	err = gitlabMgr.CreateOrUpdateFile(fileName, string(configJSON), branchName, commitMessage)
	if err != nil {
		return "", err
	}

	mrTitle := fmt.Sprintf("配置更新: %s", config.Key)
	mrDescription := fmt.Sprintf("**配置键**: %s\n**环境**: %s\n**修改原因**: %s\n**操作人**: %s",
		config.Key, config.Environment, reason, operator)

	mrURL, err := gitlabMgr.CreateMergeRequest(branchName, "main", mrTitle, mrDescription)
	if err != nil {
		return "", err
	}

	return mrURL, nil
}

func (s *ConfigService) SubmitToGitHub(config *model.Config, githubMgr *GitHubMgr, operator string, reason string, projectName string) (string, error) {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("configs/%s/%s.json", projectName, config.Environment)
	branchName := fmt.Sprintf("config-update-%s-%s-%d", projectName, config.Environment, time.Now().Unix())
	commitMessage := fmt.Sprintf("Update config: %s/%s\n\nReason: %s\nOperator: %s", projectName, config.Environment, reason, operator)

	err = githubMgr.CreateBranch("main", branchName)
	if err != nil {
		return "", err
	}

	err = githubMgr.CreateOrUpdateFile(fileName, string(configJSON), branchName, commitMessage)
	if err != nil {
		return "", err
	}

	prTitle := fmt.Sprintf("配置更新: %s/%s", projectName, config.Environment)
	prDescription := fmt.Sprintf("**项目**: %s\n**环境**: %s\n**修改原因**: %s\n**操作人**: %s",
		projectName, config.Environment, reason, operator)

	prURL, err := githubMgr.CreatePullRequest(branchName, "main", prTitle, prDescription)
	if err != nil {
		return "", err
	}

	return prURL, nil
}
