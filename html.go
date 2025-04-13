package assetmin

type htmlHandler struct {
	*asset
}

func NewHtmlHandler(ac *AssetConfig) *asset {
	h := NewFileHandler(htmlMainFileName, "text/html", ac, nil)

	hh := &htmlHandler{
		asset: h,
	}
	// Configurar el procesador personalizado para manejar los módulos HTML
	h.customFileProcessor = hh.processModuleFile

	// Agregar el marcador de inicio de los módulos HTML
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

	// Agregar el marcador de cierre de los módulos HTML
	h.contentClose = append(h.contentClose, &contentFile{
		path: "index-close.html",
		content: []byte(`<script src="main.js" type="text/javascript"></script>
</body>
</html>`),
	})

	return h
}

func (h *htmlHandler) processModuleFile(event string, f *contentFile) error {

	return nil
}
