package grpc

import (
	"context"

	"github.com/par1ram/silence/rpc/dpi-bypass/api/proto"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/ports"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DPIBypassHandler gRPC обработчик для dpi-bypass сервиса
type DPIBypassHandler struct {
	proto.UnimplementedDpiBypassServiceServer
	dpiService ports.DPIBypassService
	logger     *zap.Logger
}

// NewDPIBypassHandler создает новый gRPC обработчик
func NewDPIBypassHandler(dpiService ports.DPIBypassService, logger *zap.Logger) *DPIBypassHandler {
	return &DPIBypassHandler{
		dpiService: dpiService,
		logger:     logger,
	}
}

// Health проверка здоровья сервиса
func (h *DPIBypassHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	h.logger.Debug("health check requested")

	return &proto.HealthResponse{
		Status:    "healthy",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// CreateBypassConfig создает конфигурацию обхода
func (h *DPIBypassHandler) CreateBypassConfig(ctx context.Context, req *proto.CreateBypassConfigRequest) (*proto.BypassConfig, error) {
	h.logger.Debug("create bypass config requested", zap.String("name", req.Name))

	domainConfig := &domain.CreateBypassConfigRequest{
		Name:        req.Name,
		Description: req.Description,
		Type:        h.convertBypassType(req.Type),
		Method:      h.convertBypassMethod(req.Method),
		Parameters:  req.Parameters,
	}

	config, err := h.dpiService.CreateBypassConfig(ctx, domainConfig)
	if err != nil {
		h.logger.Error("failed to create bypass config", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create bypass config: %v", err)
	}

	return h.domainConfigToProto(config), nil
}

// GetBypassConfig получает конфигурацию обхода
func (h *DPIBypassHandler) GetBypassConfig(ctx context.Context, req *proto.GetBypassConfigRequest) (*proto.BypassConfig, error) {
	h.logger.Debug("get bypass config requested", zap.String("id", req.Id))

	config, err := h.dpiService.GetBypassConfig(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to get bypass config", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get bypass config: %v", err)
	}

	return h.domainConfigToProto(config), nil
}

// ListBypassConfigs получает список конфигураций обхода
func (h *DPIBypassHandler) ListBypassConfigs(ctx context.Context, req *proto.ListBypassConfigsRequest) (*proto.ListBypassConfigsResponse, error) {
	h.logger.Debug("list bypass configs requested")

	filters := &domain.BypassConfigFilters{
		Type:   h.convertBypassType(req.Type),
		Status: h.convertBypassStatus(req.Status),
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	configs, total, err := h.dpiService.ListBypassConfigs(ctx, filters)
	if err != nil {
		h.logger.Error("failed to list bypass configs", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list bypass configs: %v", err)
	}

	protoConfigs := make([]*proto.BypassConfig, len(configs))
	for i, config := range configs {
		protoConfigs[i] = h.domainConfigToProto(config)
	}

	return &proto.ListBypassConfigsResponse{
		Configs: protoConfigs,
		Total:   int32(total),
	}, nil
}

// UpdateBypassConfig обновляет конфигурацию обхода
func (h *DPIBypassHandler) UpdateBypassConfig(ctx context.Context, req *proto.UpdateBypassConfigRequest) (*proto.BypassConfig, error) {
	h.logger.Debug("update bypass config requested", zap.String("id", req.Id))

	domainReq := &domain.UpdateBypassConfigRequest{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Type:        h.convertBypassType(req.Type),
		Method:      h.convertBypassMethod(req.Method),
		Parameters:  req.Parameters,
	}

	config, err := h.dpiService.UpdateBypassConfig(ctx, domainReq)
	if err != nil {
		h.logger.Error("failed to update bypass config", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update bypass config: %v", err)
	}

	return h.domainConfigToProto(config), nil
}

// DeleteBypassConfig удаляет конфигурацию обхода
func (h *DPIBypassHandler) DeleteBypassConfig(ctx context.Context, req *proto.DeleteBypassConfigRequest) (*proto.DeleteBypassConfigResponse, error) {
	h.logger.Debug("delete bypass config requested", zap.String("id", req.Id))

	err := h.dpiService.DeleteBypassConfig(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to delete bypass config", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete bypass config: %v", err)
	}

	return &proto.DeleteBypassConfigResponse{
		Success: true,
	}, nil
}

// StartBypass запускает обход
func (h *DPIBypassHandler) StartBypass(ctx context.Context, req *proto.StartBypassRequest) (*proto.StartBypassResponse, error) {
	h.logger.Debug("start bypass requested", zap.String("config_id", req.ConfigId))

	domainReq := &domain.StartBypassRequest{
		ConfigID:   req.ConfigId,
		TargetHost: req.TargetHost,
		TargetPort: int(req.TargetPort),
		Options:    req.Options,
	}

	session, err := h.dpiService.StartBypass(ctx, domainReq)
	if err != nil {
		h.logger.Error("failed to start bypass", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to start bypass: %v", err)
	}

	return &proto.StartBypassResponse{
		Success:   true,
		SessionId: session.ID,
		Message:   "Bypass started successfully",
	}, nil
}

// StopBypass останавливает обход
func (h *DPIBypassHandler) StopBypass(ctx context.Context, req *proto.StopBypassRequest) (*proto.StopBypassResponse, error) {
	h.logger.Debug("stop bypass requested", zap.String("session_id", req.SessionId))

	err := h.dpiService.StopBypass(ctx, req.SessionId)
	if err != nil {
		h.logger.Error("failed to stop bypass", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stop bypass: %v", err)
	}

	return &proto.StopBypassResponse{
		Success: true,
		Message: "Bypass stopped successfully",
	}, nil
}

// GetBypassStatus получает статус обхода
func (h *DPIBypassHandler) GetBypassStatus(ctx context.Context, req *proto.GetBypassStatusRequest) (*proto.GetBypassStatusResponse, error) {
	h.logger.Debug("get bypass status requested", zap.String("session_id", req.SessionId))

	bypassStatus, err := h.dpiService.GetBypassStatus(ctx, req.SessionId)
	if err != nil {
		h.logger.Error("failed to get bypass status", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get bypass status: %v", err)
	}

	return &proto.GetBypassStatusResponse{
		SessionId:       bypassStatus.SessionID,
		Status:          h.convertBypassStatusToProto(bypassStatus.Status),
		ConfigId:        bypassStatus.ConfigID,
		TargetHost:      bypassStatus.TargetHost,
		TargetPort:      int32(bypassStatus.TargetPort),
		StartedAt:       timestamppb.New(bypassStatus.StartedAt),
		DurationSeconds: bypassStatus.DurationSeconds,
		Message:         bypassStatus.Message,
	}, nil
}

// GetBypassStats получает статистику обхода
func (h *DPIBypassHandler) GetBypassStats(ctx context.Context, req *proto.GetBypassStatsRequest) (*proto.BypassStats, error) {
	h.logger.Debug("get bypass stats requested", zap.String("session_id", req.SessionId))

	stats, err := h.dpiService.GetBypassStats(ctx, req.SessionId)
	if err != nil {
		h.logger.Error("failed to get bypass stats", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get bypass stats: %v", err)
	}

	return &proto.BypassStats{
		Id:                     stats.ID,
		ConfigId:               stats.ConfigID,
		SessionId:              stats.SessionID,
		BytesSent:              stats.BytesSent,
		BytesReceived:          stats.BytesReceived,
		PacketsSent:            stats.PacketsSent,
		PacketsReceived:        stats.PacketsReceived,
		ConnectionsEstablished: stats.ConnectionsEstablished,
		ConnectionsFailed:      stats.ConnectionsFailed,
		SuccessRate:            stats.SuccessRate,
		AverageLatency:         stats.AverageLatency,
		StartTime:              timestamppb.New(stats.StartTime),
		EndTime:                timestamppb.New(stats.EndTime),
	}, nil
}

// GetBypassHistory получает историю обхода
func (h *DPIBypassHandler) GetBypassHistory(ctx context.Context, req *proto.GetBypassHistoryRequest) (*proto.GetBypassHistoryResponse, error) {
	h.logger.Debug("get bypass history requested", zap.String("config_id", req.ConfigId))

	historyReq := &domain.BypassHistoryRequest{
		ConfigID:  req.ConfigId,
		StartTime: req.StartTime.AsTime(),
		EndTime:   req.EndTime.AsTime(),
		Limit:     int(req.Limit),
		Offset:    int(req.Offset),
	}

	entries, total, err := h.dpiService.GetBypassHistory(ctx, historyReq)
	if err != nil {
		h.logger.Error("failed to get bypass history", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get bypass history: %v", err)
	}

	protoEntries := make([]*proto.BypassHistoryEntry, len(entries))
	for i, entry := range entries {
		protoEntries[i] = &proto.BypassHistoryEntry{
			Id:               entry.ID,
			ConfigId:         entry.ConfigID,
			SessionId:        entry.SessionID,
			TargetHost:       entry.TargetHost,
			TargetPort:       int32(entry.TargetPort),
			Status:           h.convertBypassStatusToProto(entry.Status),
			StartedAt:        timestamppb.New(entry.StartedAt),
			EndedAt:          timestamppb.New(entry.EndedAt),
			DurationSeconds:  entry.DurationSeconds,
			BytesTransferred: entry.BytesTransferred,
			ErrorMessage:     entry.ErrorMessage,
		}
	}

	return &proto.GetBypassHistoryResponse{
		Entries: protoEntries,
		Total:   int32(total),
	}, nil
}

// AddBypassRule добавляет правило обхода
func (h *DPIBypassHandler) AddBypassRule(ctx context.Context, req *proto.AddBypassRuleRequest) (*proto.BypassRule, error) {
	h.logger.Debug("add bypass rule requested", zap.String("config_id", req.ConfigId))

	domainReq := &domain.AddBypassRuleRequest{
		ConfigID:   req.ConfigId,
		Name:       req.Name,
		Type:       h.convertRuleType(req.Type),
		Action:     h.convertRuleAction(req.Action),
		Pattern:    req.Pattern,
		Parameters: req.Parameters,
		Priority:   int(req.Priority),
	}

	rule, err := h.dpiService.AddBypassRule(ctx, domainReq)
	if err != nil {
		h.logger.Error("failed to add bypass rule", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to add bypass rule: %v", err)
	}

	return h.domainRuleToProto(rule), nil
}

// UpdateBypassRule обновляет правило обхода
func (h *DPIBypassHandler) UpdateBypassRule(ctx context.Context, req *proto.UpdateBypassRuleRequest) (*proto.BypassRule, error) {
	h.logger.Debug("update bypass rule requested", zap.String("id", req.Id))

	domainReq := &domain.UpdateBypassRuleRequest{
		ID:         req.Id,
		Name:       req.Name,
		Type:       h.convertRuleType(req.Type),
		Action:     h.convertRuleAction(req.Action),
		Pattern:    req.Pattern,
		Parameters: req.Parameters,
		Priority:   int(req.Priority),
		Enabled:    req.Enabled,
	}

	rule, err := h.dpiService.UpdateBypassRule(ctx, domainReq)
	if err != nil {
		h.logger.Error("failed to update bypass rule", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update bypass rule: %v", err)
	}

	return h.domainRuleToProto(rule), nil
}

// DeleteBypassRule удаляет правило обхода
func (h *DPIBypassHandler) DeleteBypassRule(ctx context.Context, req *proto.DeleteBypassRuleRequest) (*proto.DeleteBypassRuleResponse, error) {
	h.logger.Debug("delete bypass rule requested", zap.String("id", req.Id))

	err := h.dpiService.DeleteBypassRule(ctx, req.Id)
	if err != nil {
		h.logger.Error("failed to delete bypass rule", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete bypass rule: %v", err)
	}

	return &proto.DeleteBypassRuleResponse{
		Success: true,
	}, nil
}

// ListBypassRules получает список правил обхода
func (h *DPIBypassHandler) ListBypassRules(ctx context.Context, req *proto.ListBypassRulesRequest) (*proto.ListBypassRulesResponse, error) {
	h.logger.Debug("list bypass rules requested", zap.String("config_id", req.ConfigId))

	filters := &domain.BypassRuleFilters{
		ConfigID: req.ConfigId,
		Type:     h.convertRuleType(req.Type),
		Enabled:  req.Enabled,
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
	}

	rules, total, err := h.dpiService.ListBypassRules(ctx, filters)
	if err != nil {
		h.logger.Error("failed to list bypass rules", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list bypass rules: %v", err)
	}

	protoRules := make([]*proto.BypassRule, len(rules))
	for i, rule := range rules {
		protoRules[i] = h.domainRuleToProto(rule)
	}

	return &proto.ListBypassRulesResponse{
		Rules: protoRules,
		Total: int32(total),
	}, nil
}

// Helper methods for conversions

func (h *DPIBypassHandler) convertBypassType(protoType proto.BypassType) domain.BypassType {
	switch protoType {
	case proto.BypassType_BYPASS_TYPE_DOMAIN_FRONTING:
		return domain.BypassTypeDomainFronting
	case proto.BypassType_BYPASS_TYPE_SNI_MASKING:
		return domain.BypassTypeSNIMasking
	case proto.BypassType_BYPASS_TYPE_PACKET_FRAGMENTATION:
		return domain.BypassTypePacketFragmentation
	case proto.BypassType_BYPASS_TYPE_PROTOCOL_OBFUSCATION:
		return domain.BypassTypeProtocolObfuscation
	case proto.BypassType_BYPASS_TYPE_TUNNEL_OBFUSCATION:
		return domain.BypassTypeTunnelObfuscation
	default:
		return domain.BypassType("")
	}
}

func (h *DPIBypassHandler) convertBypassMethod(protoMethod proto.BypassMethod) domain.BypassMethod {
	switch protoMethod {
	case proto.BypassMethod_BYPASS_METHOD_HTTP_HEADER:
		return domain.BypassMethodHTTPHeader
	case proto.BypassMethod_BYPASS_METHOD_TLS_HANDSHAKE:
		return domain.BypassMethodTLSHandshake
	case proto.BypassMethod_BYPASS_METHOD_TCP_FRAGMENT:
		return domain.BypassMethodTCPFragment
	case proto.BypassMethod_BYPASS_METHOD_UDP_FRAGMENT:
		return domain.BypassMethodUDPFragment
	case proto.BypassMethod_BYPASS_METHOD_PROXY_CHAIN:
		return domain.BypassMethodProxyChain
	default:
		return domain.BypassMethod("")
	}
}

func (h *DPIBypassHandler) convertBypassStatus(protoStatus proto.BypassStatus) domain.BypassStatus {
	switch protoStatus {
	case proto.BypassStatus_BYPASS_STATUS_INACTIVE:
		return domain.BypassStatusInactive
	case proto.BypassStatus_BYPASS_STATUS_ACTIVE:
		return domain.BypassStatusActive
	case proto.BypassStatus_BYPASS_STATUS_ERROR:
		return domain.BypassStatusError
	case proto.BypassStatus_BYPASS_STATUS_TESTING:
		return domain.BypassStatusTesting
	default:
		return domain.BypassStatus("")
	}
}

func (h *DPIBypassHandler) convertBypassStatusToProto(domainStatus domain.BypassStatus) proto.BypassStatus {
	switch domainStatus {
	case domain.BypassStatusInactive:
		return proto.BypassStatus_BYPASS_STATUS_INACTIVE
	case domain.BypassStatusActive:
		return proto.BypassStatus_BYPASS_STATUS_ACTIVE
	case domain.BypassStatusError:
		return proto.BypassStatus_BYPASS_STATUS_ERROR
	case domain.BypassStatusTesting:
		return proto.BypassStatus_BYPASS_STATUS_TESTING
	default:
		return proto.BypassStatus_BYPASS_STATUS_UNSPECIFIED
	}
}

func (h *DPIBypassHandler) convertRuleType(protoType proto.RuleType) domain.RuleType {
	switch protoType {
	case proto.RuleType_RULE_TYPE_DOMAIN:
		return domain.RuleTypeDomain
	case proto.RuleType_RULE_TYPE_IP:
		return domain.RuleTypeIP
	case proto.RuleType_RULE_TYPE_PORT:
		return domain.RuleTypePort
	case proto.RuleType_RULE_TYPE_PROTOCOL:
		return domain.RuleTypeProtocol
	case proto.RuleType_RULE_TYPE_REGEX:
		return domain.RuleTypeRegex
	default:
		return domain.RuleType("")
	}
}

func (h *DPIBypassHandler) convertRuleAction(protoAction proto.RuleAction) domain.RuleAction {
	switch protoAction {
	case proto.RuleAction_RULE_ACTION_ALLOW:
		return domain.RuleActionAllow
	case proto.RuleAction_RULE_ACTION_BLOCK:
		return domain.RuleActionBlock
	case proto.RuleAction_RULE_ACTION_BYPASS:
		return domain.RuleActionBypass
	case proto.RuleAction_RULE_ACTION_FRAGMENT:
		return domain.RuleActionFragment
	case proto.RuleAction_RULE_ACTION_OBFUSCATE:
		return domain.RuleActionObfuscate
	default:
		return domain.RuleAction("")
	}
}

func (h *DPIBypassHandler) domainConfigToProto(config *domain.BypassConfig) *proto.BypassConfig {
	protoRules := make([]*proto.BypassRule, len(config.Rules))
	for i, rule := range config.Rules {
		protoRules[i] = h.domainRuleToProto(rule)
	}

	return &proto.BypassConfig{
		Id:          config.ID,
		Name:        config.Name,
		Description: config.Description,
		Type:        h.convertBypassTypeToProto(config.Type),
		Method:      h.convertBypassMethodToProto(config.Method),
		Status:      h.convertBypassStatusToProto(config.Status),
		Parameters:  config.Parameters,
		Rules:       protoRules,
		CreatedAt:   timestamppb.New(config.CreatedAt),
		UpdatedAt:   timestamppb.New(config.UpdatedAt),
	}
}

func (h *DPIBypassHandler) domainRuleToProto(rule *domain.BypassRule) *proto.BypassRule {
	return &proto.BypassRule{
		Id:         rule.ID,
		ConfigId:   rule.ConfigID,
		Name:       rule.Name,
		Type:       h.convertRuleTypeToProto(rule.Type),
		Action:     h.convertRuleActionToProto(rule.Action),
		Pattern:    rule.Pattern,
		Parameters: rule.Parameters,
		Priority:   int32(rule.Priority),
		Enabled:    rule.Enabled,
		CreatedAt:  timestamppb.New(rule.CreatedAt),
		UpdatedAt:  timestamppb.New(rule.UpdatedAt),
	}
}

func (h *DPIBypassHandler) convertBypassTypeToProto(domainType domain.BypassType) proto.BypassType {
	switch domainType {
	case domain.BypassTypeDomainFronting:
		return proto.BypassType_BYPASS_TYPE_DOMAIN_FRONTING
	case domain.BypassTypeSNIMasking:
		return proto.BypassType_BYPASS_TYPE_SNI_MASKING
	case domain.BypassTypePacketFragmentation:
		return proto.BypassType_BYPASS_TYPE_PACKET_FRAGMENTATION
	case domain.BypassTypeProtocolObfuscation:
		return proto.BypassType_BYPASS_TYPE_PROTOCOL_OBFUSCATION
	case domain.BypassTypeTunnelObfuscation:
		return proto.BypassType_BYPASS_TYPE_TUNNEL_OBFUSCATION
	default:
		return proto.BypassType_BYPASS_TYPE_UNSPECIFIED
	}
}

func (h *DPIBypassHandler) convertBypassMethodToProto(domainMethod domain.BypassMethod) proto.BypassMethod {
	switch domainMethod {
	case domain.BypassMethodHTTPHeader:
		return proto.BypassMethod_BYPASS_METHOD_HTTP_HEADER
	case domain.BypassMethodTLSHandshake:
		return proto.BypassMethod_BYPASS_METHOD_TLS_HANDSHAKE
	case domain.BypassMethodTCPFragment:
		return proto.BypassMethod_BYPASS_METHOD_TCP_FRAGMENT
	case domain.BypassMethodUDPFragment:
		return proto.BypassMethod_BYPASS_METHOD_UDP_FRAGMENT
	case domain.BypassMethodProxyChain:
		return proto.BypassMethod_BYPASS_METHOD_PROXY_CHAIN
	default:
		return proto.BypassMethod_BYPASS_METHOD_UNSPECIFIED
	}
}

func (h *DPIBypassHandler) convertRuleTypeToProto(domainType domain.RuleType) proto.RuleType {
	switch domainType {
	case domain.RuleTypeDomain:
		return proto.RuleType_RULE_TYPE_DOMAIN
	case domain.RuleTypeIP:
		return proto.RuleType_RULE_TYPE_IP
	case domain.RuleTypePort:
		return proto.RuleType_RULE_TYPE_PORT
	case domain.RuleTypeProtocol:
		return proto.RuleType_RULE_TYPE_PROTOCOL
	case domain.RuleTypeRegex:
		return proto.RuleType_RULE_TYPE_REGEX
	default:
		return proto.RuleType_RULE_TYPE_UNSPECIFIED
	}
}

func (h *DPIBypassHandler) convertRuleActionToProto(domainAction domain.RuleAction) proto.RuleAction {
	switch domainAction {
	case domain.RuleActionAllow:
		return proto.RuleAction_RULE_ACTION_ALLOW
	case domain.RuleActionBlock:
		return proto.RuleAction_RULE_ACTION_BLOCK
	case domain.RuleActionBypass:
		return proto.RuleAction_RULE_ACTION_BYPASS
	case domain.RuleActionFragment:
		return proto.RuleAction_RULE_ACTION_FRAGMENT
	case domain.RuleActionObfuscate:
		return proto.RuleAction_RULE_ACTION_OBFUSCATE
	default:
		return proto.RuleAction_RULE_ACTION_UNSPECIFIED
	}
}
