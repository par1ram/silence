package services_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
)

func TestHealthService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HealthService Suite")
}

var _ = Describe("HealthService", func() {
	var healthService *services.HealthService

	BeforeEach(func() {
		healthService = services.NewHealthService("test-service", "1.0.0").(*services.HealthService)
	})

	Describe("GetHealth", func() {
		It("should return health status", func() {
			status := healthService.GetHealth()

			Expect(status).NotTo(BeNil())
			Expect(status.Service).To(Equal("test-service"))
			Expect(status.Version).To(Equal("1.0.0"))
			Expect(status.Status).To(Equal("ok"))
			Expect(status.Timestamp).NotTo(BeZero())
		})
	})
})
