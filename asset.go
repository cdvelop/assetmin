package assetmin

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
)

// represents a file handler for processing and minifying assets
type asset struct {
	fileOutputName string                 // eg: main.js,style.css,index.html,sprite.svg
	outputPath     string                 // full path to output file eg: web/public/main.js
	mediatype      string                 // eg: "text/html", "text/css", "image/svg+xml"
	initCode       func() (string, error) // eg js: "console.log('hello world')". eg: css: "body{color:red}" eg: html: "<html></html>". eg: svg: "<svg></svg>"
	themeFolder    string                 // eg: web/theme

	contentOpen   []*contentFile // eg: files from theme folder
	contentMiddle []*contentFile //eg: files from modules folder
	contentClose  []*contentFile // eg: files js from testin or end tags

	customFileProcessor func(event string, f *contentFile) error // Custom processor function

}

// newAssetFile creates a new asset with the specified parameters
func newAssetFile(outputName, mediaType string, ac *AssetConfig, initCode func() (string, error)) *asset {
	handler := &asset{
		fileOutputName:      outputName,
		outputPath:          filepath.Join(ac.WebFilesFolder(), outputName),
		mediatype:           mediaType,
		initCode:            initCode,
		themeFolder:         ac.ThemeFolder(),
		contentOpen:         []*contentFile{},
		contentMiddle:       []*contentFile{},
		contentClose:        []*contentFile{},
		customFileProcessor: nil, // Default to nil
	}

	return handler
}

// contentFile represents a file with its path and content
type contentFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}

// assetHandlerFiles ej &mainJsHandler, &mainStyleCssHandler
func (h *asset) UpdateContent(filePath, event string, f *contentFile) (err error) {

	// por defecto los archivos de destino son contenido comun eg: modulos, archivos sueltos
	filesToUpdate := &h.contentMiddle

	// verificar si es de tema asi actualizamos como archivos apertura
	if strings.Contains(filePath, h.themeFolder) {
		filesToUpdate = &h.contentOpen
	}

	switch event {
	case "create", "write":

		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			(*filesToUpdate)[idx] = f
		} else { // si no existe lo agregamos
			*filesToUpdate = append(*filesToUpdate, f)
		}

	case "rename": // cuando se renombra un archivo, se crea uno nuevo y se elimina el antiguo

	// se debe buscar el contenido anterior

	case "remove", "delete":
		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			*filesToUpdate = slices.Delete((*filesToUpdate), idx, idx+1)
		}
	}

	// If a custom processor is provided, use it for content-specific processing
	if h.customFileProcessor != nil && len(f.content) > 0 {
		err = h.customFileProcessor(event, f)
	}

	return
}

func findFileIndex(files []*contentFile, filePath string) int {
	for i, f := range files {
		if f.path == filePath {
			return i
		}
	}
	return -1
}

// WriteContent processes the asset content and writes it to the provided buffer
func (h *asset) WriteContent(buf *bytes.Buffer) {
	if h.initCode != nil {
		initCode, err := h.initCode()
		if err == nil {
			buf.WriteString(initCode)
		}
	}

	// Write open content first
	for _, f := range h.contentOpen {
		buf.Write(f.content)
		buf.WriteString("\n") // Add newline between files
	}

	// Then write middle content files
	for _, f := range h.contentMiddle {
		buf.Write(f.content)
		buf.WriteString("\n") // Add newline between files
	}

	// Then write close content files
	for _, f := range h.contentClose {
		buf.Write(f.content)
		buf.WriteString("\n") // Add newline between files
	}
}
