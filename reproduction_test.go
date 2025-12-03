package assetmin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInitialRegistrationScenario reproduces the real-world scenario where:
// 1. File watcher InitialRegistration sends "create" events for all existing files
// 2. Later, a "write" event happens to one of those files
// 3. The output should contain ALL previous files PLUS the updated content
// This test reproduces the bug where only the last modified file remains in output
func TestInitialRegistrationScenario(t *testing.T) {
	env := setupTestEnv("initial_registration_scenario", t)
	env.AssetsHandler.WriteOnDisk = true
	defer env.CleanDirectory()

	// Step 1: Simulate initial registration phase
	// Create multiple JS files on disk first (simulating existing project files)
	file1Path := filepath.Join(env.BaseDir, "modules", "module1", "script1.js")
	file2Path := filepath.Join(env.BaseDir, "modules", "module2", "script2.js")
	file3Path := filepath.Join(env.BaseDir, "web", "theme", "theme.js")

	// Create directory structure
	require.NoError(t, os.MkdirAll(filepath.Dir(file1Path), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(file2Path), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(file3Path), 0755))

	// Write initial content to files
	file1Content := "console.log('Module 1 functionality');"
	file2Content := "console.log('Module 2 functionality');"
	file3Content := "console.log('Theme initialization');"

	require.NoError(t, os.WriteFile(file1Path, []byte(file1Content), 0644))
	require.NoError(t, os.WriteFile(file2Path, []byte(file2Content), 0644))
	require.NoError(t, os.WriteFile(file3Path, []byte(file3Content), 0644))

	// Step 2: Simulate what InitialRegistration does - send "create" events for all files
	env.AssetsHandler.WriteOnDisk = false // Initially disabled, like in real scenario

	t.Log("=== Phase 1: Initial Registration (sending 'create' events) ===")
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("theme.js", ".js", file3Path, "create"))

	// At this point, main.js should NOT exist yet (WriteOnDisk is false)
	_, err := os.Stat(env.MainJsPath)
	require.True(t, os.IsNotExist(err), "main.js should not exist after create events with WriteOnDisk=false")

	// Step 3: Simulate a file modification that should trigger compilation
	t.Log("=== Phase 2: File modification (sending 'write' event) ===")
	modifiedFile2Content := "console.log('Module 2 functionality - UPDATED!');"
	require.NoError(t, os.WriteFile(file2Path, []byte(modifiedFile2Content), 0644))

	// This write event should enable WriteOnDisk and create the main.js with ALL content
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "write"))

	// Step 4: Verify that main.js contains ALL files content, not just the last modified one
	require.FileExists(t, env.MainJsPath, "main.js should exist after write event")

	content, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "Should be able to read main.js")
	contentStr := string(content)

	// Check that ALL three files are present in the output
	require.Contains(t, contentStr, "Module 1 functionality", "Content from module1/script1.js should be present")
	require.Contains(t, contentStr, "Module 2 functionality - UPDATED", "Updated content from module2/script2.js should be present")
	require.Contains(t, contentStr, "Theme initialization", "Content from theme/theme.js should be present")

	t.Log("Final main.js content length:", len(contentStr))
	t.Log("Content preview:", contentStr[:min(200, len(contentStr))])
}

// TestScenarioAfterWriteEvent tests what happens when another write event occurs
func TestScenarioAfterWriteEvent(t *testing.T) {
	env := setupTestEnv("after_write_scenario", t)
	env.AssetsHandler.WriteOnDisk = true
	defer env.CleanDirectory()

	// Create initial files
	file1Path := filepath.Join(env.BaseDir, "script1.js")
	file2Path := filepath.Join(env.BaseDir, "script2.js")

	require.NoError(t, os.WriteFile(file1Path, []byte("console.log('File 1');"), 0644))
	require.NoError(t, os.WriteFile(file2Path, []byte("console.log('File 2');"), 0644))

	// Phase 1: Create events (WriteOnDisk = false)
	env.AssetsHandler.WriteOnDisk = false
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "create"))

	// Phase 2: First write event
	require.NoError(t, os.WriteFile(file1Path, []byte("console.log('File 1 - UPDATED');"), 0644))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "write"))

	// Verify both files are in output
	content1, _ := os.ReadFile(env.MainJsPath)
	require.Contains(t, string(content1), "File 1 - UPDATED")
	require.Contains(t, string(content1), "File 2")

	// Phase 3: Second write event to different file
	require.NoError(t, os.WriteFile(file2Path, []byte("console.log('File 2 - UPDATED');"), 0644))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "write"))

	// Verify BOTH files are still in output (this is where the bug might occur)
	content2, _ := os.ReadFile(env.MainJsPath)
	require.Contains(t, string(content2), "File 1 - UPDATED", "First file content should still be present")
	require.Contains(t, string(content2), "File 2 - UPDATED", "Second file updated content should be present")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
