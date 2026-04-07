package auth_test

import (
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dathan/go-project-template/internal/auth"
)

var _ = Describe("JWTService", func() {
	var svc *auth.JWTService

	BeforeEach(func() {
		svc = auth.NewJWTService("test-secret-32-chars-long-enough!", 1*time.Hour)
	})

	Describe("Sign and Verify", func() {
		It("returns valid claims for a signed token", func() {
			id := uuid.New()
			token, err := svc.Sign(id, "user")
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			claims, err := svc.Verify(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(claims.UserID).To(Equal(id))
			Expect(claims.Role).To(Equal("user"))
			Expect(claims.Impersonating).To(BeNil())
		})

		It("rejects a tampered token", func() {
			token, _ := svc.Sign(uuid.New(), "user")
			_, err := svc.Verify(token + "tampered")
			Expect(err).To(HaveOccurred())
		})

		It("rejects a token signed with a different secret", func() {
			other := auth.NewJWTService("different-secret-32-chars-long!!", 1*time.Hour)
			token, _ := other.Sign(uuid.New(), "user")
			_, err := svc.Verify(token)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("SignImpersonation", func() {
		It("embeds admin and target IDs in claims", func() {
			adminID := uuid.New()
			targetID := uuid.New()
			token, err := svc.SignImpersonation(adminID, targetID, "user")
			Expect(err).NotTo(HaveOccurred())

			claims, err := svc.Verify(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(claims.UserID).To(Equal(targetID))
			Expect(claims.AdminID).NotTo(BeNil())
			Expect(*claims.AdminID).To(Equal(adminID))
			Expect(claims.Impersonating).NotTo(BeNil())
		})

		It("expires sooner than a regular token", func() {
			longSvc := auth.NewJWTService("test-secret-32-chars-long-enough!", 24*time.Hour)
			adminID := uuid.New()
			targetID := uuid.New()
			token, _ := longSvc.SignImpersonation(adminID, targetID, "user")
			claims, err := longSvc.Verify(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(claims.ExpiresAt.Time).To(BeTemporally("~", time.Now().Add(1*time.Hour), 5*time.Second))
		})
	})
})
