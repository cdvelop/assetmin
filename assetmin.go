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

type WorkMode int

const (
	MemoryMode WorkMode = iota // Serve from memory cache (default)
	DiskMode                     // Write to disk + serve from cache
)

type AssetMin struct {
	mu sync.Mutex // Added mutex for synchronization
	*Config
	mainStyleCssHandler *asset
	mainJsHandler       *asset
	spriteSvgHandler    *asset
	faviconSvgHandler   *asset
	indexHtmlHandler    *asset
	min                 *minify.M
	workMode            WorkMode // Current work mode
	jsMainFileName      string   // eg: script.js
	cssMainFileName     string   // eg: style.css
	svgMainFileName     string   // eg: icons.svg
	svgFaviconFileName  string // eg: favicon.svg
	htmlMainFileName    string // eg: index.html
}

type Config struct {
	OutputDir               string                 // eg: web/static, web/public, web/assets
	Logger                  func(message ...any)   // Renamed from io.Writer to a function type
	GetRuntimeInitializerJS func() (string, error) // javascript code to initialize the wasm or other handlers
	AppName                 string                 // Application name for templates (default: "MyApp")
	AssetsURLPrefix         string                 // New: for HTTP routes
}

func NewAssetMin(ac *Config) *AssetMin {
	c := &AssetMin{
		Config:             ac,
		min:                minify.New(),
		jsMainFileName:     "script.js",
		cssMainFileName:    "style.css",
		svgMainFileName:    "icons.svg",
		svgFaviconFileName: "favicon.svg",
		htmlMainFileName:   "index.html",
	}

	if c.AppName == "" {
		c.AppName = "MyApp"
	}

	c.mainStyleCssHandler = newAssetFile(c.cssMainFileName, "text/css", ac, nil)
	c.mainJsHandler = newAssetFile(c.jsMainFileName, "text/javascript", ac, ac.GetRuntimeInitializerJS)
	c.spriteSvgHandler = NewSvgHandler(ac, c.svgMainFileName)
	c.faviconSvgHandler = NewFaviconSvgHandler(ac, c.svgFaviconFileName)
	c.indexHtmlHandler = NewHtmlHandler(ac, c.htmlMainFileName, c.cssMainFileName, c.jsMainFileName)
	c.min.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepEndTags:      true,
		KeepWhitespace:   true,
		KeepQuotes:       true,
	})

	c.min.AddFunc("text/css", css.Minify)
	c.min.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	c.min.AddFunc("image/svg+xml", svg.Minify)

	c.mainJsHandler.initCode = c.startCodeJS

	return c
}

func (c *AssetMin) SupportedExtensions() []string {
	return []string{".js", ".css", ".svg", ".html"}
}

func (c *AssetMin) writeMessage(messages ...any) {
	if c.Logger != nil {
		c.Logger(messages...)
	}
}

func fileExists(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (c *AssetMin) EnsureOutputDirectoryExists() {
	outputDir := c.OutputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.writeMessage("dont create output dir", err)
	}
}

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
	}

	if fh != nil {
		if err := c.processAsset(fh); err != nil {
			c.writeMessage("Error refreshing asset "+extension, err)
		}
	}
}

// SetWorkMode sets the work mode for AssetMin.
func (c *AssetMin) SetWorkMode(mode WorkMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workMode = mode
}

// GetWorkMode returns the current work mode of AssetMin.
func (c *AssetMin) GetWorkMode() WorkMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.workMode
}
