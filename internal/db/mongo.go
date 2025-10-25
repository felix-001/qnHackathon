package db

import (
	"context"
	"time"

	"github.com/felix-001/qnHackathon/internal/config"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(conf config.MongoConf) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(conf.URL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to connect to MongoDB")
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to ping MongoDB")
		return nil, err
	}

	log.Logger.Info().Msg("Successfully connected to MongoDB")

	return &MongoDB{
		Client:   client,
		Database: client.Database(conf.Database),
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}
