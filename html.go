package assetmin

import "strings"

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
	// Si hay contenido, significa que el archivo de salida existe
	if content != "" {
		openContent, closeContent := parseExistingHtmlContent(content)

		// Limpiamos los contenidos anteriores
		h.asset.contentOpen = h.asset.contentOpen[:0]
		h.asset.contentClose = h.asset.contentClose[:0]

		// Reemplazamos el contenido de apertura y cierre con el encontrado
		h.asset.contentOpen = append(h.asset.contentOpen, &contentFile{
			path:    "existing-index-open.html",
			content: []byte(openContent),
		})

		h.asset.contentClose = append(h.asset.contentClose, &contentFile{
			path:    "existing-index-close.html",
			content: []byte(closeContent),
		})
	}
}

// parseExistingHtmlContent analiza un archivo HTML existente para identificar
// las secciones de apertura y cierre, considerando dónde deben insertarse los módulos
func parseExistingHtmlContent(content string) (openContent, closeContent string) {
	// Buscar un marcador explícito de commentario
	if i := strings.Index(content, "<!-- MODULES_PLACEHOLDER -->"); i != -1 {
		return content[:i], content[i+len("<!-- MODULES_PLACEHOLDER -->"):]
	}

	// Buscar un marcador de plantilla Go
	if i := strings.Index(content, "{{.Modules}}"); i != -1 {
		return content[:i], content[i+len("{{.Modules}}"):]
	}

	lines := strings.Split(content, "\n")
	var splitIndex int

	// 1. Buscar dentro de un tag <main> si existe
	inMain := false
	for i, line := range lines {
		lineLower := strings.ToLower(strings.TrimSpace(line))

		if strings.Contains(lineLower, "<main") {
			inMain = true
			continue
		}

		if inMain && strings.Contains(lineLower, "</main>") {
			splitIndex = i
			break
		}
	}

	// 2. Si no se encontró dentro de <main>, buscar antes del primer <script>
	if splitIndex == 0 {
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), "<script") {
				splitIndex = i
				break
			}
		}
	}

	// 3. Si no hay <script>, buscar antes de </body>
	if splitIndex == 0 {
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), "</body>") {
				splitIndex = i
				break
			}
		}
	}

	// Si todavía no tenemos un punto, usar el final
	if splitIndex == 0 {
		splitIndex = len(lines)
	}

	return strings.Join(lines[:splitIndex], "\n"), strings.Join(lines[splitIndex:], "\n")
}
