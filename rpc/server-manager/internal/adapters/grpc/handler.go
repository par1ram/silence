package grpc

import (
	"context"

	"github.com/par1ram/silence/rpc/server-manager/api/proto"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ServerManagerHandler gRPC обработчик для server-manager сервиса
type ServerManagerHandler struct {
	proto.UnimplementedServerManagerServiceServer
	serverService ports.ServerService
	logger        *zap.Logger
}

// NewServerManagerHandler создает новый gRPC обработчик
func NewServerManagerHandler(serverService ports.ServerService, logger *zap.Logger) *ServerManagerHandler {
	return &ServerManagerHandler{
		serverService: serverService,
		logger:        logger,
	}
}

// Health проверка здоровья сервиса
func (h *ServerManagerHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	h.logger.Debug("health check requested")

	return &proto.HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// CreateServer создает новый сервер
func (h *ServerManagerHandler) CreateServer(ctx context.Context, req *proto.CreateServerRequest) (*proto.Server, error) {
	h.logger.Debug("create server requested", zap.String("name", req.Name))

	domainReq := &domain.CreateServerRequest{
		Name:   req.Name,
		Type:   h.convertServerType(req.Type),
		Region: req.Region,
		Config: req.Config,
	}

	server, err := h.serverService.CreateServer(ctx, domainReq)
	if err != nil {
		h.logger.Error("failed to create server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create server: %v", err)
	}

	return h.domainServerToProto(server), nil
}

// GetServer получает сервер по ID
func (h *ServerManagerHandler) GetServer(ctx context.Context, req *proto.GetServerRequest) (*proto.Server, error) {
	h.logger.Debug("get server requested", zap.String("id", req.Id))

	server, err := h.serverService.GetServer(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get server: %v", err)
	}

	return h.domainServerToProto(server), nil
}

// ListServers получает список серверов
func (h *ServerManagerHandler) ListServers(ctx context.Context, req *proto.ListServersRequest) (*proto.ListServersResponse, error) {
	h.logger.Debug("list servers requested")

	filtersMap := map[string]interface{}{
		"type":   h.convertServerType(req.Type),
		"region": req.Region,
		"status": h.convertServerStatus(req.Status),
		"limit":  int(req.Limit),
		"offset": int(req.Offset),
	}

	servers, err := h.serverService.ListServers(ctx, filtersMap)
	if err != nil {
		h.logger.Error("failed to list servers", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list servers: %v", err)
	}

	protoServers := make([]*proto.Server, len(servers))
	for i, server := range servers {
		protoServers[i] = h.domainServerToProto(server)
	}

	return &proto.ListServersResponse{
		Servers: protoServers,
		Total:   int32(len(servers)),
	}, nil
}

// UpdateServer обновляет сервер
func (h *ServerManagerHandler) UpdateServer(ctx context.Context, req *proto.UpdateServerRequest) (*proto.Server, error) {
	h.logger.Debug("update server requested", zap.String("id", req.Id))

	domainReq := &domain.UpdateServerRequest{
		Name:   req.Name,
		Config: req.Config,
	}

	server, err := h.serverService.UpdateServer(ctx, req.Id, domainReq)
	if err != nil {
		h.logger.Error("failed to update server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update server: %v", err)
	}

	return h.domainServerToProto(server), nil
}

// DeleteServer удаляет сервер
func (h *ServerManagerHandler) DeleteServer(ctx context.Context, req *proto.DeleteServerRequest) (*proto.DeleteServerResponse, error) {
	h.logger.Debug("delete server requested", zap.String("id", req.Id))

	err := h.serverService.DeleteServer(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to delete server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete server: %v", err)
	}

	return &proto.DeleteServerResponse{
		Success: true,
	}, nil
}

// StartServer запускает сервер
func (h *ServerManagerHandler) StartServer(ctx context.Context, req *proto.StartServerRequest) (*proto.StartServerResponse, error) {
	h.logger.Debug("start server requested", zap.String("id", req.Id))

	err := h.serverService.StartServer(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to start server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to start server: %v", err)
	}

	return &proto.StartServerResponse{
		Success: true,
		Message: "Server started successfully",
	}, nil
}

// StopServer останавливает сервер
func (h *ServerManagerHandler) StopServer(ctx context.Context, req *proto.StopServerRequest) (*proto.StopServerResponse, error) {
	h.logger.Debug("stop server requested", zap.String("id", req.Id))

	err := h.serverService.StopServer(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to stop server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stop server: %v", err)
	}

	return &proto.StopServerResponse{
		Success: true,
		Message: "Server stopped successfully",
	}, nil
}

// RestartServer перезапускает сервер
func (h *ServerManagerHandler) RestartServer(ctx context.Context, req *proto.RestartServerRequest) (*proto.RestartServerResponse, error) {
	h.logger.Debug("restart server requested", zap.String("id", req.Id))

	err := h.serverService.RestartServer(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to restart server", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to restart server: %v", err)
	}

	return &proto.RestartServerResponse{
		Success: true,
		Message: "Server restarted successfully",
	}, nil
}

// GetServerStats получает статистику сервера
func (h *ServerManagerHandler) GetServerStats(ctx context.Context, req *proto.GetServerStatsRequest) (*proto.ServerStats, error) {
	h.logger.Debug("get server stats requested", zap.String("id", req.Id))

	stats, err := h.serverService.GetServerStats(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get server stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get server stats: %v", err)
	}

	return &proto.ServerStats{
		ServerId:     stats.ServerID,
		CpuUsage:     stats.CPUUsage,
		MemoryUsage:  stats.MemoryUsage,
		DiskUsage:    stats.StorageUsage,
		NetworkUsage: 0.0, // Заглушка
		Connections:  int32(stats.RequestCount),
		Uptime:       stats.Uptime,
		Timestamp:    timestamppb.New(stats.Timestamp),
	}, nil
}

// GetServerHealth получает состояние здоровья сервера
func (h *ServerManagerHandler) GetServerHealth(ctx context.Context, req *proto.GetServerHealthRequest) (*proto.ServerHealth, error) {
	h.logger.Debug("get server health requested", zap.String("id", req.Id))

	health, err := h.serverService.GetServerHealth(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get server health", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get server health: %v", err)
	}

	return &proto.ServerHealth{
		ServerId:  health.ServerID,
		Status:    string(health.Status),
		Message:   health.Message,
		Checks:    h.convertHealthChecksToProto(health.Checks),
		Timestamp: timestamppb.New(health.LastCheckAt),
	}, nil
}

// MonitorServer мониторинг сервера (stream)
func (h *ServerManagerHandler) MonitorServer(req *proto.MonitorServerRequest, stream proto.ServerManagerService_MonitorServerServer) error {
	h.logger.Debug("monitor server requested", zap.String("id", req.Id))

	// Заглушка для мониторинга - отправляем тестовое событие
	ctx := stream.Context()

	// Создаем тестовое событие
	testEvent := &proto.ServerMonitorEvent{
		ServerId:  req.Id,
		EventType: "info",
		Timestamp: timestamppb.Now(),
	}

	if err := stream.Send(testEvent); err != nil {
		h.logger.Error("failed to send monitor event", zap.Error(err))
		return err
	}

	// Ждем завершения контекста
	<-ctx.Done()
	return ctx.Err()
}

// GetServersByType получает серверы по типу
func (h *ServerManagerHandler) GetServersByType(ctx context.Context, req *proto.GetServersByTypeRequest) (*proto.GetServersByTypeResponse, error) {
	h.logger.Debug("get servers by type requested", zap.String("type", req.Type.String()))

	filtersMap := map[string]interface{}{
		"type": h.convertServerType(req.Type),
	}

	servers, err := h.serverService.ListServers(ctx, filtersMap)
	if err != nil {
		h.logger.Error("failed to get servers by type", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get servers by type: %v", err)
	}

	protoServers := make([]*proto.Server, len(servers))
	for i, server := range servers {
		protoServers[i] = h.domainServerToProto(server)
	}

	return &proto.GetServersByTypeResponse{
		Servers: protoServers,
	}, nil
}

// GetServersByRegion получает серверы по региону
func (h *ServerManagerHandler) GetServersByRegion(ctx context.Context, req *proto.GetServersByRegionRequest) (*proto.GetServersByRegionResponse, error) {
	h.logger.Debug("get servers by region requested", zap.String("region", req.Region))

	filtersMap := map[string]interface{}{
		"region": req.Region,
	}

	servers, err := h.serverService.ListServers(ctx, filtersMap)
	if err != nil {
		h.logger.Error("failed to get servers by region", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get servers by region: %v", err)
	}

	protoServers := make([]*proto.Server, len(servers))
	for i, server := range servers {
		protoServers[i] = h.domainServerToProto(server)
	}

	return &proto.GetServersByRegionResponse{
		Servers: protoServers,
	}, nil
}

// GetServersByStatus получает серверы по статусу
func (h *ServerManagerHandler) GetServersByStatus(ctx context.Context, req *proto.GetServersByStatusRequest) (*proto.GetServersByStatusResponse, error) {
	h.logger.Debug("get servers by status requested", zap.String("status", req.Status.String()))

	filtersMap := map[string]interface{}{
		"status": h.convertServerStatus(req.Status),
	}

	servers, err := h.serverService.ListServers(ctx, filtersMap)
	if err != nil {
		h.logger.Error("failed to get servers by status", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get servers by status: %v", err)
	}

	protoServers := make([]*proto.Server, len(servers))
	for i, server := range servers {
		protoServers[i] = h.domainServerToProto(server)
	}

	return &proto.GetServersByStatusResponse{
		Servers: protoServers,
	}, nil
}

// ScaleServer масштабирует сервер
func (h *ServerManagerHandler) ScaleServer(ctx context.Context, req *proto.ScaleServerRequest) (*proto.ScaleServerResponse, error) {
	h.logger.Debug("scale server requested", zap.String("id", req.Id))

	// Заглушка для масштабирования - просто возвращаем успех
	h.logger.Info("scale server requested", zap.String("id", req.Id))

	return &proto.ScaleServerResponse{
		Success: true,
		Message: "Server scaled successfully",
	}, nil
}

// CreateBackup создает резервную копию сервера
func (h *ServerManagerHandler) CreateBackup(ctx context.Context, req *proto.CreateBackupRequest) (*proto.CreateBackupResponse, error) {
	h.logger.Debug("create backup requested", zap.String("server_id", req.ServerId))

	// Заглушка для создания резервной копии
	err := h.serverService.CreateBackup(ctx, req.ServerId)
	if err != nil {
		h.logger.Error("failed to create backup", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create backup: %v", err)
	}

	return &proto.CreateBackupResponse{
		Success:  true,
		BackupId: "backup-" + req.ServerId,
		Message:  "Backup created successfully",
	}, nil
}

// RestoreBackup восстанавливает сервер из резервной копии
func (h *ServerManagerHandler) RestoreBackup(ctx context.Context, req *proto.RestoreBackupRequest) (*proto.RestoreBackupResponse, error) {
	h.logger.Debug("restore backup requested", zap.String("backup_id", req.BackupId))

	// Заглушка для восстановления резервной копии
	err := h.serverService.RestoreBackup(ctx, req.ServerId, req.BackupId)
	if err != nil {
		h.logger.Error("failed to restore backup", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to restore backup: %v", err)
	}

	return &proto.RestoreBackupResponse{
		Success: true,
		Message: "Backup restored successfully",
	}, nil
}

// UpdateServerSoftware обновляет ПО сервера
func (h *ServerManagerHandler) UpdateServerSoftware(ctx context.Context, req *proto.UpdateServerSoftwareRequest) (*proto.UpdateServerSoftwareResponse, error) {
	h.logger.Debug("update server software requested", zap.String("server_id", req.ServerId))

	// Заглушка для обновления ПО сервера
	updateReq := &domain.UpdateRequest{
		ServerID: req.ServerId,
		Version:  req.Version,
		Force:    req.Force,
	}

	err := h.serverService.StartUpdate(ctx, updateReq)
	if err != nil {
		h.logger.Error("failed to update server software", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update server software: %v", err)
	}

	return &proto.UpdateServerSoftwareResponse{
		Success: true,
		Message: "Server software updated successfully",
	}, nil
}

// Helper methods for type conversion

func (h *ServerManagerHandler) convertServerType(protoType proto.ServerType) domain.ServerType {
	switch protoType {
	case proto.ServerType_SERVER_TYPE_VPN:
		return domain.ServerTypeVPN
	case proto.ServerType_SERVER_TYPE_DPI:
		return domain.ServerTypeDPI
	case proto.ServerType_SERVER_TYPE_GATEWAY:
		return domain.ServerTypeGateway
	case proto.ServerType_SERVER_TYPE_ANALYTICS:
		return domain.ServerTypeAnalytics
	default:
		return domain.ServerTypeVPN
	}
}

func (h *ServerManagerHandler) convertServerStatus(protoStatus proto.ServerStatus) domain.ServerStatus {
	switch protoStatus {
	case proto.ServerStatus_SERVER_STATUS_CREATING:
		return domain.ServerStatusCreating
	case proto.ServerStatus_SERVER_STATUS_RUNNING:
		return domain.ServerStatusRunning
	case proto.ServerStatus_SERVER_STATUS_STOPPED:
		return domain.ServerStatusStopped
	case proto.ServerStatus_SERVER_STATUS_ERROR:
		return domain.ServerStatusError
	case proto.ServerStatus_SERVER_STATUS_DELETING:
		return domain.ServerStatusDeleting
	default:
		return domain.ServerStatusStopped
	}
}

func (h *ServerManagerHandler) convertServerTypeToProto(domainType domain.ServerType) proto.ServerType {
	switch domainType {
	case domain.ServerTypeVPN:
		return proto.ServerType_SERVER_TYPE_VPN
	case domain.ServerTypeDPI:
		return proto.ServerType_SERVER_TYPE_DPI
	case domain.ServerTypeGateway:
		return proto.ServerType_SERVER_TYPE_GATEWAY
	case domain.ServerTypeAnalytics:
		return proto.ServerType_SERVER_TYPE_ANALYTICS
	default:
		return proto.ServerType_SERVER_TYPE_VPN
	}
}

func (h *ServerManagerHandler) convertServerStatusToProto(domainStatus domain.ServerStatus) proto.ServerStatus {
	switch domainStatus {
	case domain.ServerStatusCreating:
		return proto.ServerStatus_SERVER_STATUS_CREATING
	case domain.ServerStatusRunning:
		return proto.ServerStatus_SERVER_STATUS_RUNNING
	case domain.ServerStatusStopped:
		return proto.ServerStatus_SERVER_STATUS_STOPPED
	case domain.ServerStatusError:
		return proto.ServerStatus_SERVER_STATUS_ERROR
	case domain.ServerStatusDeleting:
		return proto.ServerStatus_SERVER_STATUS_DELETING
	default:
		return proto.ServerStatus_SERVER_STATUS_STOPPED
	}
}

func (h *ServerManagerHandler) convertMonitorEventTypeToProto(eventType string) string {
	return eventType
}

func (h *ServerManagerHandler) domainServerToProto(server *domain.Server) *proto.Server {
	return &proto.Server{
		Id:        server.ID,
		Name:      server.Name,
		Type:      h.convertServerTypeToProto(server.Type),
		Status:    h.convertServerStatusToProto(server.Status),
		Region:    server.Region,
		Ip:        server.IP,
		Port:      int32(server.Port),
		Cpu:       server.CPU,
		Memory:    server.Memory,
		Disk:      server.Disk,
		Network:   server.Network,
		Config:    make(map[string]string),
		CreatedAt: timestamppb.New(server.CreatedAt),
		UpdatedAt: timestamppb.New(server.UpdatedAt),
	}
}

func (h *ServerManagerHandler) convertHealthChecksToProto(checks []map[string]interface{}) []*proto.HealthCheck {
	protoChecks := make([]*proto.HealthCheck, len(checks))
	for i, check := range checks {
		protoChecks[i] = &proto.HealthCheck{
			Name:         check["name"].(string),
			Status:       check["status"].(string),
			Message:      check["message"].(string),
			ResponseTime: 0.0, // Заглушка
		}
	}
	return protoChecks
}
