package agent

import (
	"context"
	"errors"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/google/uuid"
	ConfigBuilder "github.com/keloran/go-config"
	mungo "github.com/keloran/go-config/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoOperations interface {
	GetMongoClient(ctx context.Context, config mungo.Mongo) error
	Disconnect(ctx context.Context) error
	InsertOne(ctx context.Context, document interface{}) (interface{}, error)
	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult
}

type RealMongoOperations struct {
	Client     *mongo.Client
	Collection string
	Database   string
}

func (r *RealMongoOperations) GetMongoClient(ctx context.Context, config mungo.Mongo) error {
	client, err := mungo.GetMongoClient(ctx, config)
	if err != nil {
		return logs.Errorf("error getting mongo client: %v", err)
	}
	r.Client = client
	return nil
}
func (r *RealMongoOperations) Disconnect(ctx context.Context) error {
	return r.Client.Disconnect(ctx)
}
func (r *RealMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	return r.Client.Database(r.Database).Collection(r.Collection).InsertOne(ctx, document)
}
func (r *RealMongoOperations) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return r.Client.Database(r.Database).Collection(r.Collection).FindOne(ctx, filter)
}

type Agent struct {
	ID        string
	Secret    string
	Name      string
	CompanyId string
}

type Service struct {
	ConfigBuilder.Config
	context.Context

	MongoOps MongoOperations
}

func NewAgentService(ctx context.Context, cfg ConfigBuilder.Config, ops MongoOperations) *Service {
	return &Service{
		Config:   cfg,
		Context:  ctx,
		MongoOps: ops,
	}
}

func (s *Service) CreateAgent(agentName, companyId string) (Agent, error) {
	if err := s.MongoOps.GetMongoClient(s.Context, s.Mongo); err != nil {
		return Agent{}, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := s.MongoOps.Disconnect(s.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	id, err := uuid.NewRandom()
	if err != nil {
		return Agent{}, logs.Errorf("error generating id uuid: %v", err)
	}
	secret, err := uuid.NewRandom()
	if err != nil {
		return Agent{}, logs.Errorf("error generating secret uuid: %v", err)
	}

	a := Agent{
		ID:        id.String(),
		Secret:    secret.String(),
		Name:      agentName,
		CompanyId: companyId,
	}

	_, err = s.MongoOps.InsertOne(s.Context, a)
	if err != nil {
		return Agent{}, logs.Errorf("error inserting agent: %v", err)
	}

	return a, nil
}

func (s *Service) ValidateAgent(id, secret string) (bool, error) {
	if s.Local.Development {
		return true, nil
	}

	if err := s.MongoOps.GetMongoClient(s.Context, s.Mongo); err != nil {
		return false, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := s.MongoOps.Disconnect(s.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	a := Agent{}
	if err := s.MongoOps.FindOne(s.Context, &bson.M{
		"id":     id,
		"secret": secret,
	}).Decode(&a); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, logs.Errorf("error decoding agent: %v", err)
	}

	return true, nil
}
