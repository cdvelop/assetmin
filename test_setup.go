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
	MainCssPath   string
	AssetsHandler *AssetMin
	t             *testing.T
}

// CleanDirectory removes all content from the test directory but keeps the directory itself
func (env *TestEnvironment) CleanDirectory() {
	if _, err := os.Stat(env.BaseDir); err == nil {
		// env.t.Log("Cleaning test directory content...")
		// Remove content but keep the directory
		entries, err := os.ReadDir(env.BaseDir)
		if err == nil {
			for _, entry := range entries {
				entryPath := filepath.Join(env.BaseDir, entry.Name())
				os.RemoveAll(entryPath)
			}
		} else {
			env.t.Fatalf("Error reading directory: %v", err)
		}
	}
}

// setupTestEnv configures a minimal environment for testing AssetMin
// default write to disk is true, but can be set to false for testing purposes
func setupTestEnv(testCase string, t *testing.T) *TestEnvironment {
	// Create real directory instead of a temporary one
	baseDir := filepath.Join(".", "test", testCase)
	themeDir := filepath.Join(baseDir, "web", "theme")
	publicDir := filepath.Join(baseDir, "web", "public")
	modulesDir := filepath.Join(baseDir, "modules")

	// Create asset configuration with logging using t.Log
	config := &AssetConfig{
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
	assetsHandler := NewAssetMin(config)
	assetsHandler.WriteOnDisk = true

	// Create only the base directory if it doesn't exist
	err := os.MkdirAll(baseDir, 0755)
	require.NoError(t, err, "Failed to create base directory")

	return &TestEnvironment{
		BaseDir:       baseDir,
		ThemeDir:      themeDir,
		PublicDir:     publicDir,
		ModulesDir:    modulesDir,
		MainJsPath:    assetsHandler.jsHandler.outputPath,
		MainCssPath:   assetsHandler.cssHandler.outputPath,
		AssetsHandler: assetsHandler,
		t:             t,
	}
}

// CreateThemeDir creates the theme directory if it doesn't exist
func (env *TestEnvironment) CreateThemeDir() *TestEnvironment {
	err := os.MkdirAll(env.ThemeDir, 0755)
	require.NoError(env.t, err, "Failed to create theme directory")
	return env
}

// CreatePublicDir creates the public directory if it doesn't exist
func (env *TestEnvironment) CreatePublicDir() *TestEnvironment {
	err := os.MkdirAll(env.PublicDir, 0755)
	require.NoError(env.t, err, "Failed to create public directory")
	return env
}

// CreateModulesDir creates the modules directory if it doesn't exist
func (env *TestEnvironment) CreateModulesDir() *TestEnvironment {
	err := os.MkdirAll(env.ModulesDir, 0755)
	require.NoError(env.t, err, "Failed to create modules directory")
	return env
}
