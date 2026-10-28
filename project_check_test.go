//go:build projectcheck

package wf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

const projectCheckChildEnv = "GO_WEB_FRAME_PROJECT_CHECK_CHILD"

// TestEntireProject runs every Go module's complete test suite, race detector,
// and vet checks. It is opt-in so normal test runs do not execute themselves
// recursively.
//
// Run with:
//
//	go test -tags projectcheck -run '^TestEntireProject$' -v .
func TestEntireProject(t *testing.T) {
	if os.Getenv(projectCheckChildEnv) == "1" {
		t.Skip("nested project check")
	}

	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	modules, err := findGoModules(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) == 0 {
		t.Fatal("no go.mod files found")
	}

	commands := []struct {
		name string
		args []string
	}{
		{name: "test", args: []string{"test", "./...", "-count=1", "-shuffle=on"}},
		{name: "race", args: []string{"test", "-race", "./...", "-count=1", "-shuffle=on"}},
		{name: "vet", args: []string{"vet", "./..."}},
		{name: "tidy", args: []string{"mod", "tidy", "-diff"}},
	}

	for _, module := range modules {
		module := module
		label := moduleLabel(root, module)
		for _, command := range commands {
			command := command
			t.Run(label+"/"+command.name, func(t *testing.T) {
				runProjectGoCommand(t, module, command.args...)
			})
		}
	}
}

func findGoModules(root string) ([]string, error) {
	modules := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == "go.mod" {
			modules = append(modules, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(modules)
	return modules, nil
}

func moduleLabel(root string, module string) string {
	relative, err := filepath.Rel(root, module)
	if err != nil || relative == "." {
		return "root"
	}
	return strings.ReplaceAll(relative, string(filepath.Separator), "/")
}

func runProjectGoCommand(t *testing.T, directory string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), projectCheckChildEnv+"=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("go %s timed out in %s", strings.Join(args, " "), directory)
	}
	if err != nil {
		if isTidyDiffCommand(args) && tidyDiffOnlyChangesLineEndings(output) {
			t.Log("ignored CRLF-only go.mod/go.sum differences")
			return
		}
		t.Fatalf(
			"go %s failed in %s: %v\n%s",
			strings.Join(args, " "),
			directory,
			err,
			projectCheckOutput(output),
		)
	}
}

func isTidyDiffCommand(args []string) bool {
	return len(args) == 3 && args[0] == "mod" && args[1] == "tidy" && args[2] == "-diff"
}

func tidyDiffOnlyChangesLineEndings(output []byte) bool {
	normalized := strings.ReplaceAll(string(output), "\r\n", "\n")
	removed := make([]string, 0)
	added := make([]string, 0)
	for _, line := range strings.Split(normalized, "\n") {
		switch {
		case strings.HasPrefix(line, "---"), strings.HasPrefix(line, "+++"):
			continue
		case strings.HasPrefix(line, "-"):
			removed = append(removed, strings.TrimPrefix(line, "-"))
		case strings.HasPrefix(line, "+"):
			added = append(added, strings.TrimPrefix(line, "+"))
		}
	}
	return len(removed) > 0 && strings.Join(removed, "\n") == strings.Join(added, "\n")
}

func projectCheckOutput(output []byte) string {
	const limit = 16000
	if len(output) <= limit {
		return string(output)
	}
	return "... output truncated ...\n" + string(output[len(output)-limit:])
}
