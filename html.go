package assetmin

import "slices"

type htmlHandler struct {
	*asset
}

func NewHtmlHandler(ac *AssetConfig) *asset {
	h := newAssetFile(htmlMainFileName, "text/html", ac, nil)

	hh := &htmlHandler{
		asset: h,
	}
	// Configurar el procesador personalizado para manejar los módulos HTML
	h.customFileProcessor = hh.customFileProcessor

	//  default marcador de inicio index HTML
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

	// default marcador de cierre index HTML
	h.contentClose = append(h.contentClose, &contentFile{
		path: "index-close.html",
		content: []byte(`<script src="main.js" type="text/javascript"></script>
</body>
</html>`),
	})

	return h
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
