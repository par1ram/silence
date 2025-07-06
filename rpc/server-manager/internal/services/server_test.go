package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"github.com/par1ram/silence/rpc/server-manager/internal/services"
	. "github.com/par1ram/silence/rpc/server-manager/internal/services/mocks"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mock_server.go -package=services_test github.com/par1ram/silence/rpc/server-manager/internal/ports ServerRepository,StatsRepository,HealthRepository,ScalingRepository,BackupRepository,UpdateRepository

func TestServerService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServerService Suite")
}

var _ = Describe("ServerService", func() {
	var serverService *services.ServerService
	var ctx context.Context
	var logger *zap.Logger
	var mockServerRepo *MockServerRepository
	var mockStatsRepo *MockStatsRepository
	var mockHealthRepo *MockHealthRepository
	var mockScalingRepo *MockScalingRepository
	var mockBackupRepo *MockBackupRepository
	var mockUpdateRepo *MockUpdateRepository
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockServerRepo = NewMockServerRepository(ctrl)
		mockStatsRepo = NewMockStatsRepository(ctrl)
		mockHealthRepo = NewMockHealthRepository(ctrl)
		mockScalingRepo = NewMockScalingRepository(ctrl)
		mockBackupRepo = NewMockBackupRepository(ctrl)
		mockUpdateRepo = NewMockUpdateRepository(ctrl)
		logger = zap.NewNop()
		serverService = services.NewServerService(
			mockServerRepo,
			mockStatsRepo,
			mockHealthRepo,
			mockScalingRepo,
			mockBackupRepo,
			mockUpdateRepo,
			nil, // docker adapter - will be mocked in tests that need it
			logger,
		).(*services.ServerService)
		ctx = context.Background()
	})

	Describe("GetServer", func() {
		It("should return server", func() {
			serverID := "test-server-id"
			expectedServer := &domain.Server{
				ID:     serverID,
				Name:   "test-server",
				Type:   domain.ServerTypeVPN,
				Status: domain.ServerStatusRunning,
				Region: "us-east-1",
			}

			mockServerRepo.EXPECT().GetByID(ctx, serverID).Return(expectedServer, nil)

			server, err := serverService.GetServer(ctx, serverID)

			Expect(err).To(BeNil())
			Expect(server).NotTo(BeNil())
			Expect(server.ID).To(Equal(expectedServer.ID))
			Expect(server.Name).To(Equal(expectedServer.Name))
		})

		It("should return error for nonexistent server", func() {
			serverID := "nonexistent-server-id"

			mockServerRepo.EXPECT().GetByID(ctx, serverID).Return(nil, fmt.Errorf("server not found"))

			server, err := serverService.GetServer(ctx, serverID)

			Expect(err).NotTo(BeNil())
			Expect(server).To(BeNil())
		})
	})

	Describe("ListServers", func() {
		It("should return list of servers", func() {
			expectedServers := []*domain.Server{
				{
					ID:     "server-1",
					Name:   "server-1",
					Type:   domain.ServerTypeVPN,
					Status: domain.ServerStatusRunning,
					Region: "us-east-1",
				},
				{
					ID:     "server-2",
					Name:   "server-2",
					Type:   domain.ServerTypeVPN,
					Status: domain.ServerStatusRunning,
					Region: "us-west-1",
				},
			}

			mockServerRepo.EXPECT().List(ctx, gomock.Any()).Return(expectedServers, nil)

			servers, err := serverService.ListServers(ctx, map[string]interface{}{})

			Expect(err).To(BeNil())
			Expect(servers).NotTo(BeNil())
			Expect(len(servers)).To(Equal(2))
		})

		It("should return empty list if no servers", func() {
			mockServerRepo.EXPECT().List(ctx, gomock.Any()).Return([]*domain.Server{}, nil)

			servers, err := serverService.ListServers(ctx, map[string]interface{}{})

			Expect(err).To(BeNil())
			Expect(servers).NotTo(BeNil())
			Expect(len(servers)).To(Equal(0))
		})
	})
})
