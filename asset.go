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
	themeFolder    string                 // eg: web/theme

	contentOpen   []*contentFile // eg: files from theme folder
	contentMiddle []*contentFile //eg: files from modules folder
	contentClose  []*contentFile // eg: files js from testin or end tags

	notifyMeIfOutputFileExists func(content string) // optional callback to notify if content if != "" file exists
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
		fileOutputName:             outputName,
		outputPath:                 filepath.Join(ac.OutputDir(), outputName),
		mediatype:                  mediaType,
		initCode:                   initCode,
		themeFolder:                ac.ThemeFolder(),
		contentOpen:                []*contentFile{},
		contentMiddle:              []*contentFile{},
		contentClose:               []*contentFile{},
		notifyMeIfOutputFileExists: nil, // Default to nil notification
	}

	return handler
}

// assetHandlerFiles ej &mainJsHandler, &mainStyleCssHandler
func (h *asset) UpdateContent(filePath, event string, f *contentFile) (err error) {

	// Para archivos HTML, manejar theme/index.html de forma especial
	if strings.HasSuffix(h.fileOutputName, ".html") && strings.Contains(filePath, h.themeFolder) && strings.HasSuffix(filePath, "index.html") {
		// theme/index.html debe reemplazar completamente el contenido HTML
		switch event {
		case "create", "write", "modify":
			// Limpiar solo el contenido open y close, mantener modules
			h.contentOpen = h.contentOpen[:0]
			h.contentClose = h.contentClose[:0]

			// Analizar el contenido del theme/index.html para dividirlo en secciones
			openContent, closeContent := parseExistingHtmlContent(string(f.content))

			// Establecer el nuevo contenido
			h.contentOpen = append(h.contentOpen, &contentFile{
				path:    "theme-index-open.html",
				content: []byte(openContent),
			})

			h.contentClose = append(h.contentClose, &contentFile{
				path:    "theme-index-close.html",
				content: []byte(closeContent),
			})

		case "remove", "delete":
			// Si se elimina theme/index.html, volver al HTML por defecto
			h.contentOpen = h.contentOpen[:0]
			h.contentClose = h.contentClose[:0]

			// Restaurar contenido por defecto HTML
			h.contentOpen = append(h.contentOpen, &contentFile{
				path: "index-open.html",
				content: []byte(`<!doctype html>
<html>
<head>
	<meta charset="utf-8">
	<title></title>
	<link rel="stylesheet" href="style.css" type="text/css" />
</head>
<body>`),
			})

			h.contentClose = append(h.contentClose, &contentFile{
				path: "index-close.html",
				content: []byte(`<script src="main.js" type="text/javascript"></script>
</body>
</html>`),
			})
		}
		return nil
	}

	// Lógica original para módulos HTML regulares y otros archivos
	// por defecto los archivos de destino son contenido comun eg: modulos, archivos sueltos
	filesToUpdate := &h.contentMiddle

	// verificar si es de tema así actualizamos como archivos apertura (para CSS, JS, etc.)
	// pero NO para index.html de tema (ya manejado arriba)
	if strings.Contains(filePath, h.themeFolder) && !strings.HasSuffix(filePath, "index.html") {
		filesToUpdate = &h.contentOpen
	}

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

		// Debug: log what was updated (commented out)
		// if h.fileOutputName == "main.js" {
		//     fmt.Printf("DEBUG asset.UpdateContent: %s event=%s, total files=%d\n", filePath, event, len(*filesToUpdate))
		//     for i, cf := range *filesToUpdate {
		//         fmt.Printf("DEBUG   [%d] path=%s size=%d\n", i, cf.path, len(cf.content))
		//     }
		// }

	case "rename": // cuando se renombra un archivo, se crea uno nuevo y se elimina el antiguo
		// Previously we removed the old entry here. That causes the create event
		// for the new file to append a new entry, potentially duplicating or
		// losing ordering information. Instead, treat rename as a no-op and let
		// the subsequent create/write event reuse/update the existing entry.
		// No action required here.

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
