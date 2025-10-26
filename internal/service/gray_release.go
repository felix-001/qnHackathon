package service

import (
	"context"
	"fmt"
	"time"

	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GrayReleaseService struct {
	db *db.MongoDB
}

func NewGrayReleaseService(db *db.MongoDB) *GrayReleaseService {
	return &GrayReleaseService{db: db}
}

func (s *GrayReleaseService) CreateGrayRelease(config *model.GrayReleaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.Status = "active"

	result, err := s.db.Database.Collection("gray_releases").InsertOne(ctx, config)
	if err != nil {
		return err
	}

	config.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (s *GrayReleaseService) ListGrayReleases(projectID, environment string) ([]*model.GrayReleaseConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}
	if environment != "" {
		filter["environment"] = environment
	}

	cursor, err := s.db.Database.Collection("gray_releases").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var configs []*model.GrayReleaseConfig
	if err = cursor.All(ctx, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func (s *GrayReleaseService) GetGrayRelease(id string) (*model.GrayReleaseConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var config model.GrayReleaseConfig
	err = s.db.Database.Collection("gray_releases").FindOne(ctx, bson.M{"_id": objID}).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *GrayReleaseService) UpdateGrayRelease(id string, config *model.GrayReleaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	config.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"version":     config.Version,
			"rules":       config.Rules,
			"status":      config.Status,
			"description": config.Description,
			"updatedAt":   config.UpdatedAt,
		},
	}

	_, err = s.db.Database.Collection("gray_releases").UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (s *GrayReleaseService) DeleteGrayRelease(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = s.db.Database.Collection("gray_releases").DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (s *GrayReleaseService) GetDeviceStats(projectID, environment string) ([]*model.GrayReleaseStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if projectID != "" {
		filter["projectId"] = projectID
	}
	if environment != "" {
		filter["environment"] = environment
	}

	pipeline := []bson.M{
		{"$match": filter},
		{
			"$group": bson.M{
				"_id":   "$currentVersion",
				"count": bson.M{"$sum": 1},
				"devices": bson.M{
					"$push": bson.M{
						"isp":        "$isp",
						"region":     "$region",
						"province":   "$province",
						"dataCenter": "$dataCenter",
					},
				},
			},
		},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := s.db.Database.Collection("device_gray_status").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Version string `bson:"_id"`
		Count   int    `bson:"count"`
		Devices []struct {
			ISP        string `bson:"isp"`
			Region     string `bson:"region"`
			Province   string `bson:"province"`
			DataCenter string `bson:"dataCenter"`
		} `bson:"devices"`
	}

	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make([]*model.GrayReleaseStats, 0, len(results))
	for _, result := range results {
		dimensionStats := make(map[string]int)
		ispCount := make(map[string]int)
		regionCount := make(map[string]int)
		provinceCount := make(map[string]int)
		dcCount := make(map[string]int)

		for _, device := range result.Devices {
			if device.ISP != "" {
				ispCount[device.ISP]++
			}
			if device.Region != "" {
				regionCount[device.Region]++
			}
			if device.Province != "" {
				provinceCount[device.Province]++
			}
			if device.DataCenter != "" {
				dcCount[device.DataCenter]++
			}
		}

		for k, v := range ispCount {
			dimensionStats[fmt.Sprintf("isp_%s", k)] = v
		}
		for k, v := range regionCount {
			dimensionStats[fmt.Sprintf("region_%s", k)] = v
		}
		for k, v := range provinceCount {
			dimensionStats[fmt.Sprintf("province_%s", k)] = v
		}
		for k, v := range dcCount {
			dimensionStats[fmt.Sprintf("datacenter_%s", k)] = v
		}

		stats = append(stats, &model.GrayReleaseStats{
			Version:     result.Version,
			DeviceCount: result.Count,
			ByDimension: dimensionStats,
		})
	}

	return stats, nil
}

func (s *GrayReleaseService) FullRelease(projectID, environment, version string, operator string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"projectId":   projectID,
		"environment": environment,
	}

	update := bson.M{
		"$set": bson.M{
			"currentVersion": version,
			"status":         "released",
			"updatedAt":      time.Now(),
		},
	}

	_, err := s.db.Database.Collection("device_gray_status").UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	grayFilter := bson.M{
		"projectId":   projectID,
		"environment": environment,
		"status":      "active",
	}
	grayUpdate := bson.M{
		"$set": bson.M{
			"status":    "completed",
			"updatedAt": time.Now(),
		},
	}
	_, err = s.db.Database.Collection("gray_releases").UpdateMany(ctx, grayFilter, grayUpdate)

	return err
}

func (s *GrayReleaseService) UpdateDeviceStatus(device *model.DeviceGrayStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	device.UpdatedAt = time.Now()

	filter := bson.M{
		"nodeId":      device.NodeID,
		"projectId":   device.ProjectID,
		"environment": device.Environment,
	}

	update := bson.M{
		"$set": bson.M{
			"nodeName":       device.NodeName,
			"projectName":    device.ProjectName,
			"currentVersion": device.CurrentVersion,
			"isp":            device.ISP,
			"region":         device.Region,
			"province":       device.Province,
			"dataCenter":     device.DataCenter,
			"status":         device.Status,
			"updatedAt":      device.UpdatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.db.Database.Collection("device_gray_status").UpdateOne(ctx, filter, update, opts)
	return err
}

func (s *GrayReleaseService) CheckDeviceGrayRule(device *model.DeviceGrayStatus) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"projectId":   device.ProjectID,
		"environment": device.Environment,
		"status":      "active",
	}

	cursor, err := s.db.Database.Collection("gray_releases").Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return false, "", err
	}
	defer cursor.Close(ctx)

	var configs []*model.GrayReleaseConfig
	if err = cursor.All(ctx, &configs); err != nil {
		return false, "", err
	}

	for _, config := range configs {
		if s.matchGrayRule(device, config.Rules) {
			return true, config.Version, nil
		}
	}

	return false, "", nil
}

func (s *GrayReleaseService) matchGrayRule(device *model.DeviceGrayStatus, rules []model.GrayReleaseRule) bool {
	if len(rules) == 0 {
		return false
	}

	for _, rule := range rules {
		matched := false
		var deviceValue string

		switch rule.Dimension {
		case "isp":
			deviceValue = device.ISP
		case "region":
			deviceValue = device.Region
		case "province":
			deviceValue = device.Province
		case "datacenter":
			deviceValue = device.DataCenter
		default:
			continue
		}

		for _, value := range rule.Values {
			if value == deviceValue {
				matched = true
				break
			}
		}

		if !matched {
			return false
		}
	}

	return true
}
