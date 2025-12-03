package assetmin

import (
	"os"
	"regexp"

	"sync"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

type AssetMin struct {
	mu sync.Mutex // Added mutex for synchronization
	*Config
	mainStyleCssHandler *asset
	mainJsHandler       *asset
	spriteSvgHandler    *asset
	faviconSvgHandler   *asset
	indexHtmlHandler    *asset
	// indexHtmlHandler *asset
	min *minify.M

	WriteOnDisk bool // Indica si se debe escribir en disco

	jsMainFileName     string // eg: script.js
	cssMainFileName    string // eg: style.css
	svgMainFileName    string // eg: icons.svg
	svgFaviconFileName string // eg: favicon.svg
	htmlMainFileName   string // eg: index.html
}

type Config struct {
	ThemeFolder             func() string          // eg: web/theme
	OutputDir               func() string          // eg: web/static, web/public, web/assets
	Logger                  func(message ...any)   // Renamed from io.Writer to a function type
	GetRuntimeInitializerJS func() (string, error) // javascript code to initialize the wasm or other handlers
	AppName                 string                 // Application name for templates (default: "MyApp")
}

func NewAssetMin(ac *Config) *AssetMin {
	c := &AssetMin{
		Config:      ac,
		min:         minify.New(),
		WriteOnDisk: true, // Default to true so library writes output by default; tests may disable it explicitly
		// initialize file name fields with previous constant values
		jsMainFileName:     "script.js",
		cssMainFileName:    "style.css",
		svgMainFileName:    "icons.svg",
		svgFaviconFileName: "favicon.svg",
		htmlMainFileName:   "index.html",
	}

	// Set default AppName if not provided
	if c.AppName == "" {
		c.AppName = "MyApp"
	}

	// handlers will be initialized after filename fields are set
	c.mainStyleCssHandler = newAssetFile(c.cssMainFileName, "text/css", ac, nil)
	c.mainJsHandler = newAssetFile(c.jsMainFileName, "text/javascript", ac, ac.GetRuntimeInitializerJS)
	c.spriteSvgHandler = NewSvgHandler(ac, c.svgMainFileName)
	c.faviconSvgHandler = NewFaviconSvgHandler(ac, c.svgFaviconFileName)
	c.indexHtmlHandler = NewHtmlHandler(ac, c.htmlMainFileName, c.cssMainFileName, c.jsMainFileName)
	c.min.Add("text/html", &html.Minifier{
		KeepDocumentTags: true, // para mantener las etiquetas html,head,body. tambien mantiene las etiquetas de cierre
		KeepEndTags:      true, // preserve all end tags
		KeepWhitespace:   true, // preserve whitespace to maintain structure
		KeepQuotes:       true, // preserve quotes in attribute values
	})

	c.min.AddFunc("text/css", css.Minify)
	c.min.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	c.min.AddFunc("image/svg+xml", svg.Minify)

	c.mainJsHandler.initCode = c.startCodeJS

	// No need to initialize output paths again as newAssetFile already does this
	// Ensure output directories exist
	c.EnsureOutputDirectoryExists()

	// Check if output files already exist
	c.NotifyIfOutputFilesExist()

	// If any output file already exists on disk, enable writing to disk so
	// the handler behaves as if it had been previously generating output.
	// This avoids tests (and real usage) needing to force WriteOnDisk externally.
	if fileExists(c.mainStyleCssHandler.outputPath) != "" ||
		fileExists(c.mainJsHandler.outputPath) != "" ||
		fileExists(c.spriteSvgHandler.outputPath) != "" ||
		fileExists(c.faviconSvgHandler.outputPath) != "" ||
		fileExists(c.indexHtmlHandler.outputPath) != "" {
		c.WriteOnDisk = true
	}

	return c
}

func (c *AssetMin) MainInputFileRelativePath() string {
	return c.indexHtmlHandler.themeFolder
}

func (c *AssetMin) SupportedExtensions() []string {
	return []string{".js", ".css", ".svg", ".html"}
}

// writeMessage writes a message to the configured Logger
func (c *AssetMin) writeMessage(messages ...any) {
	if c.Logger != nil {
		c.Logger(messages...)
	}
}

// NotifyIfOutputFilesExist checks if the output files for all assets already exist
func (c *AssetMin) NotifyIfOutputFilesExist() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Notify handlers if their notification callbacks are set
	if c.mainStyleCssHandler.notifyMeIfOutputFileExists != nil {
		c.mainStyleCssHandler.notifyMeIfOutputFileExists(fileExists(c.mainStyleCssHandler.outputPath))
	}

	if c.mainJsHandler.notifyMeIfOutputFileExists != nil {
		c.mainJsHandler.notifyMeIfOutputFileExists(fileExists(c.mainJsHandler.outputPath))
	}

	if c.spriteSvgHandler.notifyMeIfOutputFileExists != nil {
		c.spriteSvgHandler.notifyMeIfOutputFileExists(fileExists(c.spriteSvgHandler.outputPath))
	}

	if c.faviconSvgHandler.notifyMeIfOutputFileExists != nil {
		c.faviconSvgHandler.notifyMeIfOutputFileExists(fileExists(c.faviconSvgHandler.outputPath))
	}

	if c.indexHtmlHandler.notifyMeIfOutputFileExists != nil {
		c.indexHtmlHandler.notifyMeIfOutputFileExists(fileExists(c.indexHtmlHandler.outputPath))
	}

}

// Helper function to check if a file exists and return its content
func fileExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // Return empty string if file doesn't exist or can't be read
	}
	return string(data) // Return file content as string
}

// crea el directorio de salida si no existe
func (c *AssetMin) EnsureOutputDirectoryExists() {
	// Ensure main output directory exists
	outputDir := c.OutputDir()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.writeMessage("dont create output dir", err)
	}
}

// RefreshAsset triggers a rebuild of the asset handler for the given extension.
// This is useful when external dependencies (like TinyWasm's initializer JS) change.
// extension: e.g. ".js", ".css"
func (c *AssetMin) RefreshAsset(extension string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var fh *asset
	switch extension {
	case ".js":
		fh = c.mainJsHandler
	case ".css":
		fh = c.mainStyleCssHandler
	case ".svg":
		// Could be sprite or favicon, but usually we refresh JS/CSS
		// For now, let's assume we don't need to refresh SVG via this method
		// or we could add a specific method for it.
		// Given the requirement is for JS (wasm_exec), we focus on JS.
	}

	if fh != nil {
		// Force write to disk if not enabled, similar to UpdateAssetContent logic
		if !c.WriteOnDisk {
			c.WriteOnDisk = true
		}
		// Trigger processAndWrite
		// We pass a dummy event name as it's a manual refresh
		if err := c.processAndWrite(fh, "RefreshAsset "+extension); err != nil {
			c.writeMessage("Error refreshing asset "+extension, err)
		}
	}
}
