package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/auth/internal/adapters/database"
	"github.com/par1ram/silence/api/auth/internal/domain"
	"go.uber.org/zap"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Suite")
}

var _ = Describe("PostgresRepository", func() {
	var (
		mockDB sqlmock.Sqlmock
		db     *sql.DB
		repo   *database.PostgresRepository
		ctx    context.Context
		logger *zap.Logger
	)

	BeforeEach(func() {
		var err error
		db, mockDB, err = sqlmock.New()
		Expect(err).To(BeNil())

		ctx = context.Background()
		logger = zap.NewNop()
		repo = database.NewPostgresRepository(db, logger).(*database.PostgresRepository)
	})

	AfterEach(func() {
		db.Close()
	})

	Describe("Create", func() {
		It("should create user successfully", func() {
			user := &domain.User{
				ID:        "test-id",
				Email:     "test@example.com",
				Password:  "hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockDB.ExpectExec("INSERT INTO users").
				WithArgs(user.ID, user.Email, user.Password, user.Role, user.Status, user.CreatedAt, user.UpdatedAt).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.Create(ctx, user)
			Expect(err).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return error when database fails", func() {
			user := &domain.User{
				ID:        "test-id",
				Email:     "test@example.com",
				Password:  "hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockDB.ExpectExec("INSERT INTO users").
				WithArgs(user.ID, user.Email, user.Password, user.Role, user.Status, user.CreatedAt, user.UpdatedAt).
				WillReturnError(sql.ErrConnDone)

			err := repo.Create(ctx, user)
			Expect(err).To(Not(BeNil()))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("GetByEmail", func() {
		It("should return user when found", func() {
			email := "test@example.com"
			expectedUser := &domain.User{
				ID:        "test-id",
				Email:     email,
				Password:  "hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "status", "created_at", "updated_at"}).
				AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.Status, expectedUser.CreatedAt, expectedUser.UpdatedAt)

			mockDB.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
				WithArgs(email).
				WillReturnRows(rows)

			user, err := repo.GetByEmail(ctx, email)
			Expect(err).To(BeNil())
			Expect(user).To(Equal(expectedUser))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return error when user not found", func() {
			email := "nonexistent@example.com"

			mockDB.ExpectQuery("SELECT (.+) FROM users WHERE email = \\$1").
				WithArgs(email).
				WillReturnError(sql.ErrNoRows)

			user, err := repo.GetByEmail(ctx, email)
			Expect(err).To(Not(BeNil()))
			Expect(user).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("GetByID", func() {
		It("should return user when found", func() {
			id := "test-id"
			expectedUser := &domain.User{
				ID:        id,
				Email:     "test@example.com",
				Password:  "hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "status", "created_at", "updated_at"}).
				AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.Role, expectedUser.Status, expectedUser.CreatedAt, expectedUser.UpdatedAt)

			mockDB.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
				WithArgs(id).
				WillReturnRows(rows)

			user, err := repo.GetByID(ctx, id)
			Expect(err).To(BeNil())
			Expect(user).To(Equal(expectedUser))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return error when user not found", func() {
			id := "nonexistent-id"

			mockDB.ExpectQuery("SELECT (.+) FROM users WHERE id = \\$1").
				WithArgs(id).
				WillReturnError(sql.ErrNoRows)

			user, err := repo.GetByID(ctx, id)
			Expect(err).To(Not(BeNil()))
			Expect(user).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("Update", func() {
		It("should update user successfully", func() {
			user := &domain.User{
				ID:        "test-id",
				Email:     "new@example.com",
				Password:  "new_hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleAdmin,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockDB.ExpectExec("UPDATE users").
				WithArgs(user.ID, user.Email, user.Role, user.Status, user.UpdatedAt).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.Update(ctx, user)
			Expect(err).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return error when user not found", func() {
			user := &domain.User{
				ID:        "nonexistent-id",
				Email:     "test@example.com",
				Password:  "hashed_password",
				Status:    domain.StatusActive,
				Role:      domain.RoleUser,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			mockDB.ExpectExec("UPDATE users").
				WithArgs(user.ID, user.Email, user.Role, user.Status, user.UpdatedAt).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.Update(ctx, user)
			Expect(err).To(Not(BeNil()))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("Delete", func() {
		It("should delete user successfully", func() {
			id := "test-id"

			mockDB.ExpectExec("DELETE FROM users WHERE id = \\$1").
				WithArgs(id).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.Delete(ctx, id)
			Expect(err).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return error when user not found", func() {
			id := "nonexistent-id"

			mockDB.ExpectExec("DELETE FROM users WHERE id = \\$1").
				WithArgs(id).
				WillReturnResult(sqlmock.NewResult(0, 0))

			err := repo.Delete(ctx, id)
			Expect(err).To(Not(BeNil()))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("List", func() {
		It("should return users list successfully", func() {
			filter := &domain.UserFilter{
				Limit:  10,
				Offset: 0,
			}

			expectedUsers := []*domain.User{
				{
					ID:        "user1",
					Email:     "user1@example.com",
					Password:  "hash1",
					Status:    domain.StatusActive,
					Role:      domain.RoleUser,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					ID:        "user2",
					Email:     "user2@example.com",
					Password:  "hash2",
					Status:    domain.StatusActive,
					Role:      domain.RoleAdmin,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			// Сначала мок на count
			mockDB.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE 1=1`).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(len(expectedUsers)))

			// Затем мок на выборку пользователей
			rows := sqlmock.NewRows([]string{"id", "email", "password", "role", "status", "created_at", "updated_at"})
			for _, user := range expectedUsers {
				rows.AddRow(user.ID, user.Email, user.Password, user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
			}
			mockDB.ExpectQuery(`SELECT id, email, password, role, status, created_at, updated_at FROM users WHERE 1=1 ORDER BY created_at DESC LIMIT \$1`).
				WillReturnRows(rows)

			users, total, err := repo.List(ctx, filter)
			Expect(err).To(BeNil())
			Expect(users).To(Equal(expectedUsers))
			Expect(total).To(Equal(2))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})

		It("should return empty list when no users", func() {
			filter := &domain.UserFilter{
				Limit:  10,
				Offset: 0,
			}

			mockDB.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE 1=1`).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

			mockDB.ExpectQuery(`SELECT id, email, password, role, status, created_at, updated_at FROM users WHERE 1=1 ORDER BY created_at DESC LIMIT \$1`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "status", "created_at", "updated_at"}))

			users, total, err := repo.List(ctx, filter)
			Expect(err).To(BeNil())
			Expect(users).To(BeEmpty())
			Expect(total).To(Equal(0))
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("UpdateStatus", func() {
		It("should update user status successfully", func() {
			id := "test-id"
			status := domain.StatusBlocked

			mockDB.ExpectExec(`UPDATE users SET status = \$2, updated_at = NOW\(\) WHERE id = \$1`).
				WithArgs(id, status).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.UpdateStatus(ctx, id, status)
			Expect(err).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})

	Describe("UpdateRole", func() {
		It("should update user role successfully", func() {
			id := "test-id"
			role := domain.RoleAdmin

			mockDB.ExpectExec(`UPDATE users SET role = \$2, updated_at = NOW\(\) WHERE id = \$1`).
				WithArgs(id, role).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.UpdateRole(ctx, id, role)
			Expect(err).To(BeNil())
			Expect(mockDB.ExpectationsWereMet()).To(BeNil())
		})
	})
})
