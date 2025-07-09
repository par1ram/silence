package auth

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/par1ram/silence/api/gateway/api/proto/auth"
)

// Client gRPC клиент для auth сервиса
type Client struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

// NewClient создает новый gRPC клиент для auth сервиса
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := pb.NewAuthServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// Health проверяет здоровье сервиса
func (c *Client) Health(ctx context.Context) (*pb.HealthResponse, error) {
	return c.client.Health(ctx, &pb.HealthRequest{})
}

// Login выполняет аутентификацию пользователя
func (c *Client) Login(ctx context.Context, email, password string) (*pb.AuthResponse, error) {
	req := &pb.LoginRequest{
		Email:    email,
		Password: password,
	}
	return c.client.Login(ctx, req)
}

// Register регистрирует нового пользователя
func (c *Client) Register(ctx context.Context, email, password string) (*pb.AuthResponse, error) {
	req := &pb.RegisterRequest{
		Email:    email,
		Password: password,
	}
	return c.client.Register(ctx, req)
}

// GetMe получает информацию о текущем пользователе
func (c *Client) GetMe(ctx context.Context, token string) (*pb.User, error) {
	req := &pb.GetMeRequest{
		Token: token,
	}
	return c.client.GetMe(ctx, req)
}

// CreateUser создает нового пользователя
func (c *Client) CreateUser(ctx context.Context, email, password string, role pb.UserRole) (*pb.User, error) {
	req := &pb.CreateUserRequest{
		Email:    email,
		Password: password,
		Role:     role,
	}
	return c.client.CreateUser(ctx, req)
}

// GetUser получает пользователя по ID
func (c *Client) GetUser(ctx context.Context, id string) (*pb.User, error) {
	req := &pb.GetUserRequest{
		Id: id,
	}
	return c.client.GetUser(ctx, req)
}

// UpdateUser обновляет пользователя
func (c *Client) UpdateUser(ctx context.Context, id, email string, role pb.UserRole, status pb.UserStatus) (*pb.User, error) {
	req := &pb.UpdateUserRequest{
		Id:     id,
		Email:  email,
		Role:   role,
		Status: status,
	}
	return c.client.UpdateUser(ctx, req)
}

// DeleteUser удаляет пользователя
func (c *Client) DeleteUser(ctx context.Context, id string) (*pb.DeleteUserResponse, error) {
	req := &pb.DeleteUserRequest{
		Id: id,
	}
	return c.client.DeleteUser(ctx, req)
}

// ListUsers получает список пользователей
func (c *Client) ListUsers(ctx context.Context, role pb.UserRole, status pb.UserStatus, email string, limit, offset int32) (*pb.ListUsersResponse, error) {
	req := &pb.ListUsersRequest{
		Role:   role,
		Status: status,
		Email:  email,
		Limit:  limit,
		Offset: offset,
	}
	return c.client.ListUsers(ctx, req)
}

// BlockUser блокирует пользователя
func (c *Client) BlockUser(ctx context.Context, id string) (*pb.BlockUserResponse, error) {
	req := &pb.BlockUserRequest{
		Id: id,
	}
	return c.client.BlockUser(ctx, req)
}

// UnblockUser разблокирует пользователя
func (c *Client) UnblockUser(ctx context.Context, id string) (*pb.UnblockUserResponse, error) {
	req := &pb.UnblockUserRequest{
		Id: id,
	}
	return c.client.UnblockUser(ctx, req)
}

// ChangeUserRole изменяет роль пользователя
func (c *Client) ChangeUserRole(ctx context.Context, id string, role pb.UserRole) (*pb.ChangeUserRoleResponse, error) {
	req := &pb.ChangeUserRoleRequest{
		Id:   id,
		Role: role,
	}
	return c.client.ChangeUserRole(ctx, req)
}
