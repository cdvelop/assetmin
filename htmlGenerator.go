package assetmin

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var embeddedFS embed.FS

// createDefaultFileIfNotExist is a helper method to create default files from embedded templates
// It handles the common logic for checking existence, reading templates, and writing files.
func (a *AssetMin) createDefaultFileIfNotExist(fileName, templatePath, fileType string) *AssetMin {
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

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		if a.Logger != nil {
			a.Logger("Error creating directory:", err)
		}
		return a
	}

	if err := os.WriteFile(targetPath, raw, 0o644); err != nil {
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

// CreateDefaultIndexHtmlIfNotExist creates a default index.html file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultIndexHtmlIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.htmlMainFileName, "templates/index_basic.html", "HTML")
}

// CreateDefaultCssIfNotExist creates a default CSS file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultCssIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.cssMainFileName, "templates/style_basic.css", "CSS")
}

// CreateDefaultJsIfNotExist creates a default JavaScript file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultJsIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist(a.jsMainFileName, "templates/script_basic.js", "JS")
}

// CreateDefaultFaviconIfNotExist creates a default favicon.svg file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultFaviconIfNotExist() *AssetMin {
	return a.createDefaultFileIfNotExist("favicon.svg", "templates/favicon_basic.svg", "Favicon")
}
