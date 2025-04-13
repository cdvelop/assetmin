package assetmin

// represents a file handler for processing and minifying assets
type asset struct {
	fileOutputName string                 // eg: main.js,style.css,index.html,sprite.svg
	outputPath     string                 // full path to output file eg: web/public/main.js
	mediatype      string                 // eg: "text/html", "text/css", "image/svg+xml"
	initCode       func() (string, error) // eg js: "console.log('hello world')". eg: css: "body{color:red}" eg: html: "<html></html>". eg: svg: "<svg></svg>"
	themeFolder    string                 // eg: web/theme

	contentOpen   []*contentFile // eg: files from theme folder
	contentMiddle []*contentFile //eg: files from modules folder
	contentClose  []*contentFile // eg: files js from testin

	customFileProcessor func(event string, f *contentFile) error // Custom processor function

}

// contentFile represents a file with its path and content
type contentFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}
