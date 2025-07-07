# New Connection Endpoints Implementation Plan

## Overview

This document outlines the implementation plan for adding specialized connection endpoints to the Silence VPN Gateway Service to improve user experience and provide more granular control over VPN connections.

## Current State

### Existing Endpoint
- `POST /api/v1/connect` - Universal VPN + DPI Bypass connection

### Current Handler
Location: `silence/api/gateway/internal/adapters/http/handlers.go`
```go
func (h *Handlers) ConnectHandler(w http.ResponseWriter, r *http.Request) {
    // Creates both DPI bypass and VPN tunnel
    // Returns combined connection info
}
```

## Proposed New Endpoints

### 1. VPN-Only Connection
```
POST /api/v1/connect/vpn
```
**Purpose**: Create and start only a VPN tunnel without DPI bypass
**Use Case**: Users who only need VPN without circumvention

### 2. DPI Bypass-Only Connection
```
POST /api/v1/connect/dpi
```
**Purpose**: Create and start only DPI bypass without VPN
**Use Case**: Users who only need circumvention without VPN

### 3. Method-Specific Connections
```
POST /api/v1/connect/shadowsocks
POST /api/v1/connect/v2ray
POST /api/v1/connect/obfs4
```
**Purpose**: Direct connection using specific bypass methods
**Use Case**: Users who prefer specific circumvention methods

### 4. Connection Management
```
POST /api/v1/disconnect
GET /api/v1/connect/status
```
**Purpose**: Manage and monitor active connections
**Use Case**: Proper connection lifecycle management

## Implementation Plan

### Phase 1: Core Handler Implementation

#### 1.1 Add New Handler Functions
Add to `silence/api/gateway/internal/adapters/http/handlers.go`:

