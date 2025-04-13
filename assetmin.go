package assetmin

import (
	"os"
	"path/filepath"
	"regexp"

	"sync"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

const (
	jsMainFileName   = "main.js"    // eg: "main.js"
	cssMainFileName  = "style.css"  // eg: "style.css"
	svgMainFileName  = "sprite.svg" // eg: "sprite.svg"
	htmlMainFileName = "index.html" // eg: "index.html"
)

type AssetMin struct {
	mu sync.Mutex // Added mutex for synchronization
	*AssetConfig
	mainStyleCssHandler *fileHandler
	mainJsHandler       *fileHandler
	spriteSvgHandler    *spriteSvgHandler
	indexHtmlHandler    *indexHtmlHandler
	// indexHtmlHandler *fileHandler
	min *minify.M

	WriteOnDisk bool // Indica si se debe escribir en disco
}

type AssetConfig struct {
	ThemeFolder             func() string          // eg: web/theme
	WebFilesFolder          func() string          // eg: web/static, web/public, web/assets
	Print                   func(messages ...any)  // eg: fmt.Println
	GetRuntimeInitializerJS func() (string, error) // javascript code to initialize the wasm or other handlers
}

type customContentProcessor func(content []byte, event string) ([]byte, error)

// represents a file handler for processing and minifying assets
type fileHandler struct {
	fileOutputName string                 // eg: main.js,style.css,index.html,sprite.svg
	outputPath     string                 // full path to output file eg: web/public/main.js
	mediatype      string                 // eg: "text/html", "text/css", "image/svg+xml"
	initCode       func() (string, error) // eg js: "console.log('hello world')". eg: css: "body{color:red}" eg: html: "<html></html>". eg: svg: "<svg></svg>"
	themeFolder    string                 // eg: web/theme

	contentOpen   []*contentFile // eg: files from theme folder
	contentMiddle []*contentFile //eg: files from modules folder
	contentClose  []*contentFile // eg: files js from testin

	processor customContentProcessor // Custom processor function

}

// contentFile represents a file with its path and content
type contentFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}

// NewFileHandler creates a new fileHandler with the specified parameters
func NewFileHandler(outputName, mediaType string, ac *AssetConfig, initCode func() (string, error)) *fileHandler {
	handler := &fileHandler{
		fileOutputName: outputName,
		outputPath:     filepath.Join(ac.WebFilesFolder(), outputName),
		mediatype:      mediaType,
		initCode:       initCode,
		themeFolder:    ac.ThemeFolder(),
		contentOpen:    []*contentFile{},
		contentMiddle:  []*contentFile{},
		contentClose:   []*contentFile{},
	}

	return handler
}

func NewAssetMin(ac *AssetConfig) *AssetMin {
	c := &AssetMin{
		AssetConfig:         ac,
		mainStyleCssHandler: NewFileHandler(cssMainFileName, "text/css", ac, nil),
		mainJsHandler:       NewFileHandler(jsMainFileName, "text/javascript", ac, ac.GetRuntimeInitializerJS),
		spriteSvgHandler:    NewSvgHandler(ac),
		indexHtmlHandler:    NewHtmlHandler(ac),
		min:                 minify.New(),
		WriteOnDisk:         false, // Default to false
	}

	c.min.AddFunc("text/html", html.Minify)
	c.min.AddFunc("text/css", css.Minify)
	c.min.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	c.min.AddFunc("image/svg+xml", svg.Minify)

	c.mainJsHandler.initCode = c.startCodeJS

	// No need to initialize output paths again as NewFileHandler already does this
	// Ensure output directories exist
	c.EnsureOutputDirectoryExists()

	return c
}

// crea el directorio de salida si no existe
func (c *AssetMin) EnsureOutputDirectoryExists() {
	// Ensure main output directory exists
	outputDir := c.WebFilesFolder()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.Print("dont create output dir", err)
	}

}
