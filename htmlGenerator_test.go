package assetmin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDefaultIndexHtmlIfNotExist(t *testing.T) {
	t.Run("creates index.html when it doesn't exist", func(t *testing.T) {
		// Setup: Create temporary directories
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		// Ensure directories exist
		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		// Track log messages
		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		// Create AssetMin instance
		ac := &AssetConfig{
			ThemeFolder:    func() string { return themeDir },
			WebFilesFolder: func() string { return outputDir },
			Logger:         logger,
		}
		am := NewAssetMin(ac)

		// Execute: Call the method
		result := am.CreateDefaultIndexHtmlIfNotExist()

		// Verify: Check return value
		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		// Verify: Check file was created
		targetPath := filepath.Join(themeDir, "index.html")
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("Expected index.html to be created at %s", targetPath)
		}

		// Verify: Check file content
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)
		expectedPhrases := []string{
			"<!DOCTYPE html>",
			"<html",
			"<head>",
			"<body>",
			"</html>",
			`<link rel="stylesheet" href="style.css"`,
			`<script src="script.js"`,
		}

		for _, phrase := range expectedPhrases {
			if !strings.Contains(contentStr, phrase) {
				t.Errorf("Expected generated HTML to contain '%s', but it didn't", phrase)
			}
		}

		// Verify: Check log message
		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "Generated HTML file at") && strings.Contains(msg, targetPath) {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected log message about file generation")
		}
	})

	t.Run("skips creation when index.html already exists", func(t *testing.T) {
		// Setup: Create temporary directories
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		// Create existing index.html with custom content
		existingContent := "existing custom content"
		existingPath := filepath.Join(themeDir, "index.html")
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Track log messages
		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		// Create AssetMin instance
		ac := &AssetConfig{
			ThemeFolder:    func() string { return themeDir },
			WebFilesFolder: func() string { return outputDir },
			Logger:         logger,
		}
		am := NewAssetMin(ac)

		// Execute: Call the method
		result := am.CreateDefaultIndexHtmlIfNotExist()

		// Verify: Check return value
		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		// Verify: Check file content wasn't changed
		content, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatalf("Failed to read existing file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected existing file content to remain unchanged")
		}

		// Verify: Check log message about skipping
		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "already exists") && strings.Contains(msg, "skipping generation") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected log message about skipping file generation")
		}
	})

	t.Run("handles error when template cannot be read", func(t *testing.T) {
		// This test is mainly for coverage, as the embedded template should always be readable
		// in normal circumstances. We just verify the method doesn't panic.
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		ac := &AssetConfig{
			ThemeFolder:    func() string { return themeDir },
			WebFilesFolder: func() string { return outputDir },
			Logger:         func(messages ...any) {}, // Silent logger
		}
		am := NewAssetMin(ac)

		// This should not panic even if there were errors
		result := am.CreateDefaultIndexHtmlIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance even on error")
		}
	})
}

// Helper function to convert any type to string for logging
func anyToString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(v.(interface{ String() string }).String()), "\n"), "\n"))
}
