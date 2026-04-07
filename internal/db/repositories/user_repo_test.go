package repositories_test

import (
	"context"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dathan/go-project-template/internal/db/models"
	"github.com/dathan/go-project-template/internal/db/repositories"
)

// Uses CGO-free SQLite in-memory for fast unit tests (no Postgres required).
// Integration tests against a real Postgres container live in test/integration/.
var _ = Describe("UserRepo", func() {
	var (
		repo *repositories.UserRepo
		ctx  context.Context
	)

	BeforeEach(func() {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(db.AutoMigrate(&models.User{})).To(Succeed())
		repo = repositories.NewUserRepo(db)
		ctx = context.Background()
	})

	Describe("Create and GetByID", func() {
		It("persists a user and retrieves it by ID", func() {
			u := &models.User{
				Email:      "alice@example.com",
				Name:       "Alice",
				Provider:   "github",
				ProviderID: "gh-001",
				Role:       models.RoleUser,
			}
			Expect(repo.Create(ctx, u)).To(Succeed())
			Expect(u.ID).NotTo(Equal(uuid.Nil))

			got, err := repo.GetByID(ctx, u.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Email).To(Equal("alice@example.com"))
		})

		It("returns an error for a non-existent ID", func() {
			_, err := repo.GetByID(ctx, uuid.New())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetByEmail", func() {
		It("finds a user by email", func() {
			u := &models.User{Email: "bob@example.com", Provider: "google", ProviderID: "g-002", Role: models.RoleUser}
			Expect(repo.Create(ctx, u)).To(Succeed())

			got, err := repo.GetByEmail(ctx, "bob@example.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(got.ID).To(Equal(u.ID))
		})
	})

	Describe("SetPaidAt", func() {
		It("marks a user as paid", func() {
			u := &models.User{Email: "carol@example.com", Provider: "slack", ProviderID: "sl-003", Role: models.RoleUser}
			Expect(repo.Create(ctx, u)).To(Succeed())

			now := time.Now().Truncate(time.Second)
			Expect(repo.SetPaidAt(ctx, u.ID, now)).To(Succeed())

			got, err := repo.GetByID(ctx, u.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.IsPaid()).To(BeTrue())
		})
	})

	Describe("List", func() {
		It("returns paginated users and total count", func() {
			for i := range 5 {
				u := &models.User{
					Email:      "user" + string(rune('a'+i)) + "@example.com",
					Provider:   "github",
					ProviderID: "id-" + string(rune('a'+i)),
					Role:       models.RoleUser,
				}
				Expect(repo.Create(ctx, u)).To(Succeed())
			}
			users, total, err := repo.List(ctx, models.ListOptions{Limit: 3, Offset: 0})
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(BeNumerically("==", 5))
			Expect(users).To(HaveLen(3))
		})
	})
})
