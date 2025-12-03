package assetmin

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var embeddedFS embed.FS

// templateData holds the data to be used in template parsing
type templateData struct {
	AppName string
}

// createDefaultFileIfNotExist is a helper method to create default files from embedded templates
// It handles the common logic for checking existence, reading templates, and writing files.
// If appName is provided, it parses the template replacing {{.AppName}} placeholders.
// It writes the MINIFIED content directly to the output directory (OutputDir) if the source file (ThemeFolder) does not exist.
func (a *AssetMin) createDefaultFileIfNotExist(fileName, templatePath, mediaType, appName string) *AssetMin {
	// Check if file exists in SOURCE (ThemeFolder)
	sourcePath := filepath.Join(a.ThemeFolder(), fileName)
	if _, err := os.Stat(sourcePath); err == nil {
		if a.Logger != nil {
			a.Logger(mediaType, "source file already exists at", sourcePath, ", skipping default generation")
		}
		return a
	}

	// Read template
	raw, errRead := embeddedFS.ReadFile(templatePath)
	if errRead != nil {
		if a.Logger != nil {
			a.Logger("Error reading embedded template:", errRead)
		}
		return a
	}

	content := raw

	// If appName is provided, parse the template
	if appName != "" {
		data := templateData{AppName: appName}
		// Simple string replacement for {{.AppName}}
		contentStr := string(raw)
		contentStr = replacePlaceholder(contentStr, "{{.AppName}}", data.AppName)
		content = []byte(contentStr)
	}

	// Minify content
	var minifiedBuf bytes.Buffer
	if err := a.min.Minify(mediaType, &minifiedBuf, bytes.NewReader(content)); err != nil {
		if a.Logger != nil {
			a.Logger("Error minifying default template for", fileName, ":", err)
		}
		return a
	}

	// Determine output path (OutputDir)
	targetPath := filepath.Join(a.OutputDir(), fileName)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		if a.Logger != nil {
			a.Logger("Error creating directory:", err)
		}
		return a
	}

	if err := os.WriteFile(targetPath, minifiedBuf.Bytes(), 0o644); err != nil {
		if a.Logger != nil {
			a.Logger("Error writing", mediaType, "file:", err)
		}
		return a
	}

	if a.Logger != nil {
		a.Logger("Generated default minified", mediaType, "file at", targetPath)
	}

	return a
}

// replacePlaceholder is a simple helper to replace template placeholders
func replacePlaceholder(content, placeholder, value string) string {
	result := ""
	for i := 0; i < len(content); i++ {
		if i+len(placeholder) <= len(content) && content[i:i+len(placeholder)] == placeholder {
			result += value
			i += len(placeholder) - 1
		} else {
			result += string(content[i])
		}
	}
	return result
}

// CreateDefaultIndexHtmlIfNotExist creates a default index.html file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
// Uses AppName from Config to replace {{.AppName}} placeholder in the template.
func (a *AssetMin) CreateDefaultIndexHtmlIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.htmlMainFileName, "templates/index_basic.html", "text/html", a.AppName)
}

// CreateDefaultCssIfNotExist creates a default CSS file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultCssIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.cssMainFileName, "templates/style_basic.css", "text/css", "")
}

// CreateDefaultJsIfNotExist creates a default JavaScript file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
// Uses AppName from Config to replace {{.AppName}} placeholder in the template.
func (a *AssetMin) CreateDefaultJsIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.jsMainFileName, "templates/script_basic.js", "text/javascript", a.AppName)
}

// CreateDefaultFaviconIfNotExist creates a default favicon.svg file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultFaviconIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist("favicon.svg", "templates/favicon_basic.svg", "image/svg+xml", "")
}
