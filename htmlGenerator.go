package assetmin

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var embeddedFS embed.FS

// CreateDefaultIndexHtmlIfNotExist creates a default index.html file from the embedded template
// It never overwrites an existing file and returns the AssetMin instance for method chaining.
func (a *AssetMin) CreateDefaultIndexHtmlIfNotExist() *AssetMin {
	// Build target path from Config
	targetPath := filepath.Join(a.ThemeFolder(), a.htmlMainFileName)

	// Never overwrite existing files
	if _, err := os.Stat(targetPath); err == nil {
		if a.Logger != nil {
			a.Logger("HTML file already exists at", targetPath, ", skipping generation")
		}
		return a
	}

	// Read embedded template
	raw, errRead := embeddedFS.ReadFile("templates/index_basic.html")
	if errRead != nil {
		if a.Logger != nil {
			a.Logger("Error reading embedded template:", errRead)
		}
		return a
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		if a.Logger != nil {
			a.Logger("Error creating directory:", err)
		}
		return a
	}

	// Write the file
	if err := os.WriteFile(targetPath, raw, 0o644); err != nil {
		if a.Logger != nil {
			a.Logger("Error writing HTML file:", err)
		}
		return a
	}

	if a.Logger != nil {
		a.Logger("Generated HTML file at", targetPath)
	}

	return a
}
