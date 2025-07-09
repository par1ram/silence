package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/par1ram/silence/api/auth/api/proto"
	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/mocks"
)

func TestServer_Health(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	resp, err := server.Health(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.NotNil(t, resp.Timestamp)
}

func TestServer_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful login", func(t *testing.T) {
		req := &pb.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedUser := &domain.User{
			ID:        "user123",
			Email:     "test@example.com",
			Role:      domain.RoleUser,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		expectedResponse := &domain.AuthResponse{
			Token: "jwt-token-123",
			User:  expectedUser,
		}

		mockAuthService.EXPECT().
			Login(gomock.Any(), &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			}).
			Return(expectedResponse, nil)

		resp, err := server.Login(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "jwt-token-123", resp.Token)
		assert.Equal(t, "user123", resp.User.Id)
		assert.Equal(t, "test@example.com", resp.User.Email)
		assert.Equal(t, pb.UserRole_USER_ROLE_USER, resp.User.Role)
		assert.Equal(t, pb.UserStatus_USER_STATUS_ACTIVE, resp.User.Status)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req := &pb.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		mockAuthService.EXPECT().
			Login(gomock.Any(), &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			}).
			Return(nil, domain.ErrInvalidCredentials)

		resp, err := server.Login(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}

func TestServer_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful registration", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Email:    "newuser@example.com",
			Password: "password123",
		}

		expectedUser := &domain.User{
			ID:        "user456",
			Email:     "newuser@example.com",
			Role:      domain.RoleUser,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		expectedResponse := &domain.AuthResponse{
			Token: "jwt-token-456",
			User:  expectedUser,
		}

		mockAuthService.EXPECT().
			Register(gomock.Any(), &domain.RegisterRequest{
				Email:    "newuser@example.com",
				Password: "password123",
			}).
			Return(expectedResponse, nil)

		resp, err := server.Register(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "jwt-token-456", resp.Token)
		assert.Equal(t, "user456", resp.User.Id)
		assert.Equal(t, "newuser@example.com", resp.User.Email)
	})

	t.Run("user already exists", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}

		mockAuthService.EXPECT().
			Register(gomock.Any(), &domain.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			}).
			Return(nil, domain.ErrUserAlreadyExists)

		resp, err := server.Register(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServer_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful get me", func(t *testing.T) {
		req := &pb.GetMeRequest{
			Token: "jwt-token-123",
		}

		expectedUser := &domain.User{
			ID:        "user123",
			Email:     "test@example.com",
			Role:      domain.RoleAdmin,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockAuthService.EXPECT().
			GetMe(gomock.Any(), "jwt-token-123").
			Return(expectedUser, nil)

		resp, err := server.GetMe(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "user123", resp.Id)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, pb.UserRole_USER_ROLE_ADMIN, resp.Role)
		assert.Equal(t, pb.UserStatus_USER_STATUS_ACTIVE, resp.Status)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := &pb.GetMeRequest{
			Token: "invalid-token",
		}

		mockAuthService.EXPECT().
			GetMe(gomock.Any(), "invalid-token").
			Return(nil, domain.ErrInvalidCredentials)

		resp, err := server.GetMe(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}

func TestServer_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful user creation", func(t *testing.T) {
		req := &pb.CreateUserRequest{
			Email:    "admin@example.com",
			Password: "adminpass123",
			Role:     pb.UserRole_USER_ROLE_ADMIN,
		}

		expectedUser := &domain.User{
			ID:        "admin123",
			Email:     "admin@example.com",
			Role:      domain.RoleAdmin,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockUserService.EXPECT().
			CreateUser(gomock.Any(), &domain.CreateUserRequest{
				Email:    "admin@example.com",
				Password: "adminpass123",
				Role:     domain.RoleAdmin,
			}).
			Return(expectedUser, nil)

		resp, err := server.CreateUser(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "admin123", resp.Id)
		assert.Equal(t, "admin@example.com", resp.Email)
		assert.Equal(t, pb.UserRole_USER_ROLE_ADMIN, resp.Role)
	})
}

