package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/par1ram/silence/api/auth/internal/domain"
	"github.com/par1ram/silence/api/auth/internal/mocks"
	"github.com/par1ram/silence/api/auth/internal/ports"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = Describe("AuthService", func() {
	var (
		ctrl         *gomock.Controller
		mockUserRepo *mocks.MockUserRepository
		mockHasher   *mocks.MockPasswordHasher
		mockTokenGen *mocks.MockTokenGenerator
		authService  ports.AuthService
		logger       *zap.Logger
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockUserRepo = mocks.NewMockUserRepository(ctrl)
		mockHasher = mocks.NewMockPasswordHasher(ctrl)
		mockTokenGen = mocks.NewMockTokenGenerator(ctrl)
		logger = zap.NewNop()
		authService = NewAuthService(mockUserRepo, mockHasher, mockTokenGen, logger)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Register", func() {
		It("should register user successfully", func() {
			req := &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			// Мокаем проверку существования пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(nil, domain.ErrUserNotFound)

			// Мокаем хеширование пароля
			hashedPassword := "hashed_password_123"
			mockHasher.EXPECT().
				Hash(req.Password).
				Return(hashedPassword, nil)

			// Мокаем создание пользователя
			mockUserRepo.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, user *domain.User) error {
					Expect(user.Email).To(Equal(req.Email))
					Expect(user.Password).To(Equal(hashedPassword))
					Expect(user.Role).To(Equal(domain.RoleUser))
					Expect(user.Status).To(Equal(domain.StatusActive))
					Expect(user.ID).To(Not(BeEmpty()))
					return nil
				})

			// Мокаем генерацию токена
			expectedToken := "jwt_token_here"
			mockTokenGen.EXPECT().
				GenerateToken(gomock.Any()).
				Return(expectedToken, nil)

			// Выполняем регистрацию
			response, err := authService.Register(context.Background(), req)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(response).NotTo(BeNil())
			Expect(response.Token).To(Equal(expectedToken))
			Expect(response.User.Email).To(Equal(req.Email))
			Expect(response.User.Role).To(Equal(domain.RoleUser))
			Expect(response.User.Status).To(Equal(domain.StatusActive))
		})

		It("should return error when user already exists", func() {
			req := &domain.RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			}

			existingUser := &domain.User{
				ID:       "existing_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusActive,
			}

			// Мокаем проверку существования пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(existingUser, nil)

			// Выполняем регистрацию
			response, err := authService.Register(context.Background(), req)

			// Проверяем результат
			Expect(err).To(Equal(domain.ErrUserAlreadyExists))
			Expect(response).To(BeNil())
		})

		It("should return error when password hashing fails", func() {
			req := &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			// Мокаем проверку существования пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(nil, domain.ErrUserNotFound)

			// Мокаем ошибку хеширования пароля
			mockHasher.EXPECT().
				Hash(req.Password).
				Return("", domain.ErrInvalidCredentials)

			// Выполняем регистрацию
			response, err := authService.Register(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to hash password"))
			Expect(response).To(BeNil())
		})

		It("should return error when user creation fails", func() {
			req := &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			// Мокаем проверку существования пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(nil, domain.ErrUserNotFound)

			// Мокаем хеширование пароля
			hashedPassword := "hashed_password_123"
			mockHasher.EXPECT().
				Hash(req.Password).
				Return(hashedPassword, nil)

			// Мокаем ошибку создания пользователя
			mockUserRepo.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(domain.ErrInvalidCredentials)

			// Выполняем регистрацию
			response, err := authService.Register(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create user"))
			Expect(response).To(BeNil())
		})

		It("should return error when token generation fails", func() {
			req := &domain.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			// Мокаем проверку существования пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(nil, domain.ErrUserNotFound)

			// Мокаем хеширование пароля
			hashedPassword := "hashed_password_123"
			mockHasher.EXPECT().
				Hash(req.Password).
				Return(hashedPassword, nil)

			// Мокаем создание пользователя
			mockUserRepo.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil)

			// Мокаем ошибку генерации токена
			mockTokenGen.EXPECT().
				GenerateToken(gomock.Any()).
				Return("", domain.ErrInvalidCredentials)

			// Выполняем регистрацию
			response, err := authService.Register(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to generate token"))
			Expect(response).To(BeNil())
		})
	})

	Describe("Login", func() {
		It("should login user successfully", func() {
			req := &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			user := &domain.User{
				ID:       "user_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusActive,
			}

			// Мокаем получение пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(user, nil)

			// Мокаем проверку пароля
			mockHasher.EXPECT().
				Verify(req.Password, user.Password).
				Return(true)

			// Мокаем генерацию токена
			expectedToken := "jwt_token_here"
			mockTokenGen.EXPECT().
				GenerateToken(user).
				Return(expectedToken, nil)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(response).NotTo(BeNil())
			Expect(response.Token).To(Equal(expectedToken))
			Expect(response.User).To(Equal(user))
		})

		It("should return error when user not found", func() {
			req := &domain.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			}

			// Мокаем ошибку получения пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(nil, domain.ErrUserNotFound)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(Equal(domain.ErrInvalidCredentials))
			Expect(response).To(BeNil())
		})

		It("should return error when user is blocked", func() {
			req := &domain.LoginRequest{
				Email:    "blocked@example.com",
				Password: "password123",
			}

			user := &domain.User{
				ID:       "user_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusBlocked,
			}

			// Мокаем получение пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(user, nil)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("user account is blocked"))
			Expect(response).To(BeNil())
		})

		It("should return error when user is inactive", func() {
			req := &domain.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			}

			user := &domain.User{
				ID:       "user_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusInactive,
			}

			// Мокаем получение пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(user, nil)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("user account is inactive"))
			Expect(response).To(BeNil())
		})

		It("should return error when password is incorrect", func() {
			req := &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "wrong_password",
			}

			user := &domain.User{
				ID:       "user_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusActive,
			}

			// Мокаем получение пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(user, nil)

			// Мокаем неправильную проверку пароля
			mockHasher.EXPECT().
				Verify(req.Password, user.Password).
				Return(false)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(Equal(domain.ErrInvalidCredentials))
			Expect(response).To(BeNil())
		})

		It("should return error when token generation fails", func() {
			req := &domain.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			user := &domain.User{
				ID:       "user_id",
				Email:    req.Email,
				Password: "hashed_password",
				Role:     domain.RoleUser,
				Status:   domain.StatusActive,
			}

			// Мокаем получение пользователя
			mockUserRepo.EXPECT().
				GetByEmail(gomock.Any(), req.Email).
				Return(user, nil)

			// Мокаем проверку пароля
			mockHasher.EXPECT().
				Verify(req.Password, user.Password).
				Return(true)

			// Мокаем ошибку генерации токена
			mockTokenGen.EXPECT().
				GenerateToken(user).
				Return("", domain.ErrInvalidCredentials)

			// Выполняем вход
			response, err := authService.Login(context.Background(), req)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to generate token"))
			Expect(response).To(BeNil())
		})
	})

	Describe("ValidateToken", func() {
		It("should validate token successfully", func() {
			token := "valid_jwt_token"
			expectedClaims := &domain.Claims{
				UserID: "user_id",
				Email:  "test@example.com",
				Role:   domain.RoleUser,
			}

			// Мокаем валидацию токена
			mockTokenGen.EXPECT().
				ValidateToken(token).
				Return(expectedClaims, nil)

			// Выполняем валидацию
			claims, err := authService.ValidateToken(token)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(claims).To(Equal(expectedClaims))
		})

		It("should return error when token is invalid", func() {
			token := "invalid_jwt_token"

			// Мокаем ошибку валидации токена
			mockTokenGen.EXPECT().
				ValidateToken(token).
				Return(nil, domain.ErrInvalidCredentials)

			// Выполняем валидацию
			claims, err := authService.ValidateToken(token)

			// Проверяем результат
			Expect(err).To(Equal(domain.ErrInvalidCredentials))
			Expect(claims).To(BeNil())
		})
	})
})

