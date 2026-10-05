// Package test holds the Terratest suite for modules/pet-file.
//
// Each test copies the module to a temp dir (so tests can run in parallel without
// sharing .terraform or state), runs real terraform against it, and destroys whatever it
// created. Run with: cd test && go test -v -count=1 -timeout 10m ./...
package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/core/v2/files"
	"github.com/gruntwork-io/terratest/modules/core/v2/random"
	"github.com/gruntwork-io/terratest/modules/terraform/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const moduleDir = "../modules/pet-file"

// copyModule gives each test its own working copy of the module.
func copyModule(t *testing.T) string {
	t.Helper()
	dir, err := files.CopyTerraformFolderToTemp(moduleDir, t.Name())
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

// TestPetFileApply applies the module, checks its outputs and the file it wrote, then
// destroys it and checks the file is gone.
func TestPetFileApply(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	dir := copyModule(t)
	outDir := t.TempDir()
	prefix := "tt" + strings.ToLower(random.UniqueID())

	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: dir,
		Vars: map[string]any{
			"prefix":     prefix,
			"pet_length": 3,
			"output_dir": outDir,
		},
		NoColor: true,
	})

	destroyed := false
	defer func() {
		if !destroyed {
			terraform.DestroyContext(t, ctx, opts)
		}
	}()
	terraform.InitAndApplyContext(t, ctx, opts)

	name := terraform.OutputContext(t, ctx, opts, "name")
	assert.True(t, strings.HasPrefix(name, prefix+"-"), "name %q should start with %q", name, prefix+"-")
	assert.Len(t, strings.Split(name, "-"), 4, "prefix + 3 words")

	path := terraform.OutputContext(t, ctx, opts, "file_path")
	assert.Equal(t, filepath.Join(outDir, name+".txt"), path)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "Hello, "+name+"!\n", string(content))

	// A second apply must change nothing.
	assert.Equal(t, 0, terraform.InitAndPlanWithExitCodeContext(t, ctx, opts), "plan after apply should be empty")

	terraform.DestroyContext(t, ctx, opts)
	destroyed = true
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "destroy should remove %s", path)
}

// TestPetFileRejectsBadLength checks the variable validation without creating anything.
func TestPetFileRejectsBadLength(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	opts := &terraform.Options{
		TerraformDir: copyModule(t),
		Vars:         map[string]any{"pet_length": 9},
		NoColor:      true,
	}
	_, err := terraform.InitAndPlanContextE(t, ctx, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pet_length must be between 1 and 5")
}