func TestServer_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful get user", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Id: "user123",
		}

		expectedUser := &domain.User{
			ID:        "user123",
			Email:     "test@example.com",
			Role:      domain.RoleUser,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockUserService.EXPECT().
			GetUser(gomock.Any(), "user123").
			Return(expectedUser, nil)

		resp, err := server.GetUser(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "user123", resp.Id)
		assert.Equal(t, "test@example.com", resp.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		req := &pb.GetUserRequest{
			Id: "nonexistent",
		}

		mockUserService.EXPECT().
			GetUser(gomock.Any(), "nonexistent").
			Return(nil, domain.ErrUserNotFound)

		resp, err := server.GetUser(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)

		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_ListUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful list users", func(t *testing.T) {
		req := &pb.ListUsersRequest{
			Role:   pb.UserRole_USER_ROLE_USER,
			Status: pb.UserStatus_USER_STATUS_ACTIVE,
			Limit:  10,
			Offset: 0,
		}

		expectedUsers := []*domain.User{
			{
				ID:        "user1",
				Email:     "user1@example.com",
				Role:      domain.RoleUser,
				Status:    domain.StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "user2",
				Email:     "user2@example.com",
				Role:      domain.RoleUser,
				Status:    domain.StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		expectedResponse := &domain.UserListResponse{
			Users: expectedUsers,
			Total: 2,
		}

		role := domain.RoleUser
		status := domain.StatusActive
		mockUserService.EXPECT().
			ListUsers(gomock.Any(), &domain.UserFilter{
				Role:   &role,
				Status: &status,
				Limit:  10,
				Offset: 0,
			}).
			Return(expectedResponse, nil)

		resp, err := server.ListUsers(context.Background(), req)
		require.NoError(t, err)
		assert.Len(t, resp.Users, 2)
		assert.Equal(t, int32(2), resp.Total)
		assert.Equal(t, "user1", resp.Users[0].Id)
		assert.Equal(t, "user2", resp.Users[1].Id)
	})
}

func TestServer_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful delete user", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Id: "user123",
		}

		mockUserService.EXPECT().
			DeleteUser(gomock.Any(), "user123").
			Return(nil)

		resp, err := server.DeleteUser(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "User deleted successfully", resp.Message)
	})

	t.Run("delete user fails", func(t *testing.T) {
		req := &pb.DeleteUserRequest{
			Id: "user123",
		}

		mockUserService.EXPECT().
			DeleteUser(gomock.Any(), "user123").
			Return(domain.ErrUserNotFound)

		resp, err := server.DeleteUser(context.Background(), req)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "failed to delete user")
	})
}

func TestServer_BlockUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful block user", func(t *testing.T) {
		req := &pb.BlockUserRequest{
			Id: "user123",
		}

		mockUserService.EXPECT().
			BlockUser(gomock.Any(), "user123").
			Return(nil)

		resp, err := server.BlockUser(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "User blocked successfully", resp.Message)
	})
}

func TestServer_UnblockUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful unblock user", func(t *testing.T) {
		req := &pb.UnblockUserRequest{
			Id: "user123",
		}

		mockUserService.EXPECT().
			UnblockUser(gomock.Any(), "user123").
			Return(nil)

		resp, err := server.UnblockUser(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "User unblocked successfully", resp.Message)
	})
}

func TestServer_ChangeUserRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockUserService := mocks.NewMockUserService(ctrl)

	server := NewServer(8080, mockAuthService, mockUserService)

	t.Run("successful change user role", func(t *testing.T) {
		req := &pb.ChangeUserRoleRequest{
			Id:   "user123",
			Role: pb.UserRole_USER_ROLE_ADMIN,
		}

		mockUserService.EXPECT().
			ChangeUserRole(gomock.Any(), "user123", domain.RoleAdmin).
			Return(nil)

		resp, err := server.ChangeUserRole(context.Background(), req)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "User role changed successfully", resp.Message)
	})
}

func TestConversionHelpers(t *testing.T) {
	t.Run("convertUserToProto", func(t *testing.T) {
		user := &domain.User{
			ID:        "user123",
			Email:     "test@example.com",
			Role:      domain.RoleAdmin,
			Status:    domain.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		protoUser := convertUserToProto(user)
		assert.Equal(t, "user123", protoUser.Id)
		assert.Equal(t, "test@example.com", protoUser.Email)
		assert.Equal(t, pb.UserRole_USER_ROLE_ADMIN, protoUser.Role)
		assert.Equal(t, pb.UserStatus_USER_STATUS_ACTIVE, protoUser.Status)
	})

	t.Run("convertProtoToUserRole", func(t *testing.T) {
		tests := []struct {
			proto    pb.UserRole
			expected domain.UserRole
		}{
			{pb.UserRole_USER_ROLE_USER, domain.RoleUser},
			{pb.UserRole_USER_ROLE_MODERATOR, domain.RoleModerator},
			{pb.UserRole_USER_ROLE_ADMIN, domain.RoleAdmin},
			{pb.UserRole_USER_ROLE_UNSPECIFIED, domain.RoleUser},
		}

		for _, tt := range tests {
			result := convertProtoToUserRole(tt.proto)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("convertProtoToUserStatus", func(t *testing.T) {
		tests := []struct {
			proto    pb.UserStatus
			expected domain.UserStatus
		}{
			{pb.UserStatus_USER_STATUS_ACTIVE, domain.StatusActive},
			{pb.UserStatus_USER_STATUS_INACTIVE, domain.StatusInactive},
			{pb.UserStatus_USER_STATUS_BLOCKED, domain.StatusBlocked},
			{pb.UserStatus_USER_STATUS_UNSPECIFIED, domain.StatusActive},
		}

		for _, tt := range tests {
			result := convertProtoToUserStatus(tt.proto)
			assert.Equal(t, tt.expected, result)
		}
	})
}
