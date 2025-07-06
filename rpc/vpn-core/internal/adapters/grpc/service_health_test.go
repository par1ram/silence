package grpc_test

import (
	context "context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/api/proto"
	grpcsvc "github.com/par1ram/silence/rpc/vpn-core/internal/adapters/grpc"
	"go.uber.org/zap"
)

func TestServiceHealth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VpnCoreService Health Suite")
}

var _ = Describe("VpnCoreService.Health", func() {
	var (
		service *grpcsvc.VpnCoreService
	)

	BeforeEach(func() {
		service = grpcsvc.NewVpnCoreService(nil, nil, zap.NewNop())
	})

	It("should return ok status", func() {
		ctx := context.Background()
		resp, err := service.Health(ctx, &proto.HealthRequest{})
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())
		Expect(resp.Status).To(Equal("ok"))
		Expect(resp.Version).NotTo(BeEmpty())
	})
})
