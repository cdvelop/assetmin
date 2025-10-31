package assetmin

import (
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
func (a *AssetMin) createDefaultFileIfNotExist(fileName, templatePath, fileType, appName string) *AssetMin {
	targetPath := filepath.Join(a.ThemeFolder(), fileName)

	if _, err := os.Stat(targetPath); err == nil {
		if a.Logger != nil {
			a.Logger(fileType, "file already exists at", targetPath, ", skipping generation")
		}
		return a
	}

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

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		if a.Logger != nil {
			a.Logger("Error creating directory:", err)
		}
		return a
	}

	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		if a.Logger != nil {
			a.Logger("Error writing", fileType, "file:", err)
		}
		return a
	}

	if a.Logger != nil {
		a.Logger("Generated", fileType, "file at", targetPath)
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
// Uses AppName from AssetConfig to replace {{.AppName}} placeholder in the template.
func (a *AssetMin) CreateDefaultIndexHtmlIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.htmlMainFileName, "templates/index_basic.html", "HTML", a.AppName)
}

// CreateDefaultCssIfNotExist creates a default CSS file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultCssIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.cssMainFileName, "templates/style_basic.css", "CSS", "")
}

// CreateDefaultJsIfNotExist creates a default JavaScript file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
// Uses AppName from AssetConfig to replace {{.AppName}} placeholder in the template.
func (a *AssetMin) CreateDefaultJsIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.jsMainFileName, "templates/script_basic.js", "JS", a.AppName)
}

// CreateDefaultFaviconIfNotExist creates a default favicon.svg file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultFaviconIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist("favicon.svg", "templates/favicon_basic.svg", "Favicon", "")
}
