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
	cssHandler  *fileHandler
	jsHandler   *fileHandler
	svgHandler  *svgHandler
	htmlHandler *htmlHandler
	// htmlHandler *fileHandler
	min *minify.M

	WriteOnDisk bool // Indica si se debe escribir en disco
}

type AssetConfig struct {
	ThemeFolder             func() string          // eg: web/theme
	WebFilesFolder          func() string          // eg: web/static, web/public, web/assets
	Print                   func(messages ...any)  // eg: fmt.Println
	GetRuntimeInitializerJS func() (string, error) // javascript code to initialize the wasm or other handlers
}

type fileHandler struct {
	fileOutputName string                 // eg: main.js,style.css,index.html,sprite.svg
	outputPath     string                 // full path to output file eg: web/public/main.js
	startCode      func() (string, error) // eg: "console.log('hello world')"
	themeFolder    string                 // eg: web/theme
	themeFiles     []*assetFile           // files from theme folder
	moduleFiles    []*assetFile           // files from modules folder
	mediatype      string                 // eg: "text/html", "text/css", "image/svg+xml"
}

type assetFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}

// NewFileHandler creates a new fileHandler with the specified parameters
func NewFileHandler(outputName, mediaType string, ac *AssetConfig) *fileHandler {
	handler := &fileHandler{
		fileOutputName: outputName,
		outputPath:     filepath.Join(ac.WebFilesFolder(), outputName),
		mediatype:      mediaType,
		startCode:      ac.GetRuntimeInitializerJS,
		themeFolder:    ac.ThemeFolder(),
		themeFiles:     []*assetFile{},
		moduleFiles:    []*assetFile{},
	}

	return handler
}

func NewAssetMin(ac *AssetConfig) *AssetMin {
	c := &AssetMin{
		AssetConfig: ac,
		cssHandler:  NewFileHandler(cssMainFileName, "text/css", ac),
		jsHandler:   NewFileHandler(jsMainFileName, "text/javascript", ac),
		svgHandler:  NewSvgHandler(ac),
		htmlHandler: NewHtmlHandler(ac),
		min:         minify.New(),
		WriteOnDisk: false, // Default to false
	}

	c.min.AddFunc("text/html", html.Minify)
	c.min.AddFunc("text/css", css.Minify)
	c.min.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	c.min.AddFunc("image/svg+xml", svg.Minify)

	c.jsHandler.startCode = c.startCodeJS

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
