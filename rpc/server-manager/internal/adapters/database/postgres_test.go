package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	server := &domain.Server{
		Name:    "test-server",
		Type:    domain.ServerTypeVPN,
		Status:  domain.ServerStatusRunning,
		Region:  "us-east-1",
		IP:      "127.0.0.1",
		Port:    1194,
		CPU:     0.5,
		Memory:  1024,
		Disk:    20480,
		Network: 100,
	}

	mock.ExpectExec("INSERT INTO servers").
		WithArgs(sqlmock.AnyArg(), server.Name, server.Type, server.Status, server.Region,
			server.IP, server.Port, server.CPU, server.Memory, server.Disk, server.Network,
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), server)
	assert.NoError(t, err)
	assert.NotEmpty(t, server.ID)
	assert.False(t, server.CreatedAt.IsZero())
	assert.False(t, server.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	serverID := uuid.New().String()
	expectedServer := &domain.Server{
		ID:        serverID,
		Name:      "test-server",
		Type:      domain.ServerTypeVPN,
		Status:    domain.ServerStatusRunning,
		Region:    "us-east-1",
		IP:        "127.0.0.1",
		Port:      1194,
		CPU:       0.5,
		Memory:    1024,
		Disk:      20480,
		Network:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "status", "region", "ip", "port", "cpu", "memory", "disk", "network", "created_at", "updated_at", "deleted_at"}).
		AddRow(expectedServer.ID, expectedServer.Name, expectedServer.Type, expectedServer.Status, expectedServer.Region,
			expectedServer.IP, expectedServer.Port, expectedServer.CPU, expectedServer.Memory, expectedServer.Disk, expectedServer.Network,
			expectedServer.CreatedAt, expectedServer.UpdatedAt, nil)

	mock.ExpectQuery(`SELECT .+ FROM servers WHERE id = \$1 AND deleted_at IS NULL`).
		WithArgs(serverID).
		WillReturnRows(rows)

	server, err := repo.GetByID(context.Background(), serverID)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, expectedServer.ID, server.ID)
	assert.Equal(t, expectedServer.Name, server.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	server1 := &domain.Server{
		ID:        uuid.New().String(),
		Name:      "test-server-1",
		Type:      domain.ServerTypeVPN,
		Status:    domain.ServerStatusRunning,
		Region:    "us-east-1",
		IP:        "127.0.0.1",
		Port:      1194,
		CPU:       0.5,
		Memory:    1024,
		Disk:      20480,
		Network:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	server2 := &domain.Server{
		ID:        uuid.New().String(),
		Name:      "test-server-2",
		Type:      domain.ServerTypeDPI,
		Status:    domain.ServerStatusStopped,
		Region:    "eu-west-1",
		IP:        "127.0.0.2",
		Port:      8080,
		CPU:       0.8,
		Memory:    2048,
		Disk:      40960,
		Network:   200,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "status", "region", "ip", "port", "cpu", "memory", "disk", "network", "created_at", "updated_at", "deleted_at"}).
		AddRow(server1.ID, server1.Name, server1.Type, server1.Status, server1.Region,
			server1.IP, server1.Port, server1.CPU, server1.Memory, server1.Disk, server1.Network,
			server1.CreatedAt, server1.UpdatedAt, nil).
		AddRow(server2.ID, server2.Name, server2.Type, server2.Status, server2.Region,
			server2.IP, server2.Port, server2.CPU, server2.Memory, server2.Disk, server2.Network,
			server2.CreatedAt, server2.UpdatedAt, nil)

	mock.ExpectQuery("SELECT .+ FROM servers WHERE deleted_at IS NULL ORDER BY created_at DESC").
		WillReturnRows(rows)

	servers, err := repo.List(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, servers, 2)
	assert.Equal(t, server1.ID, servers[0].ID)
	assert.Equal(t, server2.ID, servers[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	serverID := uuid.New().String()
	updatedServer := &domain.Server{
		ID:        serverID,
		Name:      "updated-server",
		Type:      domain.ServerTypeDPI,
		Status:    domain.ServerStatusStopped,
		Region:    "eu-west-1",
		IP:        "127.0.0.3",
		Port:      8080,
		CPU:       0.8,
		Memory:    2048,
		Disk:      40960,
		Network:   200,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	mock.ExpectExec("UPDATE servers").
		WithArgs(updatedServer.ID, updatedServer.Name, updatedServer.Type, updatedServer.Status, updatedServer.Region,
			updatedServer.IP, updatedServer.Port, updatedServer.CPU, updatedServer.Memory, updatedServer.Disk, updatedServer.Network,
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(context.Background(), updatedServer)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	serverID := uuid.New().String()

	mock.ExpectExec(`UPDATE servers SET deleted_at = .+ WHERE id = .+ AND deleted_at IS NULL`).
		WithArgs(serverID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(context.Background(), serverID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetByType(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	serverType := domain.ServerTypeVPN
	expectedServer := &domain.Server{
		ID:        uuid.New().String(),
		Name:      "vpn-server",
		Type:      serverType,
		Status:    domain.ServerStatusRunning,
		Region:    "us-east-1",
		IP:        "127.0.0.1",
		Port:      1194,
		CPU:       0.5,
		Memory:    1024,
		Disk:      20480,
		Network:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "status", "region", "ip", "port", "cpu", "memory", "disk", "network", "created_at", "updated_at", "deleted_at"}).
		AddRow(expectedServer.ID, expectedServer.Name, expectedServer.Type, expectedServer.Status, expectedServer.Region,
			expectedServer.IP, expectedServer.Port, expectedServer.CPU, expectedServer.Memory, expectedServer.Disk, expectedServer.Network,
			expectedServer.CreatedAt, expectedServer.UpdatedAt, nil)

	mock.ExpectQuery(`SELECT .+ FROM servers WHERE deleted_at IS NULL AND type = \$1 ORDER BY created_at DESC`).
		WithArgs(serverType).
		WillReturnRows(rows)

	servers, err := repo.GetByType(context.Background(), serverType)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, expectedServer.ID, servers[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetByRegion(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	region := "us-east-1"
	expectedServer := &domain.Server{
		ID:        uuid.New().String(),
		Name:      "vpn-server",
		Type:      domain.ServerTypeVPN,
		Status:    domain.ServerStatusRunning,
		Region:    region,
		IP:        "127.0.0.1",
		Port:      1194,
		CPU:       0.5,
		Memory:    1024,
		Disk:      20480,
		Network:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "status", "region", "ip", "port", "cpu", "memory", "disk", "network", "created_at", "updated_at", "deleted_at"}).
		AddRow(expectedServer.ID, expectedServer.Name, expectedServer.Type, expectedServer.Status, expectedServer.Region,
			expectedServer.IP, expectedServer.Port, expectedServer.CPU, expectedServer.Memory, expectedServer.Disk, expectedServer.Network,
			expectedServer.CreatedAt, expectedServer.UpdatedAt, nil)

	mock.ExpectQuery(`SELECT .+ FROM servers WHERE deleted_at IS NULL AND region = \$1 ORDER BY created_at DESC`).
		WithArgs(region).
		WillReturnRows(rows)

	servers, err := repo.GetByRegion(context.Background(), region)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, expectedServer.ID, servers[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetByStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db, zap.NewNop())

	status := domain.ServerStatusRunning
	expectedServer := &domain.Server{
		ID:        uuid.New().String(),
		Name:      "vpn-server",
		Type:      domain.ServerTypeVPN,
		Status:    status,
		Region:    "us-east-1",
		IP:        "127.0.0.1",
		Port:      1194,
		CPU:       0.5,
		Memory:    1024,
		Disk:      20480,
		Network:   100,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "status", "region", "ip", "port", "cpu", "memory", "disk", "network", "created_at", "updated_at", "deleted_at"}).
		AddRow(expectedServer.ID, expectedServer.Name, expectedServer.Type, expectedServer.Status, expectedServer.Region,
			expectedServer.IP, expectedServer.Port, expectedServer.CPU, expectedServer.Memory, expectedServer.Disk, expectedServer.Network,
			expectedServer.CreatedAt, expectedServer.UpdatedAt, nil)

	mock.ExpectQuery(`SELECT .+ FROM servers WHERE deleted_at IS NULL AND status = \$1 ORDER BY created_at DESC`).
		WithArgs(status).
		WillReturnRows(rows)

	servers, err := repo.GetByStatus(context.Background(), status)
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, expectedServer.ID, servers[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
