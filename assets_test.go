package assetmin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetScenario(t *testing.T) {
	// Use simplified configuration with our new TestEnvironment struct
	env := setupTest(t)

	t.Run("Asset compilation scenarios", func(t *testing.T) {
		// 1. Create JS file and verify output
		jsFileName := "script1.js"
		jsFilePath := filepath.Join(env.ThemeDir, jsFileName)
		jsContent := []byte("console.log('Hello from JS');")

		t.Logf("Creating JS file: %s", jsFilePath)
		require.NoError(t, os.WriteFile(jsFilePath, jsContent, 0644))

		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "create"))

		// Verify JS file was processed
		require.FileExists(t, env.MainJsPath, "main.js should be created after adding JS file")
		content, err := os.ReadFile(env.MainJsPath)
		require.NoError(t, err)
		require.Contains(t, string(content), "console.log(\"Hello from JS\")", "main.js should contain minified content from script1.js")

		// 2. Create CSS file and verify output
		cssFileName := "style1.css"
		cssFilePath := filepath.Join(env.ThemeDir, cssFileName)
		cssContent := []byte("body { color: blue; }")

		t.Logf("Creating CSS file: %s", cssFilePath)
		require.NoError(t, os.WriteFile(cssFilePath, cssContent, 0644))

		// Direct event call
		require.NoError(t, env.AssetsHandler.NewFileEvent(cssFileName, ".css", cssFilePath, "create"))

		// Verify CSS file was processed
		require.FileExists(t, env.StyleCssPath, "style.css should be created after adding CSS file")
		content, err = os.ReadFile(env.StyleCssPath)
		require.NoError(t, err)
		require.Contains(t, string(content), "body{color:blue}", "style.css should contain minified content from style1.css")

		// 3. Update JS file and verify content is updated (not duplicated)
		updatedJsContent := []byte("console.log('Updated JS content');")
		t.Logf("Updating JS file: %s", jsFilePath)
		require.NoError(t, os.WriteFile(jsFilePath, updatedJsContent, 0644))

		// Direct call to write event
		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "write"))

		// Verify JS content was updated
		content, err = os.ReadFile(env.MainJsPath)
		require.NoError(t, err)
		jsContentStr := string(content)
		require.Contains(t, jsContentStr, "console.log(\"Updated JS content\")", "main.js must contain the updated content")
		require.NotContains(t, jsContentStr, "Hello from JS", "main.js must not contain the original content")

		// 4. Update CSS file and verify content is updated
		updatedCssContent := []byte("body { background-color: black; color: white; }")
		t.Logf("Updating CSS file: %s", cssFilePath)
		require.NoError(t, os.WriteFile(cssFilePath, updatedCssContent, 0644))

		// Direct event call
		require.NoError(t, env.AssetsHandler.NewFileEvent(cssFileName, ".css", cssFilePath, "write"))

		// Verify CSS content was updated
		content, err = os.ReadFile(env.StyleCssPath)
		require.NoError(t, err)
		cssContentStr := string(content)
		require.Contains(t, cssContentStr, "body{background-color:#000;color:#fff}", "style.css must contain the updated content with hex codes")
		require.NotContains(t, cssContentStr, "body{color:blue}", "style.css must not contain the original content")

		// 5. Remove JS file and verify it is removed from output
		t.Logf("Removing JS file")
		require.NoError(t, os.Remove(jsFilePath))

		// Direct call to remove event
		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "remove"))

		// Verify JS content was removed
		content, err = os.ReadFile(env.MainJsPath)
		require.NoError(t, err)
		jsContentStr = string(content)
		t.Logf("Content after removal: %s", jsContentStr)
		require.NotContains(t, jsContentStr, "Updated JS content", "main.js must not contain the removed content")

		// 6. Remove CSS file and verify it is removed from output
		t.Logf("Removing CSS file")
		require.NoError(t, os.Remove(cssFilePath))

		// Direct event call
		require.NoError(t, env.AssetsHandler.NewFileEvent(cssFileName, ".css", cssFilePath, "remove"))

		// Verify CSS content was removed
		content, err = os.ReadFile(env.StyleCssPath)
		require.NoError(t, err)
		cssContentStr = string(content)
		t.Logf("CSS content after removal: %s", cssContentStr)
		require.NotContains(t, cssContentStr, "background-color:#000", "style.css must not contain the removed content")
	})
}
