package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/par1ram/silence/api/gateway/internal/adapters/redis"
	sharedRedis "github.com/par1ram/silence/shared/redis"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestRedisIntegration(t *testing.T) {
	// Skip if Redis is not available
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create Redis client
	redisConfig := &sharedRedis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       1, // Use test database
		Prefix:   "test_gateway",
	}

	redisClient, err := sharedRedis.NewClient(redisConfig, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisClient.Close()

	// Clean up test data
	defer func() {
		redisClient.FlushDB(ctx)
	}()

	t.Run("RateLimiter", func(t *testing.T) {
		testRateLimiter(t, redisClient, logger)
	})

	t.Run("WebSocketSessions", func(t *testing.T) {
		testWebSocketSessions(t, redisClient, logger)
	})
}

func testRateLimiter(t *testing.T, redisClient *sharedRedis.Client, logger *zap.Logger) {
	// Create rate limiter adapter
	config := &redis.RateLimiterConfig{
		DefaultRPS:      10,
		DefaultBurst:    20,
		Window:          time.Minute,
		KeyPrefix:       "test_rate_limit",
		CleanupInterval: time.Second,
	}

	rateLimiter := redis.NewRateLimiterAdapter(redisClient, config, logger)
	defer rateLimiter.Close()

	// Test basic rate limiting
	clientIP := "192.168.1.100"
	endpoint := "test"

	// First request should be allowed
	allowed := rateLimiter.Allow(clientIP, endpoint)
	if !allowed {
		t.Error("First request should be allowed")
	}

	// Check detailed result
	result, err := rateLimiter.CheckLimit(clientIP, endpoint)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}

	if !result.Allowed {
		t.Error("Request should be allowed")
	}

	if result.Remaining >= int64(config.DefaultBurst) {
		t.Errorf("Remaining should be less than burst limit, got %d", result.Remaining)
	}

	// Test whitelist functionality
	err = rateLimiter.AddToWhitelist(clientIP)
	if err != nil {
		t.Fatalf("Failed to add to whitelist: %v", err)
	}

	// Whitelisted IP should always be allowed
	for i := 0; i < config.DefaultBurst+10; i++ {
		if !rateLimiter.Allow(clientIP, endpoint) {
			t.Error("Whitelisted IP should always be allowed")
		}
	}

	// Test stats
	stats, err := rateLimiter.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats["total_requests"].(int64) == 0 {
		t.Error("Stats should show some requests")
	}

	// Test reset stats
	err = rateLimiter.ResetStats()
	if err != nil {
		t.Fatalf("Failed to reset stats: %v", err)
	}
}