```go
// VPN-only connection handler
func (h *Handlers) ConnectVPNHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req VPNConnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Create VPN tunnel only
    vpnReq := map[string]interface{}{
        "name":          req.Name,
        "listen_port":   req.ListenPort,
        "mtu":           req.MTU,
        "auto_recovery": req.AutoRecovery,
    }
    
    vpnResp, err := h.proxyService.CreateVPNTunnel(r.Context(), vpnReq)
    if err != nil {
        h.logger.Error("failed to create VPN tunnel", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to create VPN tunnel")
        return
    }

    // Start tunnel
    if err := h.proxyService.StartVPNTunnel(r.Context(), vpnResp["id"].(string)); err != nil {
        h.logger.Error("failed to start VPN tunnel", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to start VPN tunnel")
        return
    }

    response := VPNConnectResponse{
        TunnelID:    vpnResp["id"].(string),
        TunnelName:  req.Name,
        PublicKey:   vpnResp["public_key"].(string),
        Endpoint:    vpnResp["endpoint"].(string),
        Status:      "connected",
        Config:      buildVPNConfig(vpnResp),
        CreatedAt:   time.Now(),
    }

    writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode VPN connect response")
}

// DPI-only connection handler
func (h *Handlers) ConnectDPIHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req DPIConnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Create DPI bypass only
    bypassReq := map[string]interface{}{
        "name":        req.Name,
        "method":      req.Method,
        "local_port":  req.LocalPort,
        "remote_host": req.RemoteHost,
        "remote_port": req.RemotePort,
        "password":    req.Password,
        "encryption":  req.Encryption,
    }
    
    bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
    if err != nil {
        h.logger.Error("failed to create DPI bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to create DPI bypass")
        return
    }

    // Start bypass
    if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
        h.logger.Error("failed to start DPI bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to start DPI bypass")
        return
    }

    response := DPIConnectResponse{
        BypassID:  bypassResp["id"].(string),
        Method:    req.Method,
        LocalPort: req.LocalPort,
        Status:    "running",
        Config:    buildDPIConfig(bypassResp),
        CreatedAt: time.Now(),
    }

    writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode DPI connect response")
}

// Shadowsocks-specific connection handler
func (h *Handlers) ConnectShadowsocksHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req ShadowsocksConnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Create Shadowsocks-specific bypass
    bypassReq := map[string]interface{}{
        "name":        req.Name,
        "method":      "shadowsocks",
        "local_port":  req.LocalPort,
        "remote_host": req.ServerHost,
        "remote_port": req.ServerPort,
        "password":    req.Password,
        "encryption":  req.Encryption,
        "timeout":     req.Timeout,
    }
    
    bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
    if err != nil {
        h.logger.Error("failed to create Shadowsocks bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to create Shadowsocks bypass")
        return
    }

    // Start bypass
    if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
        h.logger.Error("failed to start Shadowsocks bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to start Shadowsocks bypass")
        return
    }

    response := ShadowsocksConnectResponse{
        ConnectionID: bypassResp["id"].(string),
        LocalPort:    req.LocalPort,
        Status:       "connected",
        Config: ShadowsocksConfig{
            Server:     req.ServerHost,
            ServerPort: req.ServerPort,
            LocalPort:  req.LocalPort,
            Password:   req.Password,
            Method:     req.Encryption,
        },
        CreatedAt: time.Now(),
    }

    writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode Shadowsocks connect response")
}

// V2Ray-specific connection handler
func (h *Handlers) ConnectV2RayHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req V2RayConnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Create V2Ray-specific bypass
    bypassReq := map[string]interface{}{
        "name":        req.Name,
        "method":      "v2ray",
        "local_port":  req.LocalPort,
        "remote_host": req.ServerHost,
        "remote_port": req.ServerPort,
        "uuid":        req.UUID,
        "alter_id":    req.AlterID,
        "security":    req.Security,
        "network":     req.Network,
    }
    
    bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
    if err != nil {
        h.logger.Error("failed to create V2Ray bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to create V2Ray bypass")
        return
    }

    // Start bypass
    if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
        h.logger.Error("failed to start V2Ray bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to start V2Ray bypass")
        return
    }

    response := V2RayConnectResponse{
        ConnectionID: bypassResp["id"].(string),
        LocalPort:    req.LocalPort,
        Status:       "connected",
        Config: V2RayConfig{
            Inbounds:  buildV2RayInbounds(req),
            Outbounds: buildV2RayOutbounds(req),
            Routing:   buildV2RayRouting(),
        },
        CreatedAt: time.Now(),
    }

    writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode V2Ray connect response")
}

// Obfs4-specific connection handler
func (h *Handlers) ConnectObfs4Handler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req Obfs4ConnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Create Obfs4-specific bypass
    bypassReq := map[string]interface{}{
        "name":        req.Name,
        "method":      "obfs4",
        "local_port":  req.LocalPort,
        "bridge":      req.Bridge,
        "cert":        req.Cert,
        "iat_mode":    req.IATMode,
    }
    
    bypassResp, err := h.proxyService.CreateBypass(r.Context(), bypassReq)
    if err != nil {
        h.logger.Error("failed to create Obfs4 bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to create Obfs4 bypass")
        return
    }

    // Start bypass
    if err := h.proxyService.StartBypass(r.Context(), bypassResp["id"].(string)); err != nil {
        h.logger.Error("failed to start Obfs4 bypass", zap.Error(err))
        writeError(w, http.StatusInternalServerError, "Failed to start Obfs4 bypass")
        return
    }

    response := Obfs4ConnectResponse{
        ConnectionID: bypassResp["id"].(string),
        LocalPort:    req.LocalPort,
        Status:       "connected",
        Config: Obfs4Config{
            Bridge:    req.Bridge,
            Transport: "obfs4",
            LocalPort: req.LocalPort,
        },
        CreatedAt: time.Now(),
    }

    writeJSON(w, http.StatusCreated, response, h.logger, "failed to encode Obfs4 connect response")
}

// Disconnect handler
func (h *Handlers) DisconnectHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    var req DisconnectRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    var disconnected []string
    var errors []string

    // Handle different disconnect scenarios
    if req.All {
        // Disconnect all active connections
        if tunnels, err := h.proxyService.GetActiveTunnels(r.Context()); err == nil {
            for _, tunnel := range tunnels {
                if err := h.proxyService.StopVPNTunnel(r.Context(), tunnel.ID); err == nil {
                    disconnected = append(disconnected, tunnel.ID)
                } else {
                    errors = append(errors, fmt.Sprintf("Failed to stop tunnel %s: %v", tunnel.ID, err))
                }
            }
        }

        if bypasses, err := h.proxyService.GetActiveBypasses(r.Context()); err == nil {
            for _, bypass := range bypasses {
                if err := h.proxyService.StopBypass(r.Context(), bypass.ID); err == nil {
                    disconnected = append(disconnected, bypass.ID)
                } else {
                    errors = append(errors, fmt.Sprintf("Failed to stop bypass %s: %v", bypass.ID, err))
                }
            }
        }
    } else {
        // Disconnect specific connections
        if req.ConnectionID != "" {
            if err := h.proxyService.StopConnection(r.Context(), req.ConnectionID); err == nil {
                disconnected = append(disconnected, req.ConnectionID)
            } else {
                errors = append(errors, fmt.Sprintf("Failed to stop connection %s: %v", req.ConnectionID, err))
            }
        }

        if req.TunnelID != "" {
            if err := h.proxyService.StopVPNTunnel(r.Context(), req.TunnelID); err == nil {
                disconnected = append(disconnected, req.TunnelID)
            } else {
                errors = append(errors, fmt.Sprintf("Failed to stop tunnel %s: %v", req.TunnelID, err))
            }
        }

        if req.BypassID != "" {
            if err := h.proxyService.StopBypass(r.Context(), req.BypassID); err == nil {
                disconnected = append(disconnected, req.BypassID)
            } else {
                errors = append(errors, fmt.Sprintf("Failed to stop bypass %s: %v", req.BypassID, err))
            }
        }
    }

    // Determine response status
    status := "success"
    if len(errors) > 0 {
        if len(disconnected) > 0 {
            status = "partial"
        } else {
            status = "error"
        }
    }

    response := DisconnectResponse{
        Disconnected: disconnected,
        Status:       status,
        Errors:       errors,
    }

    writeJSON(w, http.StatusOK, response, h.logger, "failed to encode disconnect response")
}

// Connection status handler
func (h *Handlers) ConnectionStatusHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    // Get active VPN tunnels
    vpnTunnels := []VPNTunnelStatus{}
    if tunnels, err := h.proxyService.GetActiveTunnels(r.Context()); err == nil {
        for _, tunnel := range tunnels {
            stats, _ := h.proxyService.GetTunnelStats(r.Context(), tunnel.ID)
            vpnTunnels = append(vpnTunnels, VPNTunnelStatus{
                TunnelID:    tunnel.ID,
                Name:        tunnel.Name,
                Status:      tunnel.Status,
                ConnectedAt: tunnel.CreatedAt,
                PeersCount:  stats.PeersCount,
                BytesRx:     stats.BytesRx,
                BytesTx:     stats.BytesTx,
            })
        }
    }

    // Get active DPI bypasses
    dpiBypasses := []DPIBypassStatus{}
    if bypasses, err := h.proxyService.GetActiveBypasses(r.Context()); err == nil {
        for _, bypass := range bypasses {
            stats, _ := h.proxyService.GetBypassStats(r.Context(), bypass.ID)
            dpiBypasses = append(dpiBypasses, DPIBypassStatus{
                BypassID:    bypass.ID,
                Name:        bypass.Name,
                Method:      bypass.Method,
                Status:      bypass.Status,
                LocalPort:   bypass.LocalPort,
                ConnectedAt: bypass.CreatedAt,
                BytesRx:     stats.BytesRx,
                BytesTx:     stats.BytesTx,
            })
        }
    }

    // Calculate totals
    totalDataTransferred := int64(0)
    for _, tunnel := range vpnTunnels {
        totalDataTransferred += tunnel.BytesRx + tunnel.BytesTx
    }
    for _, bypass := range dpiBypasses {
        totalDataTransferred += bypass.BytesRx + bypass.BytesTx
    }

    response := ConnectionStatusResponse{
        VPNTunnels:           vpnTunnels,
        DPIBypasses:          dpiBypasses,
        ActiveConnections:    len(vpnTunnels) + len(dpiBypasses),
        TotalDataTransferred: totalDataTransferred,
        Uptime:               calculateUptime(vpnTunnels, dpiBypasses),
    }

    writeJSON(w, http.StatusOK, response, h.logger, "failed to encode connection status response")
}
```

