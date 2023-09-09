package agent

import (
	"context"
	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	pb "github.com/k8sdeploy/protobufs/generated/agent_service/v1"
	ConfigBuilder "github.com/keloran/go-config"
)

type Server struct {
	pb.UnimplementedAgentServiceServer
	ConfigBuilder.Config
}

func (s *Server) CreateAgent(ctx context.Context, in *pb.CreateAgentRequest) (*pb.CreateAgentResponse, error) {
	a := NewAgentService(ctx, s.Config, &RealMongoOperations{
		Collection: s.Mongo.Collections["agents"],
		Database:   s.Mongo.Database,
	})
	err := a.MongoOps.GetMongoClient(ctx, a.Mongo)
	if err != nil {
		return &pb.CreateAgentResponse{
			Status: pointerutil.StringPtr(err.Error()),
		}, nil
	}

	agentDetails, err := a.CreateAgent(in.GetAgentName(), in.GetCompanyId())
	if err != nil {
		return &pb.CreateAgentResponse{
			Status: pointerutil.StringPtr(err.Error()),
		}, nil
	}

	return &pb.CreateAgentResponse{
		CompanyId: agentDetails.CompanyId,
		Name:      agentDetails.Name,
		Id:        agentDetails.ID,
		Secret:    agentDetails.Secret,
	}, nil
}

func (s *Server) ValidateAgent(ctx context.Context, in *pb.ValidateAgentRequest) (*pb.ValidateAgentResponse, error) {
	a := NewAgentService(ctx, s.Config, &RealMongoOperations{
		Collection: s.Mongo.Collections["agents"],
		Database:   s.Mongo.Database,
	})
	err := a.MongoOps.GetMongoClient(ctx, a.Mongo)
	if err != nil {
		return &pb.ValidateAgentResponse{
			Status: pointerutil.StringPtr(err.Error()),
		}, nil
	}

	valid, err := a.ValidateAgent(in.GetId(), in.GetSecret())
	if err != nil {
		return &pb.ValidateAgentResponse{
			Status: pointerutil.StringPtr(err.Error()),
		}, nil
	}

	return &pb.ValidateAgentResponse{
		Valid: valid,
	}, nil
}