var _ = Describe("TokenService", func() {
	var (
		tokenService ports.TokenGenerator
		secretKey    string
		expiresIn    time.Duration
		testUser     *domain.User
	)

	BeforeEach(func() {
		secretKey = "test_secret_key_123"
		expiresIn = 24 * time.Hour
		tokenService = NewTokenService(secretKey, expiresIn)
		testUser = &domain.User{
			ID:    "user_123",
			Email: "test@example.com",
			Role:  domain.RoleUser,
		}
	})

	Describe("GenerateToken", func() {
		It("should generate valid JWT token", func() {
			// Генерируем токен
			token, err := tokenService.GenerateToken(testUser)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(token).NotTo(BeEmpty())
			Expect(len(token)).To(BeNumerically(">", 100)) // JWT токены обычно длинные
		})

		It("should generate token with correct claims", func() {
			// Генерируем токен
			token, err := tokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Валидируем токен и получаем claims
			claims, err := tokenService.ValidateToken(token)
			Expect(err).To(BeNil())

			// Проверяем claims
			Expect(claims.UserID).To(Equal(testUser.ID))
			Expect(claims.Email).To(Equal(testUser.Email))
			Expect(claims.Role).To(Equal(testUser.Role))
			Expect(claims.Issuer).To(Equal("silence-vpn"))
			Expect(claims.Subject).To(Equal(testUser.ID))
		})

		It("should generate token with admin role", func() {
			adminUser := &domain.User{
				ID:    "admin_123",
				Email: "admin@example.com",
				Role:  domain.RoleAdmin,
			}

			// Генерируем токен
			token, err := tokenService.GenerateToken(adminUser)
			Expect(err).To(BeNil())

			// Валидируем токен и получаем claims
			claims, err := tokenService.ValidateToken(token)
			Expect(err).To(BeNil())

			// Проверяем claims
			Expect(claims.Role).To(Equal(domain.RoleAdmin))
		})

		It("should generate token with custom expiration", func() {
			customExpiresIn := 1 * time.Hour
			customTokenService := NewTokenService(secretKey, customExpiresIn)

			// Генерируем токен
			token, err := customTokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Валидируем токен и получаем claims
			claims, err := customTokenService.ValidateToken(token)
			Expect(err).To(BeNil())

			// Проверяем, что токен не истек
			Expect(claims.ExpiresAt.Time).To(BeTemporally(">", time.Now()))
			Expect(claims.ExpiresAt.Time).To(BeTemporally("<", time.Now().Add(2*time.Hour)))
		})
	})

	Describe("ValidateToken", func() {
		It("should validate valid token", func() {
			// Генерируем токен
			token, err := tokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Валидируем токен
			claims, err := tokenService.ValidateToken(token)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(claims).NotTo(BeNil())
			Expect(claims.UserID).To(Equal(testUser.ID))
			Expect(claims.Email).To(Equal(testUser.Email))
		})

		It("should return error for invalid token", func() {
			invalidToken := "invalid.jwt.token"

			// Валидируем неверный токен
			claims, err := tokenService.ValidateToken(invalidToken)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse token"))
			Expect(claims).To(BeNil())
		})

		It("should return error for empty token", func() {
			// Валидируем пустой токен
			claims, err := tokenService.ValidateToken("")

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse token"))
			Expect(claims).To(BeNil())
		})

		It("should return error for token with wrong signature", func() {
			// Создаем токен с другим секретным ключом
			otherTokenService := NewTokenService("different_secret_key", expiresIn)
			token, err := otherTokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Пытаемся валидировать с оригинальным сервисом
			claims, err := tokenService.ValidateToken(token)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse token"))
			Expect(claims).To(BeNil())
		})

		It("should return error for expired token", func() {
			// Создаем сервис с очень коротким временем жизни токена
			shortExpiresIn := 1 * time.Millisecond
			shortTokenService := NewTokenService(secretKey, shortExpiresIn)

			// Генерируем токен
			token, err := shortTokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Ждем, пока токен истечет
			time.Sleep(10 * time.Millisecond)

			// Пытаемся валидировать истекший токен
			claims, err := shortTokenService.ValidateToken(token)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse token"))
			Expect(claims).To(BeNil())
		})

		It("should return error for malformed token", func() {
			malformedToken := "not.a.valid.jwt.token.structure"

			// Валидируем неправильно сформированный токен
			claims, err := tokenService.ValidateToken(malformedToken)

			// Проверяем результат
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse token"))
			Expect(claims).To(BeNil())
		})
	})

	Describe("Token roundtrip", func() {
		It("should generate and validate token successfully", func() {
			// Генерируем токен
			token, err := tokenService.GenerateToken(testUser)
			Expect(err).To(BeNil())

			// Валидируем токен
			claims, err := tokenService.ValidateToken(token)
			Expect(err).To(BeNil())

			// Проверяем, что все данные сохранились
			Expect(claims.UserID).To(Equal(testUser.ID))
			Expect(claims.Email).To(Equal(testUser.Email))
			Expect(claims.Role).To(Equal(testUser.Role))
			Expect(claims.Issuer).To(Equal("silence-vpn"))
			Expect(claims.Subject).To(Equal(testUser.ID))
			Expect(claims.ExpiresAt.Time).To(BeTemporally(">", time.Now()))
		})

		It("should work with different users", func() {
			users := []*domain.User{
				{ID: "user1", Email: "user1@example.com", Role: domain.RoleUser},
				{ID: "user2", Email: "user2@example.com", Role: domain.RoleAdmin},
				{ID: "user3", Email: "user3@example.com", Role: domain.RoleUser},
			}

			for _, user := range users {
				// Генерируем токен
				token, err := tokenService.GenerateToken(user)
				Expect(err).To(BeNil())

				// Валидируем токен
				claims, err := tokenService.ValidateToken(token)
				Expect(err).To(BeNil())

				// Проверяем данные
				Expect(claims.UserID).To(Equal(user.ID))
				Expect(claims.Email).To(Equal(user.Email))
				Expect(claims.Role).To(Equal(user.Role))
			}
		})
	})
})

