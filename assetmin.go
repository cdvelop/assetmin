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

const (
	jsMainFileName   = "main.js"    // eg: "main.js"
	cssMainFileName  = "style.css"  // eg: "style.css"
	svgMainFileName  = "sprite.svg" // eg: "sprite.svg"
	htmlMainFileName = "index.html" // eg: "index.html"
)

type AssetMin struct {
	mu sync.Mutex // Added mutex for synchronization
	*AssetConfig
	mainStyleCssHandler *asset
	mainJsHandler       *asset
	spriteSvgHandler    *asset
	indexHtmlHandler    *asset
	// indexHtmlHandler *asset
	min *minify.M

	WriteOnDisk bool // Indica si se debe escribir en disco
}

type AssetConfig struct {
	ThemeFolder             func() string          // eg: web/theme
	WebFilesFolder          func() string          // eg: web/static, web/public, web/assets
	Print                   func(messages ...any)  // eg: fmt.Println
	GetRuntimeInitializerJS func() (string, error) // javascript code to initialize the wasm or other handlers
}

func NewAssetMin(ac *AssetConfig) *AssetMin {
	c := &AssetMin{
		AssetConfig:         ac,
		mainStyleCssHandler: newAssetFile(cssMainFileName, "text/css", ac, nil),
		mainJsHandler:       newAssetFile(jsMainFileName, "text/javascript", ac, ac.GetRuntimeInitializerJS),
		spriteSvgHandler:    NewSvgHandler(ac),
		indexHtmlHandler:    NewHtmlHandler(ac),
		min:                 minify.New(),
		WriteOnDisk:         false, // Default to false
	}

	c.min.Add("text/html", &html.Minifier{
		KeepDocumentTags: true, // para mantener las etiquetas html,head,body. tambien mantiene las etiquetas de cierre
		KeepEndTags:      true, // preserve all end tags
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

	return c
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
	outputDir := c.WebFilesFolder()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.Print("dont create output dir", err)
	}
}
