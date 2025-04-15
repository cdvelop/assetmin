package assetmin

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
	// Configurar el handler de notificación de archivo de salida
	af.notifyMeIfOutputFileExists = hh.notifyMeIfOutputFileExists

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

func (h *htmlHandler) notifyMeIfOutputFileExists(content string) {
	// si es true quiere decir que el archivo de salida existe por ende debemos cambiar el contenido
	// de apertura y cierre segun este archivo

}