var _ = Describe("PasswordService", func() {
	var (
		passwordService ports.PasswordHasher
	)

	BeforeEach(func() {
		passwordService = NewPasswordService()
	})

	Describe("Hash", func() {
		It("should hash password successfully", func() {
			password := "test_password_123"

			// Хешируем пароль
			hash, err := passwordService.Hash(password)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(hash).NotTo(BeEmpty())
			Expect(hash).NotTo(Equal(password))          // Хеш не должен быть равен исходному паролю
			Expect(len(hash)).To(BeNumerically(">", 50)) // bcrypt хеши обычно длинные
		})

		It("should generate different hashes for same password", func() {
			password := "same_password"

			// Хешируем пароль дважды
			hash1, err1 := passwordService.Hash(password)
			hash2, err2 := passwordService.Hash(password)

			// Проверяем результат
			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(hash1).NotTo(Equal(hash2)) // Разные соли дают разные хеши
		})

		It("should handle empty password", func() {
			password := ""

			// Хешируем пустой пароль
			hash, err := passwordService.Hash(password)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(hash).NotTo(BeEmpty())
		})

		It("should handle special characters in password", func() {
			password := "p@ssw0rd!@#$%^&*()"

			// Хешируем пароль со спецсимволами
			hash, err := passwordService.Hash(password)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(hash).NotTo(BeEmpty())
			Expect(hash).NotTo(Equal(password))
		})

		It("should handle very long password", func() {
			password := "very_long_password_with_many_characters_to_test_bcrypt_behavior_with_extended_input_strings"

			// Хешируем длинный пароль
			hash, err := passwordService.Hash(password)

			// Проверяем результат - bcrypt имеет ограничение в 72 байта
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("password length exceeds 72 bytes"))
			Expect(hash).To(BeEmpty())
		})

		It("should handle unicode characters", func() {
			password := "пароль_с_русскими_символами_123"

			// Хешируем пароль с юникод символами
			hash, err := passwordService.Hash(password)

			// Проверяем результат
			Expect(err).To(BeNil())
			Expect(hash).NotTo(BeEmpty())
			Expect(hash).NotTo(Equal(password))
		})
	})

	Describe("Verify", func() {
		It("should verify correct password", func() {
			password := "test_password_123"

			// Хешируем пароль
			hash, err := passwordService.Hash(password)
			Expect(err).To(BeNil())

			// Проверяем правильный пароль
			isValid := passwordService.Verify(password, hash)

			// Проверяем результат
			Expect(isValid).To(BeTrue())
		})

		It("should reject incorrect password", func() {
			password := "test_password_123"
			wrongPassword := "wrong_password_456"

			// Хешируем пароль
			hash, err := passwordService.Hash(password)
			Expect(err).To(BeNil())

			// Проверяем неправильный пароль
			isValid := passwordService.Verify(wrongPassword, hash)

			// Проверяем результат
			Expect(isValid).To(BeFalse())
		})

		It("should reject empty password when hash exists", func() {
			password := "test_password_123"

			// Хешируем пароль
			hash, err := passwordService.Hash(password)
			Expect(err).To(BeNil())

			// Проверяем пустой пароль
			isValid := passwordService.Verify("", hash)

			// Проверяем результат
			Expect(isValid).To(BeFalse())
		})

		It("should handle empty hash", func() {
			password := "test_password_123"

			// Проверяем с пустым хешем
			isValid := passwordService.Verify(password, "")

			// Проверяем результат
			Expect(isValid).To(BeFalse())
		})

		It("should handle malformed hash", func() {
			password := "test_password_123"
			malformedHash := "not_a_valid_bcrypt_hash"

			// Проверяем с неправильным хешем
			isValid := passwordService.Verify(password, malformedHash)

			// Проверяем результат
			Expect(isValid).To(BeFalse())
		})

		It("should verify password with special characters", func() {
			password := "p@ssw0rd!@#$%^&*()"

			// Хешируем пароль со спецсимволами
			hash, err := passwordService.Hash(password)
			Expect(err).To(BeNil())

			// Проверяем правильный пароль
			isValid := passwordService.Verify(password, hash)

			// Проверяем результат
			Expect(isValid).To(BeTrue())
		})

		It("should verify password with unicode characters", func() {
			password := "пароль_с_русскими_символами_123"

			// Хешируем пароль с юникод символами
			hash, err := passwordService.Hash(password)
			Expect(err).To(BeNil())

			// Проверяем правильный пароль
			isValid := passwordService.Verify(password, hash)

			// Проверяем результат
			Expect(isValid).To(BeTrue())
		})
	})

	Describe("Hash and Verify roundtrip", func() {
		It("should work with multiple passwords", func() {
			passwords := []string{
				"simple_password",
				"p@ssw0rd!@#$%^&*()",
				"пароль_с_русскими_символами",
				"",
			}

			for _, password := range passwords {
				// Хешируем пароль
				hash, err := passwordService.Hash(password)
				Expect(err).To(BeNil())

				// Проверяем правильный пароль
				isValid := passwordService.Verify(password, hash)
				Expect(isValid).To(BeTrue())

				// Проверяем неправильный пароль
				isValid = passwordService.Verify(password+"_wrong", hash)
				Expect(isValid).To(BeFalse())
			}
		})

		It("should generate unique hashes for different passwords", func() {
			passwords := []string{
				"password1",
				"password2",
				"password3",
			}

			hashes := make(map[string]bool)

			for _, password := range passwords {
				// Хешируем пароль
				hash, err := passwordService.Hash(password)
				Expect(err).To(BeNil())

				// Проверяем, что хеш уникален
				Expect(hashes[hash]).To(BeFalse())
				hashes[hash] = true

				// Проверяем, что пароль валидируется
				isValid := passwordService.Verify(password, hash)
				Expect(isValid).To(BeTrue())
			}
		})
	})
})

