package assetmin

import (
	"bytes"
	"os"
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

	contentOpen   []*contentFile // eg: files from theme folder
	contentMiddle []*contentFile //eg: files from modules folder
	contentClose  []*contentFile // eg: files js from testin or end tags
}

// contentFile represents a file with its path and content
type contentFile struct {
	path    string // eg: modules/module1/file.js
	content []byte /// eg: "console.log('hello world')"
}

// WriteToDisk writes the content file to disk at the specified path
// It creates parent directories if they don't exist
func (f *contentFile) WriteToDisk() error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write content to the file
	return os.WriteFile(f.path, f.content, 0644)
}

// newAssetFile creates a new asset with the specified parameters
func newAssetFile(outputName, mediaType string, ac *Config, initCode func() (string, error)) *asset {
	handler := &asset{
		fileOutputName: outputName,
		outputPath:     filepath.Join(ac.OutputDir, outputName),
		mediatype:      mediaType,
		initCode:       initCode,
		contentOpen:    []*contentFile{},
		contentMiddle:  []*contentFile{},
		contentClose:   []*contentFile{},
	}

	return handler
}

// assetHandlerFiles ej &mainJsHandler, &mainStyleCssHandler
func (h *asset) UpdateContent(filePath, event string, f *contentFile) (err error) {
	// por defecto los archivos de destino son contenido comun eg: modulos, archivos sueltos
	filesToUpdate := &h.contentMiddle

	// Para archivos HTML, verificar si es un documento HTML completo
	// Si es así, debe ser ignorado ya que no es un módulo/fragmento
	if strings.HasSuffix(h.fileOutputName, ".html") && strings.HasSuffix(filePath, ".html") {
		// Verificar si el contenido es un documento HTML completo
		if isCompleteHtmlDocument(string(f.content)) {
			// Es un documento completo (template), ignorarlo
			// No procesamos templates como módulos
			return nil
		}
	}

	switch event {
	case "create", "write", "modify":

		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			// Exact path exists: replace content
			(*filesToUpdate)[idx] = f
		} else {
			// File with this path not found. This can happen in a rename flow where
			// a rename event is sent for the old file and a create event for the
			// new file arrives afterwards. Instead of blindly appending and
			// creating a duplicate, try to detect if this new file corresponds
			// to an existing memory entry (rename case) by comparing content.
			replaced := false
			for i, existing := range *filesToUpdate {
				if bytes.Equal(existing.content, f.content) {
					// Reuse existing entry: update its path and content
					(*filesToUpdate)[i].path = filePath
					(*filesToUpdate)[i].content = f.content
					replaced = true
					break
				}
			}
			if !replaced {
				// No match found: append as new file
				*filesToUpdate = append(*filesToUpdate, f)
			}
		}
	case "rename":
	case "remove", "delete":
		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			*filesToUpdate = slices.Delete((*filesToUpdate), idx, idx+1)
		}
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
