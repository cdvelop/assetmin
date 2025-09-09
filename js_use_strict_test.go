package assetmin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripUseStrictAndSingleOccurrence(t *testing.T) {
	// Provide a GetRuntimeInitializerJS that includes the 'use strict' we add globally
	initJS := func() (string, error) {
		// Return wasm init code without an additional 'use strict' directive.
		return "\n// wasm init code... WebAssembly.Memory", nil
	}

	env := setupTestEnv("js-use-strict", t, initJS)
	defer env.CleanDirectory()

	env.CreateModulesDir()
	env.CreateThemeDir()
	env.CreatePublicDir()

	// Create a file that already contains a leading use strict directive
	file1 := "a.js"
	path1 := filepath.Join(env.BaseDir, file1)
	require.NoError(t, os.WriteFile(path1, []byte("'use strict';\nconsole.log('A');"), 0644))

	// Create another file without the directive
	file2 := "b.js"
	path2 := filepath.Join(env.BaseDir, file2)
	require.NoError(t, os.WriteFile(path2, []byte("console.log('B');"), 0644))

	// Register both files as created
	require.NoError(t, env.AssetsHandler.NewFileEvent(file1, ".js", path1, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent(file2, ".js", path2, "create"))

	// Now trigger a write to force compilation/writing
	require.NoError(t, env.AssetsHandler.NewFileEvent(file1, ".js", path1, "write"))

	// Read generated main JS
	out, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err)
	outStr := string(out)

	// Count occurrences of use strict (both 'use strict' and "use strict")
	lower := strings.ToLower(outStr)
	count := strings.Count(lower, "use strict")
	require.Equal(t, 1, count, "There should be exactly one 'use strict' in the output")

	// Basic content checks: ensure both A and B outputs exist in the minified bundle
	require.Contains(t, outStr, "A")
	require.Contains(t, outStr, "B")
}
