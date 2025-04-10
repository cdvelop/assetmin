package assetmin

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

type AssetMin struct {
	*Config
	cssHandler *fileHandler
	jsHandler  *fileHandler
	min        *minify.M

	WriteOnDisk bool // Indica si se debe escribir en disco
}

type Config struct {
	ThemeFolder               func() string          // eg: web/theme
	WebFilesFolder            func() string          // eg: web/static, web/public, web/assets
	Print                     func(messages ...any)  // eg: fmt.Println
	JavascriptForInitializing func() (string, error) // javascript code to initialize the wasm or other handlers
}

type fileHandler struct {
	fileOutputName string                 // eg: main.js,style.css
	startCode      func() (string, error) // eg: "console.log('hello world')"
	themeFiles     []*assetFile           // files from theme folder
	moduleFiles    []*assetFile           // files from modules folder
	mediatype      string                 // eg: "text/html", "text/css", "image/svg+xml"
}

type assetFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}

func NewAssetMinify(config *Config) *AssetMin {
	c := &AssetMin{
		Config: config,
		cssHandler: &fileHandler{
			fileOutputName: "style.css",
			themeFiles:     []*assetFile{},
			moduleFiles:    []*assetFile{},
			mediatype:      "text/css",
		},
		jsHandler: &fileHandler{
			fileOutputName: "main.js",
			themeFiles:     []*assetFile{},
			moduleFiles:    []*assetFile{},
			mediatype:      "text/javascript",
		},
		min: minify.New(),
	}

	c.min.AddFunc("text/html", html.Minify)
	c.min.AddFunc("text/css", css.Minify)
	// c.min.AddFunc("text/javascript", js.Minify)
	c.min.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	c.min.AddFunc("image/svg+xml", svg.Minify)

	c.jsHandler.startCode = c.startCodeJS

	return c
}
