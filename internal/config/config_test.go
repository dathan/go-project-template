package config_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dathan/go-project-template/internal/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Load", func() {
	// Tests run from the package directory (internal/config), so we need to
	// point the working directory to the repo root where conf/ lives.
	BeforeEach(func() {
		Expect(os.Chdir("../..")).To(Succeed())
	})
	AfterEach(func() {
		// Return to original dir so other tests are unaffected.
		Expect(os.Chdir("internal/config")).To(Succeed())
	})

	It("loads conf/config.yaml and applies defaults", func() {
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())

		// Values from conf/config.yaml
		Expect(cfg.Server.Port).To(Equal(8080))
		Expect(cfg.Database.Host).To(Equal("127.0.0.1"))
		Expect(cfg.Database.Port).To(Equal(5432))
		Expect(cfg.Agent.Provider).To(Equal("claude"))
	})

	It("env var DATABASE_HOST overrides conf/config.yaml", func() {
		Expect(os.Setenv("DATABASE_HOST", "prod.db.internal")).To(Succeed())
		defer os.Unsetenv("DATABASE_HOST")

		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Database.Host).To(Equal("prod.db.internal"))
	})

	It("does not error when .env is absent", func() {
		// The project root has no .env (only .env.example) — Load should not fail.
		_, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
	})
})

// Simulates invoking the binary from a directory that has no conf/config.yaml
// (e.g. running ./bin/server from the bin/ directory). Hard-coded defaults must
// be used, and the DSN must not contain empty database/user fields.
var _ = Describe("Load (no config file — bin/ scenario)", func() {
	var origDir, tmpDir string

	BeforeEach(func() {
		var err error
		origDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		tmpDir, err = os.MkdirTemp("", "config-no-file-*")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(tmpDir)).To(Succeed())

		// Clear any env vars that could shadow the defaults.
		for _, k := range []string{
			"DATABASE_NAME", "DATABASE_USER", "DATABASE_HOST",
			"DATABASE_PASSWORD", "DATABASE_PORT",
		} {
			Expect(os.Unsetenv(k)).To(Succeed())
		}
	})

	AfterEach(func() {
		Expect(os.Chdir(origDir)).To(Succeed())
		os.RemoveAll(tmpDir)
	})

	It("applies hard-coded defaults when conf/config.yaml is absent", func() {
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())

		Expect(cfg.Database.Name).To(Equal("app_db"),
			"Database.Name must default to app_db — empty name causes DSN to connect to wrong database")
		Expect(cfg.Database.User).To(Equal("app_user"))
		Expect(cfg.Database.Host).To(Equal("127.0.0.1"))
		Expect(cfg.Server.Port).To(Equal(8080))
	})

	It("DSN contains the default database name", func() {
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())

		dsn := cfg.Database.DSN()
		Expect(dsn).To(ContainSubstring("dbname=app_db"),
			"DSN must not have an empty dbname= field")
		Expect(dsn).To(ContainSubstring("user=app_user"))
	})
})
