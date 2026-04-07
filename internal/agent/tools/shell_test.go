package tools_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dathan/go-project-template/internal/agent/tools"
)

var _ = Describe("ShellTool", func() {
	var (
		tool *tools.ShellTool
		ctx  context.Context
	)

	BeforeEach(func() {
		tool = tools.NewShellTool(5 * time.Second)
		ctx = context.Background()
	})

	It("executes a simple command and returns stdout", func() {
		out, err := tool.Execute(ctx, map[string]any{"command": "echo hello"})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("hello"))
	})

	It("returns an error for a failing command", func() {
		_, err := tool.Execute(ctx, map[string]any{"command": "exit 1"})
		Expect(err).To(HaveOccurred())
	})

	It("rejects commands not in the allowlist", func() {
		restricted := tools.NewShellTool(5*time.Second, "echo", "ls")
		_, err := restricted.Execute(ctx, map[string]any{"command": "rm -rf /"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("allowlist"))
	})

	It("returns an error when command is missing", func() {
		_, err := tool.Execute(ctx, map[string]any{})
		Expect(err).To(HaveOccurred())
	})

	It("exposes correct tool metadata", func() {
		Expect(tool.Name()).To(Equal("shell"))
		Expect(tool.Description()).NotTo(BeEmpty())
		Expect(tool.Parameters()).To(HaveKey("properties"))
	})
})
