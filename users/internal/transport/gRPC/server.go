package grpctransport

import (
	"context"
	"uuid"

	accountpb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	core_model "github.com/GinciuuKitakaze/users/internal/core/model"
	transport_mapper "github.com/GinciuuKitakaze/users/internal/transport/mapper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	accountpb.UnimplementedUserServiceServer
	userService UserService
	logger      *core_logger.Logger
}

func NewServer(userService UserService, logger *core_logger.Logger) *Server {
	return &Server{
		userService: userService,
		logger:      logger,
	}
}

type UserService interface {
	CreateUser(ctx context.Context, user core_model.CreateUser) (core_model.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (core_model.User, error)
	GetUsers(ctx context.Context, limit, offset int) ([]core_model.User, error)
	GetUserByEmailOrPhone(ctx context.Context, emailOrPhone string) (core_model.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) (core_model.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, userUpdate core_model.UpdateUser) (core_model.User, error)
}

func (s *Server) CreateUser(ctx context.Context, req *accountpb.CreateUserRequest) (*accountpb.CreateUserResponse, error) {
	u := req.GetUser()
	if u == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	user := transport_mapper.ToModelCreateUser(u)

	res, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.CreateUserResponse{
		User: transport_mapper.ToProtoUser(res),
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *accountpb.GetUserRequest) (*accountpb.GetUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	res, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.GetUserResponse{
		User: transport_mapper.ToProtoUser(res),
	}, nil
}

func (s *Server) GetUsers(ctx context.Context, req *accountpb.GetUsersRequest) (*accountpb.GetUsersResponse, error) {
	res, err := s.userService.GetUsers(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.GetUsersResponse{
		Users: transport_mapper.ToProtoUsers(res),
	}, nil
}

func (s *Server) GetUserByEmailOrPhone(ctx context.Context, req *accountpb.GetUserByEmailOrPhoneRequest) (*accountpb.GetUserByEmailOrPhoneResponse, error) {
	res, err := s.userService.GetUserByEmailOrPhone(ctx, req.EmailOrPhone)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.GetUserByEmailOrPhoneResponse{
		User: transport_mapper.ToProtoUser(res),
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *accountpb.DeleteUserRequest) (*accountpb.DeleteUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	res, err := s.userService.DeleteUser(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.DeleteUserResponse{
		User: transport_mapper.ToProtoUser(res),
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *accountpb.UpdateUserRequest) (*accountpb.UpdateUserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	u := req.GetUser()
	if u == nil {
		return nil, status.Error(codes.InvalidArgument, "user is required")
	}

	userUpdate := transport_mapper.ToModelUpdateUser(u)

	res, err := s.userService.UpdateUser(ctx, userID, userUpdate)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &accountpb.UpdateUserResponse{
		User: transport_mapper.ToProtoUser(res),
	}, nil
}
