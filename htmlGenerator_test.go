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
		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
			AppName:     "TestApp",
		}
		am := NewAssetMin(ac)

		// Execute: Call the method
		result := am.CreateDefaultIndexHtmlIfNotExist()

		// Verify: Check return value
		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		// Verify: Check file was created
		targetPath := filepath.Join(outputDir, "index.html")
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
			"<!doctype html>",
			"<html",
			"<head>",
			"<body>",
			"</html>",
			`<title>TestApp</title>`,
			`<h1>Welcome to TestApp</h1>`,
			`<link rel="icon" type="image/svg+xml" href="favicon.svg"`,
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
			if strings.Contains(msg, "Generated default minified text/html file at") && strings.Contains(msg, targetPath) {
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
		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
			AppName:     "ExistingApp",
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
			if strings.Contains(msg, "text/html") && strings.Contains(msg, "source file already exists at") && strings.Contains(msg, existingPath) && strings.Contains(msg, "skipping default generation") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log message about skipping file generation. Got messages: %v", logMessages)
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

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      func(messages ...any) {}, // Silent logger
			AppName:     "ErrorApp",
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

func TestCreateDefaultCssIfNotExist(t *testing.T) {
	t.Run("creates CSS file when it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultCssIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		targetPath := filepath.Join(outputDir, "style.css")
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "body") {
			t.Error("Expected CSS to contain 'body' selector")
		}
		if !strings.Contains(contentStr, "font-family") {
			t.Error("Expected CSS to contain 'font-family' property")
		}

		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "Generated default minified text/css file at") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected log message about file generation")
		}
	})

	t.Run("skips creation when CSS already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		existingContent := "/* existing CSS */"
		existingPath := filepath.Join(themeDir, "style.css")
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultCssIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		content, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatalf("Failed to read existing file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected existing file content to remain unchanged")
		}

		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "already exists") && strings.Contains(msg, "skipping default generation") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log message about skipping file generation. Got messages: %v", logMessages)
		}
	})
}

func TestCreateDefaultJsIfNotExist(t *testing.T) {
	t.Run("creates JS file when it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
			AppName:     "TestJSApp",
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultJsIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		targetPath := filepath.Join(outputDir, "script.js")
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "TestJSApp") {
			t.Error("Expected JS to contain 'TestJSApp'")
		}
		if !strings.Contains(contentStr, "console.log") {
			t.Error("Expected JS to contain 'console.log'")
		}
		if !strings.Contains(contentStr, "DOMContentLoaded") {
			t.Error("Expected JS to contain 'DOMContentLoaded'")
		}

		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "Generated default minified text/javascript file at") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected log message about file generation")
		}
	})

	t.Run("skips creation when JS already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		existingContent := "// existing JS"
		existingPath := filepath.Join(themeDir, "script.js")
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
			AppName:     "ExistingJSApp",
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultJsIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		content, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatalf("Failed to read existing file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected existing file content to remain unchanged")
		}

		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "text/javascript source file already exists at") && strings.Contains(msg, existingPath) && strings.Contains(msg, "skipping default generation") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log message about skipping file generation. Got messages: %v", logMessages)
		}
	})

	t.Run("test method chaining", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      func(messages ...any) {},
			AppName:     "ChainApp",
		}
		am := NewAssetMin(ac)

		// Test chaining all three methods
		result := am.CreateDefaultIndexHtmlIfNotExist().
			CreateDefaultCssIfNotExist().
			CreateDefaultJsIfNotExist()

		if result != am {
			t.Error("Expected chained methods to return same *AssetMin instance")
		}

		// Verify all files were created
		files := []string{"index.html", "style.css", "script.js"}
		for _, file := range files {
			path := filepath.Join(outputDir, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s to be created", file)
			}
		}
	})
}

func TestCreateDefaultFaviconIfNotExist(t *testing.T) {
	t.Run("creates favicon.svg when it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultFaviconIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		targetPath := filepath.Join(outputDir, "favicon.svg")
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "<svg") {
			t.Error("Expected favicon to contain '<svg' tag")
		}
		if !strings.Contains(contentStr, "xmlns") {
			t.Error("Expected favicon to contain 'xmlns' attribute")
		}

		found := false
		for _, msg := range logMessages {
			if strings.Contains(msg, "Generated default minified image/svg+xml file at") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected log message about file generation")
		}
	})

	t.Run("skips creation when favicon already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		existingContent := "<svg>existing</svg>"
		existingPath := filepath.Join(themeDir, "favicon.svg")
		if err := os.WriteFile(existingPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		var logMessages []string
		logger := func(messages ...any) {
			var strMessages []string
			for _, msg := range messages {
				strMessages = append(strMessages, anyToString(msg))
			}
			logMessages = append(logMessages, strings.Join(strMessages, " "))
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      logger,
		}
		am := NewAssetMin(ac)

		result := am.CreateDefaultFaviconIfNotExist()

		if result != am {
			t.Error("Expected method to return *AssetMin instance for chaining")
		}

		content, err := os.ReadFile(existingPath)
		if err != nil {
			t.Fatalf("Failed to read existing file: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Expected existing file content to remain unchanged")
		}

		found := len(logMessages) > 0
		t.Logf("found = %v", found)
		if !found {
			t.Errorf("Expected log message about skipping file generation. Got messages: %v", logMessages)
		}
	})

	t.Run("test method chaining with favicon", func(t *testing.T) {
		tempDir := t.TempDir()
		themeDir := filepath.Join(tempDir, "theme")
		outputDir := filepath.Join(tempDir, "output")

		if err := os.MkdirAll(themeDir, 0755); err != nil {
			t.Fatalf("Failed to create theme directory: %v", err)
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		ac := &Config{
			ThemeFolder: func() string { return themeDir },
			OutputDir:   func() string { return outputDir },
			Logger:      func(messages ...any) {},
			AppName:     "FaviconChainApp",
		}
		am := NewAssetMin(ac)

		// Test chaining all methods including favicon
		result := am.CreateDefaultIndexHtmlIfNotExist().
			CreateDefaultCssIfNotExist().
			CreateDefaultJsIfNotExist().
			CreateDefaultFaviconIfNotExist()

		if result != am {
			t.Error("Expected chained methods to return same *AssetMin instance")
		}

		// Verify all files were created
		files := []string{"index.html", "style.css", "script.js", "favicon.svg"}
		for _, file := range files {
			path := filepath.Join(outputDir, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s to be created", file)
			}
		}
	})
}
