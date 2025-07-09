package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/par1ram/silence/api/auth/api/proto"
	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/ports"
)

// Server represents the gRPC server for auth service
type Server struct {
	pb.UnimplementedAuthServiceServer
	authService ports.AuthService
	userService ports.UserService
	port        int
	server      *grpc.Server
}

// NewServer creates a new gRPC server
func NewServer(port int, authService ports.AuthService, userService ports.UserService) *Server {
	return &Server{
		authService: authService,
		userService: userService,
		port:        port,
	}
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterAuthServiceServer(s.server, s)

	log.Printf("[grpc] auth server listening on port %d", s.port)
	return s.server.Serve(lis)
}

// Stop stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		log.Printf("[grpc] stopping auth server")
		s.server.GracefulStop()
	}
	return nil
}

// Name returns the service name
func (s *Server) Name() string {
	return "auth-grpc"
}

// Health check
func (s *Server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// Login authenticates user
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	loginReq := &domain.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := s.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	return &pb.AuthResponse{
		Token: response.Token,
		User:  convertUserToProto(response.User),
	}, nil
}

// Register creates a new user account
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	registerReq := &domain.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := s.authService.Register(ctx, registerReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "registration failed: %v", err)
	}

	return &pb.AuthResponse{
		Token: response.Token,
		User:  convertUserToProto(response.User),
	}, nil
}

// GetMe gets current user information
func (s *Server) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.User, error) {
	user, err := s.authService.GetMe(ctx, req.Token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user: %v", err)
	}

	return convertUserToProto(user), nil
}

// CreateUser creates a new user
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	createReq := &domain.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Role:     convertProtoToUserRole(req.Role),
	}

	user, err := s.userService.CreateUser(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create user: %v", err)
	}

	return convertUserToProto(user), nil
}

// GetUser retrieves a user by ID
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return convertUserToProto(user), nil
}

// UpdateUser updates user information
func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	updateReq := &domain.UpdateUserRequest{
		Email:  req.Email,
		Role:   convertProtoToUserRole(req.Role),
		Status: convertProtoToUserStatus(req.Status),
	}

	user, err := s.userService.UpdateUser(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to update user: %v", err)
	}

	return convertUserToProto(user), nil
}

// DeleteUser deletes a user
func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to delete user: %v", err),
		}, nil
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// ListUsers lists users with filtering
func (s *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	filter := &domain.UserFilter{
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	if req.Role != pb.UserRole_USER_ROLE_UNSPECIFIED {
		role := convertProtoToUserRole(req.Role)
		filter.Role = &role
	}

	if req.Status != pb.UserStatus_USER_STATUS_UNSPECIFIED {
		status := convertProtoToUserStatus(req.Status)
		filter.Status = &status
	}

	if req.Email != "" {
		filter.Email = &req.Email
	}

	response, err := s.userService.ListUsers(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	users := make([]*pb.User, len(response.Users))
	for i, user := range response.Users {
		users[i] = convertUserToProto(user)
	}

	return &pb.ListUsersResponse{
		Users: users,
		Total: int32(response.Total),
	}, nil
}

// BlockUser blocks a user
func (s *Server) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	err := s.userService.BlockUser(ctx, req.Id)
	if err != nil {
		return &pb.BlockUserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to block user: %v", err),
		}, nil
	}

	return &pb.BlockUserResponse{
		Success: true,
		Message: "User blocked successfully",
	}, nil
}

// UnblockUser unblocks a user
func (s *Server) UnblockUser(ctx context.Context, req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
	err := s.userService.UnblockUser(ctx, req.Id)
	if err != nil {
		return &pb.UnblockUserResponse{
			Success: false,
			Message: fmt.Sprintf("failed to unblock user: %v", err),
		}, nil
	}

	return &pb.UnblockUserResponse{
		Success: true,
		Message: "User unblocked successfully",
	}, nil
}

// ChangeUserRole changes user role
func (s *Server) ChangeUserRole(ctx context.Context, req *pb.ChangeUserRoleRequest) (*pb.ChangeUserRoleResponse, error) {
	role := convertProtoToUserRole(req.Role)
	err := s.userService.ChangeUserRole(ctx, req.Id, role)
	if err != nil {
		return &pb.ChangeUserRoleResponse{
			Success: false,
			Message: fmt.Sprintf("failed to change user role: %v", err),
		}, nil
	}

	return &pb.ChangeUserRoleResponse{
		Success: true,
		Message: "User role changed successfully",
	}, nil
}

// Helper functions for conversion

func convertUserToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Role:      convertUserRoleToProto(user.Role),
		Status:    convertUserStatusToProto(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func convertUserRoleToProto(role domain.UserRole) pb.UserRole {
	switch role {
	case domain.RoleUser:
		return pb.UserRole_USER_ROLE_USER
	case domain.RoleModerator:
		return pb.UserRole_USER_ROLE_MODERATOR
	case domain.RoleAdmin:
		return pb.UserRole_USER_ROLE_ADMIN
	default:
		return pb.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func convertProtoToUserRole(role pb.UserRole) domain.UserRole {
	switch role {
	case pb.UserRole_USER_ROLE_USER:
		return domain.RoleUser
	case pb.UserRole_USER_ROLE_MODERATOR:
		return domain.RoleModerator
	case pb.UserRole_USER_ROLE_ADMIN:
		return domain.RoleAdmin
	default:
		return domain.RoleUser
	}
}

func convertUserStatusToProto(status domain.UserStatus) pb.UserStatus {
	switch status {
	case domain.StatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case domain.StatusInactive:
		return pb.UserStatus_USER_STATUS_INACTIVE
	case domain.StatusBlocked:
		return pb.UserStatus_USER_STATUS_BLOCKED
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func convertProtoToUserStatus(status pb.UserStatus) domain.UserStatus {
	switch status {
	case pb.UserStatus_USER_STATUS_ACTIVE:
		return domain.StatusActive
	case pb.UserStatus_USER_STATUS_INACTIVE:
		return domain.StatusInactive
	case pb.UserStatus_USER_STATUS_BLOCKED:
		return domain.StatusBlocked
	default:
		return domain.StatusActive
	}
}
