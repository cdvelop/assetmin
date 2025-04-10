package assetmin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnvironment holds all the paths and components needed for asset tests
type TestEnvironment struct {
	BaseDir       string
	ThemeDir      string
	PublicDir     string
	ModulesDir    string
	MainJsPath    string
	StyleCssPath  string
	AssetsHandler *AssetMin
	t             *testing.T
}

// CleanDirectory removes all content from the test directory but keeps the directory itself
func (env *TestEnvironment) CleanDirectory() {
	if _, err := os.Stat(env.BaseDir); err == nil {
		env.t.Log("Cleaning test directory content...")
		// Remove content but keep the directory
		entries, err := os.ReadDir(env.BaseDir)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(env.BaseDir, entry.Name())
				os.RemoveAll(entryPath)
			}
		}
	}
}

// setupTest configures a minimal environment for testing AssetMin
// defaul write to disk is true, but can be set to false for testing purposes
func setupTest(t *testing.T) *TestEnvironment {
	// Create real directory instead of a temporary one
	baseDir := filepath.Join(".", "test")
	themeDir := filepath.Join(baseDir, "web", "theme")
	publicDir := filepath.Join(baseDir, "web", "public")
	modulesDir := filepath.Join(baseDir, "modules")

	// Create directories
	require.NoError(t, os.MkdirAll(themeDir, 0755))
	require.NoError(t, os.MkdirAll(publicDir, 0755))
	require.NoError(t, os.MkdirAll(modulesDir, 0755))

	// Configure output paths
	mainJsPath := filepath.Join(publicDir, "main.js")
	styleCssPath := filepath.Join(publicDir, "style.css")

	// Create asset configuration with logging using t.Log
	config := &Config{
		ThemeFolder:    func() string { return themeDir },
		WebFilesFolder: func() string { return publicDir },
		Print: func(messages ...any) {
			t.Log(messages...)
		},
		JavascriptForInitializing: func() (string, error) {
			return "", nil
		},
	}

	// Create asset handler with disk writing enabled
	assetsHandler := NewAssetMinify(config)
	assetsHandler.WriteOnDisk = true

	return &TestEnvironment{
		BaseDir:       baseDir,
		ThemeDir:      themeDir,
		PublicDir:     publicDir,
		ModulesDir:    modulesDir,
		MainJsPath:    mainJsPath,
		StyleCssPath:  styleCssPath,
		AssetsHandler: assetsHandler,
		t:             t,
	}
}