#### 1.2 Add New Request/Response Types
Add to `silence/api/gateway/internal/adapters/http/handlers.go`:

```go
// VPN-only connection types
type VPNConnectRequest struct {
    Name         string `json:"name"`
    ListenPort   int    `json:"listen_port"`
    MTU          int    `json:"mtu"`
    AutoRecovery bool   `json:"auto_recovery"`
    ServerID     string `json:"server_id,omitempty"`
    Region       string `json:"region,omitempty"`
}

type VPNConnectResponse struct {
    TunnelID    string        `json:"tunnel_id"`
    TunnelName  string        `json:"tunnel_name"`
    PublicKey   string        `json:"public_key"`
    Endpoint    string        `json:"endpoint"`
    Status      string        `json:"status"`
    Config      VPNConfig     `json:"config"`
    CreatedAt   time.Time     `json:"created_at"`
}

type VPNConfig struct {
    Interface   string    `json:"interface"`
    PrivateKey  string    `json:"private_key"`
    Address     string    `json:"address"`
    DNS         []string  `json:"dns"`
    Peer        VPNPeer   `json:"peer"`
}

type VPNPeer struct {
    PublicKey   string   `json:"public_key"`
    AllowedIPs  []string `json:"allowed_ips"`
    Endpoint    string   `json:"endpoint"`
}

// DPI-only connection types
type DPIConnectRequest struct {
    Method     string `json:"method"`
    Name       string `json:"name"`
    RemoteHost string `json:"remote_host"`
    RemotePort int    `json:"remote_port"`
    LocalPort  int    `json:"local_port,omitempty"`
    Password   string `json:"password,omitempty"`
    Encryption string `json:"encryption,omitempty"`
    ServerID   string `json:"server_id,omitempty"`
}

type DPIConnectResponse struct {
    BypassID  string                 `json:"bypass_id"`
    Method    string                 `json:"method"`
    LocalPort int                    `json:"local_port"`
    Status    string                 `json:"status"`
    Config    map[string]interface{} `json:"config"`
    CreatedAt time.Time              `json:"created_at"`
}

// Shadowsocks-specific types
type ShadowsocksConnectRequest struct {
    Name       string `json:"name"`
    ServerHost string `json:"server_host"`
    ServerPort int    `json:"server_port"`
    Password   string `json:"password"`
    Encryption string `json:"encryption"`
    LocalPort  int    `json:"local_port,omitempty"`
    Timeout    int    `json:"timeout,omitempty"`
}

type ShadowsocksConnectResponse struct {
    ConnectionID string            `json:"connection_id"`
    LocalPort    int               `json:"local_port"`
    Status       string            `json:"status"`
    Config       ShadowsocksConfig `json:"config"`
    CreatedAt    time.Time         `json:"created_at"`
}

type ShadowsocksConfig struct {
    Server     string `json:"server"`
    ServerPort int    `json:"server_port"`
    LocalPort  int    `json:"local_port"`
    Password   string `json:"password"`
    Method     string `json:"method"`
}

// V2Ray-specific types
type V2RayConnectRequest struct {
    Name       string `json:"name"`
    ServerHost string `json:"server_host"`
    ServerPort int    `json:"server_port"`
    UUID       string `json:"uuid"`
    AlterID    int    `json:"alter_id"`
    Security   string `json:"security,omitempty"`
    Network    string `json:"network,omitempty"`
    LocalPort  int    `json:"local_port,omitempty"`
}

type V2RayConnectResponse struct {
    ConnectionID string      `json:"connection_id"`
    LocalPort    int         `json:"local_port"`
    Status       string      `json:"status"`
    Config       V2RayConfig `json:"config"`
    CreatedAt    time.Time   `json:"created_at"`
}

type V2RayConfig struct {
    Inbounds  []interface{} `json:"inbounds"`
    Outbounds []interface{} `json:"outbounds"`
    Routing   interface{}   `json:"routing"`
}

// Obfs4-specific types
type Obfs4ConnectRequest struct {
    Name      string `json:"name"`
    Bridge    string `json:"bridge"`
    Cert      string `json:"cert"`
    IATMode   string `json:"iat_mode,omitempty"`
    LocalPort int    `json:"local_port,omitempty"`
}

type Obfs4ConnectResponse struct {
    ConnectionID string      `json:"connection_id"`
    LocalPort    int         `json:"local_port"`
    Status       string      `json:"status"`
    Config       Obfs4Config `json:"config"`
    CreatedAt    time.Time   `json:"created_at"`
}

type Obfs4Config struct {
    Bridge    string `json:"bridge"`
    Transport string `json:"transport"`
    LocalPort int    `json:"local_port"`
}

// Disconnect types
type DisconnectRequest struct {
    ConnectionID string `json:"connection_id,omitempty"`
    TunnelID     string `json:"tunnel_id,omitempty"`
    BypassID     string `json:"bypass_id,omitempty"`
    All          bool   `json:"all,omitempty"`
}

type DisconnectResponse struct {
    Disconnected []string `json:"disconnected"`
    Status       string   `json:"status"`
    Errors       []string `json:"errors,omitempty"`
}

// Connection status types
type ConnectionStatusResponse struct {
    VPNTunnels           []VPNTunnelStatus `json:"vpn_tunnels"`
    DPIBypasses          []DPIBypassStatus `json:"dpi_bypasses"`
    ActiveConnections    int               `json:"active_connections"`
    TotalDataTransferred int64             `json:"total_data_transferred"`
    Uptime               int64             `json:"uptime"`
}

type VPNTunnelStatus struct {
    TunnelID    string    `json:"tunnel_id"`
    Name        string    `json:"name"`
    Status      string    `json:"status"`
    ConnectedAt time.Time `json:"connected_at"`
    PeersCount  int       `json:"peers_count"`
    BytesRx     int64     `json:"bytes_rx"`
    BytesTx     int64     `json:"bytes_tx"`
}

type DPIBypassStatus struct {
    BypassID    string    `json:"bypass_id"`
    Name        string    `json:"name"`
    Method      string    `json:"method"`
    Status      string    `json:"status"`
    LocalPort   int       `json:"local_port"`
    ConnectedAt time.Time `json:"connected_at"`
    BytesRx     int64     `json:"bytes_rx"`
    BytesTx     int64     `json:"bytes_tx"`
}
```