func testWebSocketSessions(t *testing.T, redisClient *sharedRedis.Client, logger *zap.Logger) {
	// Create WebSocket session manager
	config := &redis.WebSocketSessionConfig{
		KeyPrefix:       "test_websocket",
		SessionTTL:      time.Hour,
		CleanupInterval: time.Second,
		MaxSessions:     1000,
	}

	sessionManager := redis.NewWebSocketSessionManager(redisClient, config, logger)
	defer sessionManager.Close()

	ctx := context.Background()

	// Create test session
	session := &redis.WebSocketSession{
		ID:            "test_session_001",
		UserID:        "test_user_001",
		ClientIP:      "192.168.1.100",
		UserAgent:     "Test Agent",
		ConnectedAt:   time.Now(),
		LastActivity:  time.Now(),
		Authenticated: false,
		Subscriptions: []string{},
		Metadata:      map[string]interface{}{"test": "data"},
	}

	// Test session creation
	err := sessionManager.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test session retrieval
	retrievedSession, err := sessionManager.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrievedSession.ID != session.ID {
		t.Errorf("Retrieved session ID mismatch: got %s, want %s", retrievedSession.ID, session.ID)
	}

	if retrievedSession.UserID != session.UserID {
		t.Errorf("Retrieved session UserID mismatch: got %s, want %s", retrievedSession.UserID, session.UserID)
	}

	// Test session authentication
	err = sessionManager.AuthenticateSession(ctx, session.ID, "authenticated_user")
	if err != nil {
		t.Fatalf("Failed to authenticate session: %v", err)
	}

	// Verify authentication
	authenticatedSession, err := sessionManager.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get authenticated session: %v", err)
	}

	if !authenticatedSession.Authenticated {
		t.Error("Session should be authenticated")
	}

	if authenticatedSession.UserID != "authenticated_user" {
		t.Errorf("User ID should be updated after authentication: got %s, want %s", authenticatedSession.UserID, "authenticated_user")
	}

	// Test subscriptions
	subscription := "test_events"
	err = sessionManager.AddSubscription(ctx, session.ID, subscription)
	if err != nil {
		t.Fatalf("Failed to add subscription: %v", err)
	}

	// Verify subscription
	updatedSession, err := sessionManager.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	found := false
	for _, sub := range updatedSession.Subscriptions {
		if sub == subscription {
			found = true
			break
		}
	}

	if !found {
		t.Error("Subscription should be added to session")
	}

	// Test sessions by user
	userSessions, err := sessionManager.GetSessionsByUser(ctx, "authenticated_user")
	if err != nil {
		t.Fatalf("Failed to get sessions by user: %v", err)
	}

	if len(userSessions) != 1 {
		t.Errorf("Expected 1 session for user, got %d", len(userSessions))
	}

	// Test sessions by subscription
	subscriptionSessions, err := sessionManager.GetSessionsBySubscription(ctx, subscription)
	if err != nil {
		t.Fatalf("Failed to get sessions by subscription: %v", err)
	}

	if len(subscriptionSessions) != 1 {
		t.Errorf("Expected 1 session for subscription, got %d", len(subscriptionSessions))
	}

	// Test remove subscription
	err = sessionManager.RemoveSubscription(ctx, session.ID, subscription)
	if err != nil {
		t.Fatalf("Failed to remove subscription: %v", err)
	}

	// Verify subscription removed
	finalSession, err := sessionManager.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to get final session: %v", err)
	}

	for _, sub := range finalSession.Subscriptions {
		if sub == subscription {
			t.Error("Subscription should be removed from session")
		}
	}

	// Test stats
	stats, err := sessionManager.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get session stats: %v", err)
	}

	if stats.TotalSessions == 0 {
		t.Error("Stats should show created sessions")
	}

	if stats.AuthenticatedSessions == 0 {
		t.Error("Stats should show authenticated sessions")
	}

	// Test session deletion
	err = sessionManager.DeleteSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session is deleted
	_, err = sessionManager.GetSession(ctx, session.ID)
	if err == nil {
		t.Error("Session should be deleted")
	}
}

func TestRedisClientOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis client test in short mode")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// Create Redis client
	redisConfig := &sharedRedis.Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       2, // Use different test database
		Prefix:   "test_client",
	}

	redisClient, err := sharedRedis.NewClient(redisConfig, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer redisClient.Close()

	// Clean up test data
	defer func() {
		redisClient.FlushDB(ctx)
	}()

	// Test basic operations
	t.Run("BasicOperations", func(t *testing.T) {
		// Test Set/Get
		testData := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		err := redisClient.Set(ctx, "test_key", testData, time.Hour)
		if err != nil {
			t.Fatalf("Failed to set data: %v", err)
		}

		var retrievedData map[string]interface{}
		err = redisClient.Get(ctx, "test_key", &retrievedData)
		if err != nil {
			t.Fatalf("Failed to get data: %v", err)
		}

		if retrievedData["key1"] != testData["key1"] {
			t.Errorf("Data mismatch: got %v, want %v", retrievedData["key1"], testData["key1"])
		}

		// Test Exists
		exists, err := redisClient.Exists(ctx, "test_key")
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}

		if !exists {
			t.Error("Key should exist")
		}

		// Test Delete
		err = redisClient.Delete(ctx, "test_key")
		if err != nil {
			t.Fatalf("Failed to delete key: %v", err)
		}

		exists, err = redisClient.Exists(ctx, "test_key")
		if err != nil {
			t.Fatalf("Failed to check existence after delete: %v", err)
		}

		if exists {
			t.Error("Key should not exist after delete")
		}
	})

	// Test Hash operations
	t.Run("HashOperations", func(t *testing.T) {
		hashKey := "test_hash"

		// Test HSet/HGet
		testValue := "test_value"
		err := redisClient.HSet(ctx, hashKey, "field1", testValue)
		if err != nil {
			t.Fatalf("Failed to set hash field: %v", err)
		}

		var retrievedValue string
		err = redisClient.HGet(ctx, hashKey, "field1", &retrievedValue)
		if err != nil {
			t.Fatalf("Failed to get hash field: %v", err)
		}

		if retrievedValue != testValue {
			t.Errorf("Hash field value mismatch: got %v, want %v", retrievedValue, testValue)
		}

		// Test HIncrBy
		count, err := redisClient.HIncrBy(ctx, hashKey, "counter", 5)
		if err != nil {
			t.Fatalf("Failed to increment hash field: %v", err)
		}

		if count != 5 {
			t.Errorf("Expected counter to be 5, got %d", count)
		}

		// Test HExists
		exists, err := redisClient.HExists(ctx, hashKey, "field1")
		if err != nil {
			t.Fatalf("Failed to check hash field existence: %v", err)
		}

		if !exists {
			t.Error("Hash field should exist")
		}

		// Test HGetAll
		allFields, err := redisClient.HGetAll(ctx, hashKey)
		if err != nil {
			t.Fatalf("Failed to get all hash fields: %v", err)
		}

		if len(allFields) < 2 {
			t.Errorf("Expected at least 2 fields, got %d", len(allFields))
		}
	})

	// Test Set operations
	t.Run("SetOperations", func(t *testing.T) {
		setKey := "test_set"

		// Test SAdd
		err := redisClient.SAdd(ctx, setKey, "member1")
		if err != nil {
			t.Fatalf("Failed to add to set: %v", err)
		}

		// Test SIsMember
		isMember, err := redisClient.SIsMember(ctx, setKey, "member1")
		if err != nil {
			t.Fatalf("Failed to check set membership: %v", err)
		}

		if !isMember {
			t.Error("Member should be in set")
		}

		// Test SCard
		count, err := redisClient.SCard(ctx, setKey)
		if err != nil {
			t.Fatalf("Failed to get set cardinality: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected set cardinality to be 1, got %d", count)
		}

		// Test SMembers
		members, err := redisClient.SMembers(ctx, setKey)
		if err != nil {
			t.Fatalf("Failed to get set members: %v", err)
		}

		if len(members) != 1 {
			t.Errorf("Expected 1 member, got %d", len(members))
		}

		// Test SRem
		err = redisClient.SRem(ctx, setKey, "member1")
		if err != nil {
			t.Fatalf("Failed to remove from set: %v", err)
		}

		isMember, err = redisClient.SIsMember(ctx, setKey, "member1")
		if err != nil {
			t.Fatalf("Failed to check set membership after removal: %v", err)
		}

		if isMember {
			t.Error("Member should not be in set after removal")
		}
	})

	// Test Sorted Set operations
	t.Run("SortedSetOperations", func(t *testing.T) {
		zsetKey := "test_zset"

		// Test ZAdd
		err := redisClient.ZAdd(ctx, zsetKey, 1.0, "member1")
		if err != nil {
			t.Fatalf("Failed to add to sorted set: %v", err)
		}

		// Test ZIncrBy
		newScore, err := redisClient.ZIncrBy(ctx, zsetKey, 2.0, "member1")
		if err != nil {
			t.Fatalf("Failed to increment sorted set member: %v", err)
		}

		if newScore != 3.0 {
			t.Errorf("Expected score to be 3.0, got %f", newScore)
		}

		// Test ZCard
		count, err := redisClient.ZCard(ctx, zsetKey)
		if err != nil {
			t.Fatalf("Failed to get sorted set cardinality: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected sorted set cardinality to be 1, got %d", count)
		}

		// Test ZRange
		members, err := redisClient.ZRange(ctx, zsetKey, 0, -1)
		if err != nil {
			t.Fatalf("Failed to get sorted set range: %v", err)
		}

		if len(members) != 1 {
			t.Errorf("Expected 1 member, got %d", len(members))
		}

		// Test ZRem
		err = redisClient.ZRem(ctx, zsetKey, "member1")
		if err != nil {
			t.Fatalf("Failed to remove from sorted set: %v", err)
		}

		count, err = redisClient.ZCard(ctx, zsetKey)
		if err != nil {
			t.Fatalf("Failed to get sorted set cardinality after removal: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected sorted set cardinality to be 0, got %d", count)
		}
	})

	// Test Health check
	t.Run("HealthCheck", func(t *testing.T) {
		err := redisClient.Health(ctx)
		if err != nil {
			t.Fatalf("Redis health check failed: %v", err)
		}
	})
}
