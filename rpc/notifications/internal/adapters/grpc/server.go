package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/par1ram/silence/rpc/notifications/api/proto"
	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
)

// Server represents the gRPC server
type Server struct {
	pb.UnimplementedNotificationsServiceServer
	Dispatcher *services.DispatcherService
	Port       int
}

// NewServer creates a new gRPC server
func NewServer(port int, dispatcher *services.DispatcherService) *Server {
	return &Server{
		Dispatcher: dispatcher,
		Port:       port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterNotificationsServiceServer(server, s)

	log.Printf("[grpc] server listening on port %d", s.Port)
	return server.Serve(lis)
}

// Health check
func (s *Server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// DispatchNotification dispatches a notification
func (s *Server) DispatchNotification(ctx context.Context, req *pb.DispatchNotificationRequest) (*pb.DispatchNotificationResponse, error) {
	// Convert proto to domain model
	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	notification := &domain.Notification{
		ID:          req.Id,
		Type:        convertProtoToNotificationType(req.Type),
		Priority:    convertProtoToNotificationPriority(req.Priority),
		Title:       req.Title,
		Message:     req.Message,
		Data:        data,
		Channels:    convertProtoToNotificationChannels(req.Channels),
		Recipients:  req.Recipients,
		Source:      req.Source,
		SourceID:    req.SourceId,
		Status:      domain.NotificationStatusPending,
		MaxAttempts: int(req.MaxAttempts),
		CreatedAt:   time.Now(),
	}

	// Generate ID if not provided
	if notification.ID == "" {
		notification.ID = time.Now().Format("20060102T150405.000000000")
	}

	// Set default max attempts if not provided
	if notification.MaxAttempts == 0 {
		notification.MaxAttempts = 3
	}

	// Set scheduled time if provided
	if req.ScheduledAt != nil {
		scheduledAt := req.ScheduledAt.AsTime()
		notification.ScheduledAt = &scheduledAt
	}

	// Dispatch notification
	if err := s.Dispatcher.Dispatch(ctx, notification); err != nil {
		log.Printf("[grpc] dispatch error: %v", err)
		return &pb.DispatchNotificationResponse{
			Success: false,
			Message: fmt.Sprintf("dispatch error: %v", err),
		}, nil
	}

	return &pb.DispatchNotificationResponse{
		Success:        true,
		Message:        "notification dispatched successfully",
		NotificationId: notification.ID,
	}, nil
}

// GetNotification retrieves a notification by ID
func (s *Server) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.Notification, error) {
	// This is a stub implementation - in a real scenario, you'd retrieve from storage
	return nil, status.Error(codes.Unimplemented, "GetNotification not implemented")
}

// ListNotifications lists notifications based on filters
func (s *Server) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	// This is a stub implementation - in a real scenario, you'd retrieve from storage
	return nil, status.Error(codes.Unimplemented, "ListNotifications not implemented")
}

// UpdateNotificationStatus updates the status of a notification
func (s *Server) UpdateNotificationStatus(ctx context.Context, req *pb.UpdateNotificationStatusRequest) (*pb.Notification, error) {
	// This is a stub implementation - in a real scenario, you'd update in storage
	return nil, status.Error(codes.Unimplemented, "UpdateNotificationStatus not implemented")
}

// CreateTemplate creates a new notification template
func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.NotificationTemplate, error) {
	// This is a stub implementation - in a real scenario, you'd store in database
	return nil, status.Error(codes.Unimplemented, "CreateTemplate not implemented")
}

// GetTemplate retrieves a template by ID
func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.NotificationTemplate, error) {
	// This is a stub implementation - in a real scenario, you'd retrieve from storage
	return nil, status.Error(codes.Unimplemented, "GetTemplate not implemented")
}

// ListTemplates lists notification templates
func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	// This is a stub implementation - in a real scenario, you'd retrieve from storage
	return nil, status.Error(codes.Unimplemented, "ListTemplates not implemented")
}

// UpdateTemplate updates a notification template
func (s *Server) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.NotificationTemplate, error) {
	// This is a stub implementation - in a real scenario, you'd update in storage
	return nil, status.Error(codes.Unimplemented, "UpdateTemplate not implemented")
}

// DeleteTemplate deletes a notification template
func (s *Server) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*pb.DeleteTemplateResponse, error) {
	// This is a stub implementation - in a real scenario, you'd delete from storage
	return nil, status.Error(codes.Unimplemented, "DeleteTemplate not implemented")
}

// GetUserPreferences retrieves user notification preferences
func (s *Server) GetUserPreferences(ctx context.Context, req *pb.GetUserPreferencesRequest) (*pb.NotificationPreference, error) {
	// This is a stub implementation - in a real scenario, you'd retrieve from storage
	return nil, status.Error(codes.Unimplemented, "GetUserPreferences not implemented")
}

// UpdateUserPreferences updates user notification preferences
func (s *Server) UpdateUserPreferences(ctx context.Context, req *pb.UpdateUserPreferencesRequest) (*pb.NotificationPreference, error) {
	// This is a stub implementation - in a real scenario, you'd update in storage
	return nil, status.Error(codes.Unimplemented, "UpdateUserPreferences not implemented")
}

// GetNotificationStats retrieves notification statistics
func (s *Server) GetNotificationStats(ctx context.Context, req *pb.GetNotificationStatsRequest) (*pb.NotificationStats, error) {
	// This is a stub implementation - in a real scenario, you'd calculate from storage
	return nil, status.Error(codes.Unimplemented, "GetNotificationStats not implemented")
}

