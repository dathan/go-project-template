package tools_test

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dathan/go-project-template/internal/agent/tools"
)

var _ = Describe("FileReadTool", func() {
	var (
		readTool  *tools.FileReadTool
		writeTool *tools.FileWriteTool
		tmpDir    string
		ctx       context.Context
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "fstool-test-*")
		Expect(err).NotTo(HaveOccurred())
		readTool = tools.NewFileReadTool(tmpDir)
		writeTool = tools.NewFileWriteTool(tmpDir)
		ctx = context.Background()
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("FileWriteTool", func() {
		It("writes content to a file", func() {
			path := filepath.Join(tmpDir, "hello.txt")
			out, err := writeTool.Execute(ctx, map[string]any{
				"path":    path,
				"content": "hello world",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("hello.txt"))

			b, _ := os.ReadFile(path)
			Expect(string(b)).To(Equal("hello world"))
		})

		It("rejects paths outside the sandbox root", func() {
			_, err := writeTool.Execute(ctx, map[string]any{
				"path":    "/tmp/escape.txt",
				"content": "evil",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("escapes sandbox"))
		})
	})

	Describe("FileReadTool", func() {
		It("reads a file written by FileWriteTool", func() {
			path := filepath.Join(tmpDir, "data.txt")
			_, err := writeTool.Execute(ctx, map[string]any{"path": path, "content": "data"})
			Expect(err).NotTo(HaveOccurred())

			out, err := readTool.Execute(ctx, map[string]any{"path": path})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal("data"))
		})

		It("lists a directory", func() {
			Expect(os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte(""), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte(""), 0600)).To(Succeed())

			out, err := readTool.Execute(ctx, map[string]any{"path": tmpDir})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("a.txt"))
			Expect(out).To(ContainSubstring("b.txt"))
		})
	})
})