var _ = Describe("UserService", func() {
	var (
		userService ports.UserService
		mockRepo    *mocks.MockUserRepository
		mockHasher  *mocks.MockPasswordHasher
		ctx         context.Context
		ctrl        *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockUserRepository(ctrl)
		mockHasher = mocks.NewMockPasswordHasher(ctrl)
		userService = NewUserService(mockRepo, mockHasher)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("CreateUser", func() {
		It("should create user successfully", func() {
			req := &domain.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     domain.RoleUser,
			}

			hashedPassword := "hashed_password"
			mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, nil)
			mockHasher.EXPECT().Hash(req.Password).Return(hashedPassword, nil)
			mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
				u.ID = "user-id"
				return nil
			})

			result, err := userService.CreateUser(ctx, req)

			Expect(err).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.ID).To(Equal("user-id"))
			Expect(result.Email).To(Equal(req.Email))
			Expect(result.Password).To(Equal(hashedPassword))
			Expect(result.Role).To(Equal(req.Role))
			Expect(result.Status).To(Equal(domain.StatusActive))
			Expect(result.CreatedAt).Should(BeTemporally("~", time.Now(), time.Second))
			Expect(result.UpdatedAt).Should(BeTemporally("~", time.Now(), time.Second))
		})

		It("should return error when password hashing fails", func() {
			req := &domain.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, nil)
			mockHasher.EXPECT().Hash(req.Password).Return("", errors.New("hashing error"))

			result, err := userService.CreateUser(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when user creation fails", func() {
			req := &domain.CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
			}

			mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, nil)
			hashedPassword := "hashed_password"
			mockHasher.EXPECT().Hash(req.Password).Return(hashedPassword, nil)
			mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("creation error"))

			result, err := userService.CreateUser(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("GetUser", func() {
		It("should get user successfully", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)

			result, err := userService.GetUser(ctx, userID)

			Expect(err).To(BeNil())
			Expect(result).To(Equal(user))
		})

		It("should return error when user not found", func() {
			userID := "user-id"

			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			result, err := userService.GetUser(ctx, userID)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("ListUsers", func() {
		It("should list users successfully", func() {
			filter := &domain.UserFilter{
				Limit:  10,
				Offset: 0,
			}

			users := []*domain.User{
				{ID: "user1", Email: "user1@example.com"},
				{ID: "user2", Email: "user2@example.com"},
			}
			total := 2

			mockRepo.EXPECT().List(ctx, filter).Return(users, total, nil)

			result, err := userService.ListUsers(ctx, filter)

			Expect(err).To(BeNil())
			Expect(result.Users).To(Equal(users))
			Expect(result.Total).To(Equal(total))
		})

		It("should return error when listing fails", func() {
			filter := &domain.UserFilter{
				Limit:  10,
				Offset: 0,
			}

			mockRepo.EXPECT().List(ctx, filter).Return(nil, 0, errors.New("list error"))

			result, err := userService.ListUsers(ctx, filter)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("UpdateUser", func() {
		It("should update user successfully", func() {
			userID := "user-id"
			req := &domain.UpdateUserRequest{
				Email:  "newemail@example.com",
				Role:   domain.RoleAdmin,
				Status: domain.StatusActive,
			}

			existingUser := &domain.User{
				ID:     userID,
				Email:  "oldemail@example.com",
				Role:   domain.RoleUser,
				Status: domain.StatusInactive,
			}

			updatedUser := &domain.User{
				ID:     userID,
				Email:  req.Email,
				Role:   req.Role,
				Status: req.Status,
			}

			mockRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
			mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(nil, nil)
			mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *domain.User) error {
				*u = *updatedUser
				return nil
			})

			result, err := userService.UpdateUser(ctx, userID, req)

			Expect(err).To(BeNil())
			Expect(result).To(Equal(updatedUser))
		})

		It("should return error when user not found", func() {
			userID := "user-id"
			req := &domain.UpdateUserRequest{
				Email: "newemail@example.com",
			}

			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			result, err := userService.UpdateUser(ctx, userID, req)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when email already exists", func() {
			userID := "user-id"
			req := &domain.UpdateUserRequest{
				Email: "existing@example.com",
			}

			existingUser := &domain.User{
				ID:    userID,
				Email: "oldemail@example.com",
			}

			existingUserWithEmail := &domain.User{
				ID:    "other-user-id",
				Email: req.Email,
			}

			mockRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
			mockRepo.EXPECT().GetByEmail(ctx, req.Email).Return(existingUserWithEmail, nil)

			result, err := userService.UpdateUser(ctx, userID, req)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("DeleteUser", func() {
		It("should delete user successfully", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем удаление пользователя
			mockRepo.EXPECT().Delete(ctx, userID).Return(nil)

			err := userService.DeleteUser(ctx, userID)

			Expect(err).To(BeNil())
		})

		It("should return error when deletion fails", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем ошибку удаления
			mockRepo.EXPECT().Delete(ctx, userID).Return(errors.New("deletion error"))

			err := userService.DeleteUser(ctx, userID)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userID := "user-id"

			// Мокаем ошибку получения пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			err := userService.DeleteUser(ctx, userID)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("пользователь не найден"))
		})
	})

	Describe("BlockUser", func() {
		It("should block user successfully", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем блокировку пользователя
			mockRepo.EXPECT().UpdateStatus(ctx, userID, domain.StatusBlocked).Return(nil)

			err := userService.BlockUser(ctx, userID)

			Expect(err).To(BeNil())
		})

		It("should return error when blocking fails", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем ошибку блокировки
			mockRepo.EXPECT().UpdateStatus(ctx, userID, domain.StatusBlocked).Return(errors.New("blocking error"))

			err := userService.BlockUser(ctx, userID)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userID := "user-id"

			// Мокаем ошибку получения пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			err := userService.BlockUser(ctx, userID)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("пользователь не найден"))
		})
	})

	Describe("UnblockUser", func() {
		It("should unblock user successfully", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем разблокировку пользователя
			mockRepo.EXPECT().UpdateStatus(ctx, userID, domain.StatusActive).Return(nil)

			err := userService.UnblockUser(ctx, userID)

			Expect(err).To(BeNil())
		})

		It("should return error when unblocking fails", func() {
			userID := "user-id"
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем ошибку разблокировки
			mockRepo.EXPECT().UpdateStatus(ctx, userID, domain.StatusActive).Return(errors.New("unblocking error"))

			err := userService.UnblockUser(ctx, userID)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userID := "user-id"

			// Мокаем ошибку получения пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			err := userService.UnblockUser(ctx, userID)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("пользователь не найден"))
		})
	})

	Describe("ChangeUserRole", func() {
		It("should change user role successfully", func() {
			userID := "user-id"
			newRole := domain.RoleAdmin
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем изменение роли
			mockRepo.EXPECT().UpdateRole(ctx, userID, newRole).Return(nil)

			err := userService.ChangeUserRole(ctx, userID, newRole)

			Expect(err).To(BeNil())
		})

		It("should return error when role change fails", func() {
			userID := "user-id"
			newRole := domain.RoleAdmin
			user := &domain.User{
				ID:    userID,
				Email: "test@example.com",
				Role:  domain.RoleUser,
			}

			// Мокаем проверку существования пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(user, nil)
			// Мокаем ошибку изменения роли
			mockRepo.EXPECT().UpdateRole(ctx, userID, newRole).Return(errors.New("role change error"))

			err := userService.ChangeUserRole(ctx, userID, newRole)

			Expect(err).To(HaveOccurred())
		})

		It("should return error when user not found", func() {
			userID := "user-id"
			newRole := domain.RoleAdmin

			// Мокаем ошибку получения пользователя
			mockRepo.EXPECT().GetByID(ctx, userID).Return(nil, errors.New("user not found"))

			err := userService.ChangeUserRole(ctx, userID, newRole)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("пользователь не найден"))
		})
	})
})