// Helper functions for conversion between proto and domain models

func convertProtoToNotificationType(protoType pb.NotificationType) domain.NotificationType {
	switch protoType {
	case pb.NotificationType_NOTIFICATION_TYPE_SYSTEM_ALERT:
		return domain.NotificationTypeSystemAlert
	case pb.NotificationType_NOTIFICATION_TYPE_SERVER_DOWN:
		return domain.NotificationTypeServerDown
	case pb.NotificationType_NOTIFICATION_TYPE_SERVER_UP:
		return domain.NotificationTypeServerUp
	case pb.NotificationType_NOTIFICATION_TYPE_HIGH_LOAD:
		return domain.NotificationTypeHighLoad
	case pb.NotificationType_NOTIFICATION_TYPE_LOW_DISK_SPACE:
		return domain.NotificationTypeLowDiskSpace
	case pb.NotificationType_NOTIFICATION_TYPE_BACKUP_FAILED:
		return domain.NotificationTypeBackupFailed
	case pb.NotificationType_NOTIFICATION_TYPE_BACKUP_SUCCESS:
		return domain.NotificationTypeBackupSuccess
	case pb.NotificationType_NOTIFICATION_TYPE_UPDATE_FAILED:
		return domain.NotificationTypeUpdateFailed
	case pb.NotificationType_NOTIFICATION_TYPE_UPDATE_SUCCESS:
		return domain.NotificationTypeUpdateSuccess
	case pb.NotificationType_NOTIFICATION_TYPE_USER_LOGIN:
		return domain.NotificationTypeUserLogin
	case pb.NotificationType_NOTIFICATION_TYPE_USER_LOGOUT:
		return domain.NotificationTypeUserLogout
	case pb.NotificationType_NOTIFICATION_TYPE_USER_REGISTERED:
		return domain.NotificationTypeUserRegistered
	case pb.NotificationType_NOTIFICATION_TYPE_USER_BLOCKED:
		return domain.NotificationTypeUserBlocked
	case pb.NotificationType_NOTIFICATION_TYPE_USER_UNBLOCKED:
		return domain.NotificationTypeUserUnblocked
	case pb.NotificationType_NOTIFICATION_TYPE_PASSWORD_RESET:
		return domain.NotificationTypePasswordReset
	case pb.NotificationType_NOTIFICATION_TYPE_SUBSCRIPTION_EXPIRED:
		return domain.NotificationTypeSubscriptionExpired
	case pb.NotificationType_NOTIFICATION_TYPE_SUBSCRIPTION_RENEWED:
		return domain.NotificationTypeSubscriptionRenewed
	case pb.NotificationType_NOTIFICATION_TYPE_VPN_CONNECTED:
		return domain.NotificationTypeVPNConnected
	case pb.NotificationType_NOTIFICATION_TYPE_VPN_DISCONNECTED:
		return domain.NotificationTypeVPNDisconnected
	case pb.NotificationType_NOTIFICATION_TYPE_VPN_ERROR:
		return domain.NotificationTypeVPNError
	case pb.NotificationType_NOTIFICATION_TYPE_BYPASS_BLOCKED:
		return domain.NotificationTypeBypassBlocked
	case pb.NotificationType_NOTIFICATION_TYPE_BYPASS_SUCCESS:
		return domain.NotificationTypeBypassSuccess
	case pb.NotificationType_NOTIFICATION_TYPE_METRICS_ALERT:
		return domain.NotificationTypeMetricsAlert
	case pb.NotificationType_NOTIFICATION_TYPE_ANOMALY_DETECTED:
		return domain.NotificationTypeAnomalyDetected
	case pb.NotificationType_NOTIFICATION_TYPE_THRESHOLD_EXCEEDED:
		return domain.NotificationTypeThresholdExceeded
	default:
		return domain.NotificationTypeSystemAlert
	}
}

func convertProtoToNotificationPriority(protoPriority pb.NotificationPriority) domain.NotificationPriority {
	switch protoPriority {
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_LOW:
		return domain.NotificationPriorityLow
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_NORMAL:
		return domain.NotificationPriorityNormal
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_HIGH:
		return domain.NotificationPriorityHigh
	case pb.NotificationPriority_NOTIFICATION_PRIORITY_URGENT:
		return domain.NotificationPriorityUrgent
	default:
		return domain.NotificationPriorityNormal
	}
}

func convertProtoToNotificationChannels(protoChannels []pb.NotificationChannel) []domain.NotificationChannel {
	channels := make([]domain.NotificationChannel, len(protoChannels))
	for i, protoChannel := range protoChannels {
		switch protoChannel {
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_EMAIL:
			channels[i] = domain.NotificationChannelEmail
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_SMS:
			channels[i] = domain.NotificationChannelSMS
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_PUSH:
			channels[i] = domain.NotificationChannelPush
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_TELEGRAM:
			channels[i] = domain.NotificationChannelTelegram
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_WEBHOOK:
			channels[i] = domain.NotificationChannelWebhook
		case pb.NotificationChannel_NOTIFICATION_CHANNEL_SLACK:
			channels[i] = domain.NotificationChannelSlack
		default:
			channels[i] = domain.NotificationChannelEmail
		}
	}
	return channels
}
