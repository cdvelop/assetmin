package assetmin

import "slices"

type htmlHandler struct {
	*asset
}

// generateStylesheetLink returns HTML tag for linking a CSS stylesheet
func generateStylesheetLink() []byte {
	return []byte(`<link rel="stylesheet" href="` + cssMainFileName + `" type="text/css" />`)
}

// generateJavaScriptTag returns HTML script tag for a JavaScript file
func generateJavaScriptTag() []byte {
	return []byte(`<script src="` + jsMainFileName + `" type="text/javascript"></script>`)
}

func NewHtmlHandler(ac *AssetConfig) *asset {
	af := newAssetFile(htmlMainFileName, "text/html", ac, nil)

	hh := &htmlHandler{
		asset: af,
	}
	// Configurar el procesador personalizado para manejar los módulos HTML
	af.customFileProcessor = hh.customFileProcessor

	//  default marcador de inicio index HTML
	af.contentOpen = append(af.contentOpen, &contentFile{
		path: "index-open.html",
		content: []byte(`<!doctype html>
<html>
<head>
	<meta charset="utf-8">
	<title></title>
	` + string(generateStylesheetLink()) + `
</head>
<body>`),
	})

	// default marcador de cierre index HTML
	af.contentClose = append(af.contentClose, &contentFile{
		path: "index-close.html",
		content: []byte(string(generateJavaScriptTag()) + `
</body>
</html>`),
	})

	return af
}

func (h *htmlHandler) customFileProcessor(event string, f *contentFile) error {
	// Si el evento es "remove", buscar y eliminar el módulo del arreglo contentMiddle
	if event == "remove" {
		for i, existingFile := range h.contentMiddle {
			if existingFile.path == f.path {
				// Eliminar el archivo del arreglo
				h.contentMiddle = slices.Delete(h.contentMiddle, i, i+1)
				break
			}
		}
		return nil
	}

	// Para eventos "create" o "update"
	// Primero verificar si el archivo ya existe en contentMiddle
	exists := false
	for i, existingFile := range h.contentMiddle {
		if existingFile.path == f.path {
			// Actualizar el contenido del archivo existente
			h.contentMiddle[i].content = f.content
			exists = true
			break
		}
	}

	// Si el archivo no existe en contentMiddle, agregarlo
	if !exists {
		h.contentMiddle = append(h.contentMiddle, f)
	}

	return nil
}