#### 1.3 Update Route Registration
Update `silence/api/gateway/internal/adapters/http/server.go`:

```go
// Add these routes to the NewServer function
mux.Handle("/api/v1/connect/vpn", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectVPNHandler))))
mux.Handle("/api/v1/connect/dpi", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectDPIHandler))))
mux.Handle("/api/v1/connect/shadowsocks", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectShadowsocksHandler))))
mux.Handle("/api/v1/connect/v2ray", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectV2RayHandler))))
mux.Handle("/api/v1/connect/obfs4", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectObfs4Handler))))
mux.Handle("/api/v1/disconnect", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.DisconnectHandler))))
mux.Handle("/api/v1/connect/status", wrap(NewAuthMiddleware(jwtSecret)(http.HandlerFunc(handlers.ConnectionStatusHandler))))
```

### Phase 2: WebSocket Implementation

#### 2.1 Add WebSocket Handler
Create `silence/api/gateway/internal/adapters/http/websocket.go`:

```go
package http

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
    "go.uber.org/zap"
)

type WebSocketHandler struct {
    upgrader websocket.Upgrader
    logger   *zap.Logger
    clients  map[*websocket.Conn]bool
}

func NewWebSocketHandler(logger *zap.Logger) *WebSocketHandler {
    return &WebSocketHandler{
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Configure based on your CORS policy
            },
        },
        logger:  logger,
        clients: make(map[*websocket.Conn]bool),
    }
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        h.logger.Error("failed to upgrade websocket connection", zap.Error(err))
        return
    }
    defer conn.Close()

    h.clients[conn] = true
    defer delete(h.clients, conn)

    for {
        var message map[string]interface{}
        err := conn.ReadJSON(&message)
        if err != nil {
            h.logger.Error("failed to read websocket message", zap.Error(err))
            break
        }

        // Handle different message types
        switch message["type"] {
        case "auth":
            // Handle authentication
            h.handleAuth(conn, message)
        case "subscribe":
            // Handle subscription to events
            h.handleSubscribe(conn, message)
        case "ping":
            // Handle ping
            h.handlePing(conn)
        }
    }
}

func (h *WebSocketHandler) BroadcastConnectionStatus(status interface{}) {
    message := map[string]interface{}{
        "type":      "connection_status",
        "data":      status,
        "timestamp": time.Now(),
    }

    for conn := range h.clients {
        if err := conn.WriteJSON(message); err != nil {
            h.logger.Error("failed to send websocket message", zap.Error(err))
            conn.Close()
            delete(h.clients, conn)
        }
    }
}

func (h *WebSocketHandler) BroadcastMetrics(metrics interface{}) {
    message := map[string]interface{}{
        "type":      "metrics_update",
        "data":      metrics,
        "timestamp": time.Now(),
    }

    for conn := range h.clients {
        if err := conn.WriteJSON(message); err != nil {
            h.logger.Error("failed to send websocket message", zap.Error(err))
            conn.Close()
            delete(h.clients, conn)
        }
    }
}

func (h *WebSocketHandler) handleAuth(conn *websocket.Conn, message map[string]interface{}) {
    // Validate JWT token
    token, ok := message["token"].(string)
    if !ok {
        conn.WriteJSON(map[string]interface{}{
            "type":  "auth_error",
            "error": "Token required",
        })
        return
    }

    // Validate token (implement your JWT validation logic)
    if !h.validateToken(token) {
        conn.WriteJSON(map[string]interface{}{
            "type":  "auth_error",
            "error": "Invalid token",
        })
        return
    }

    conn.WriteJSON(map[string]interface{}{
        "type":    "auth_success",
        "message": "Authenticated successfully",
    })
}

func (h *WebSocketHandler) handleSubscribe(conn *websocket.Conn, message map[string]interface{}) {
    // Handle subscription to specific events
    events, ok := message["events"].([]interface{})
    if !ok {
        conn.WriteJSON(map[string]interface{}{
            "type":  "subscribe_error",
            "error": "Events array required",
        })
        return
    }

    // Store subscription preferences (implement as needed)
    conn.WriteJSON(map[string]interface{}{
        "type":    "subscribe_success",
        "events":  events,
        "message": "Subscribed to events",
    })
}