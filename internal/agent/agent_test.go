package agent

import (
	"context"
	"errors"
	ConfigBuilder "github.com/keloran/go-config"
	mungo "github.com/keloran/go-config/mongo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

type MockMongoOperations struct {
	shouldError bool
	exists      bool
	mockedAgent *Agent
}

func (m *MockMongoOperations) GetMongoClient(ctx context.Context, config mungo.Mongo) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

func (m *MockMongoOperations) Disconnect(ctx context.Context) error {
	return nil
}

func (m *MockMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	return nil, nil
}

func (m *MockMongoOperations) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	doc := bson.D{}

	if m.shouldError {
		// This should be adjusted based on how you handle errors in the FindOne method.
		return mongo.NewSingleResultFromDocument(nil, errors.New("mock error"), bson.DefaultRegistry)
	}

	if m.mockedAgent != nil {
		// Return the mocked agent
		return mongo.NewSingleResultFromDocument(doc, nil, bson.DefaultRegistry)
	}
	// Simulate a "not found" scenario
	return mongo.NewSingleResultFromDocument(doc, mongo.ErrNoDocuments, bson.DefaultRegistry)
}

func TestService_CreateAgent(t *testing.T) {
	ctx := context.Background()
	cfg := ConfigBuilder.Config{
		Mongo: mungo.Mongo{
			Database:    "test",
			Collections: map[string]string{"agents": "agents"},
		},
	}

	t.Run("CreateAgent", func(t *testing.T) {
		a := NewAgentService(ctx, cfg, &MockMongoOperations{})
		_, err := a.CreateAgent("test", "test")
		assert.Nil(t, err)
	})
}

func TestService_ValidateAgent(t *testing.T) {
	ctx := context.Background()
	cfg := ConfigBuilder.Config{
		Mongo: mungo.Mongo{
			Database:    "test",
			Collections: map[string]string{"agents": "agents"},
		},
	}

	t.Run("ValidateAgent_Found", func(t *testing.T) {
		a := NewAgentService(ctx, cfg, &MockMongoOperations{
			mockedAgent: &Agent{}, // presence of a mocked agent means the agent is found
		})
		valid, err := a.ValidateAgent("test", "test")
		assert.Nil(t, err)
		assert.True(t, valid)
	})

	t.Run("ValidateAgent_Not_Found", func(t *testing.T) {
		a := NewAgentService(ctx, cfg, &MockMongoOperations{})
		valid, err := a.ValidateAgent("test", "test")
		assert.NotNil(t, err)
		assert.False(t, valid)
		assert.Error(t, err)
	})

	t.Run("ValidateAgent_Error", func(t *testing.T) {
		a := NewAgentService(ctx, cfg, &MockMongoOperations{
			shouldError: true,
		})
		valid, err := a.ValidateAgent("test", "test")
		assert.NotNil(t, err)
		assert.False(t, valid)
		assert.Error(t, err)
	})
}
