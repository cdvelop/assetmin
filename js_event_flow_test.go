package assetmin

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestJSEventFlow verifies that when a project is already compiled and the watcher
// sends "create" events for existing files the output file remains unchanged.
// Then, when a new empty file is created the output should still remain unchanged.
// Finally, after writing content to the new file and sending a "write" event,
// the output should contain the previous files plus the new one.
func TestJSEventFlow(t *testing.T) {

	// Setup test environment

	env := setupTestEnv("js_event_flow", t)
	//defer env.CleanDirectory()

	// Prepare three distinct JS files in different directories
	file1Path := filepath.Join(env.BaseDir, "modules", "module1", "script1.js")
	file2Path := filepath.Join(env.BaseDir, "extras", "module2", "script2.js")
	file3Path := filepath.Join(env.BaseDir, "web", "theme", "theme.js")

	require.NoError(t, os.MkdirAll(filepath.Dir(file1Path), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(file2Path), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(file3Path), 0755))

	file1Content := "console.log('Module One');"
	file2Content := "console.log('Module Two');"
	file3Content := "console.log('Theme Code');"

	require.NoError(t, os.WriteFile(file1Path, []byte(file1Content), 0644))
	require.NoError(t, os.WriteFile(file2Path, []byte(file2Content), 0644))
	require.NoError(t, os.WriteFile(file3Path, []byte(file3Content), 0644))

	// Simulate initial compilation: send write events so main.js is produced
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "write"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "write"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("theme.js", ".js", file3Path, "write"))

	// Ensure main.js exists and capture its content
	require.FileExists(t, env.MainJsPath, "main.js must exist after initial write events")
	initialMain, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "unable to read main.js after initial compilation")

	// 1) Send "create" events for the same three files (simulating watcher initial registration)
	t.Log("Phase 1: sending 'create' events for existing files — expect no change")
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script2.js", ".js", file2Path, "create"))
	require.NoError(t, env.AssetsHandler.NewFileEvent("theme.js", ".js", file3Path, "create"))

	afterCreates, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "unable to read main.js after create events")
	// main.js should remain exactly the same
	require.True(t, bytes.Equal(initialMain, afterCreates), "main.js changed after duplicate 'create' events")

	// 2) Create a new empty JS file and send a 'create' event — output should remain unchanged
	newFilePath := filepath.Join(env.BaseDir, "modules", "module3", "newfile.js")
	require.NoError(t, os.MkdirAll(filepath.Dir(newFilePath), 0755))
	require.NoError(t, os.WriteFile(newFilePath, []byte{}, 0644))
	require.NoError(t, env.AssetsHandler.NewFileEvent("newfile.js", ".js", newFilePath, "create"))

	afterEmptyCreate, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "unable to read main.js after creating empty file")
	require.True(t, bytes.Equal(initialMain, afterEmptyCreate), "main.js changed after creating an empty file with 'create' event")

	// 3) Write content to the new file and send a 'write' event — expect previous content + new file
	addedContent := "console.log('New Module added');"
	require.NoError(t, os.WriteFile(newFilePath, []byte(addedContent), 0644))
	require.NoError(t, env.AssetsHandler.NewFileEvent("newfile.js", ".js", newFilePath, "write"))

	finalMain, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "unable to read main.js after writing new file")
	finalStr := string(finalMain)

	require.Contains(t, finalStr, "Module One", "final main.js should contain content from module1/script1.js")
	require.Contains(t, finalStr, "Module Two", "final main.js should contain content from module2/script2.js")
	require.Contains(t, finalStr, "Theme Code", "final main.js should contain content from web/theme/theme.js")
	require.Contains(t, finalStr, "New Module added", "final main.js should contain the newly written file content")

	// 4) Test rename operation: rename one of the existing files
	t.Log("Phase 4: testing rename operation")
	renamedFilePath := filepath.Join(env.BaseDir, "modules", "module1", "script1-renamed.js")
	renamedContent := "console.log('Module One Renamed');"

	// First send rename event for original file (removes it)
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1.js", ".js", file1Path, "rename"))

	// Create new file and send create event (adds it)
	require.NoError(t, os.WriteFile(renamedFilePath, []byte(renamedContent), 0644))
	require.NoError(t, env.AssetsHandler.NewFileEvent("script1-renamed.js", ".js", renamedFilePath, "create"))

	afterRename, err := os.ReadFile(env.MainJsPath)
	require.NoError(t, err, "unable to read main.js after rename operation")
	afterRenameStr := string(afterRename)

	require.NotContains(t, afterRenameStr, "console.log('Module One');", "main.js should NOT contain original script1 content after rename")
	require.Contains(t, afterRenameStr, "Module One Renamed", "main.js should contain renamed file content")
	require.Contains(t, afterRenameStr, "Module Two", "main.js should still contain script2 content")
	require.Contains(t, afterRenameStr, "Theme Code", "main.js should still contain theme content")
	require.Contains(t, afterRenameStr, "New Module added", "main.js should still contain the new module content")
}
