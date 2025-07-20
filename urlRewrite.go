package assetmin

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// rewriteAssetUrls modifica las rutas en los atributos href y src de los elementos HTML.
// Mantiene solo el nombre del archivo y añade el nuevo directorio base (newRoot) a cada ruta.
//
// Parámetros:
//   - html: El código HTML a procesar
//   - newRoot: El nuevo directorio base para las rutas de los assets
//
// Retorna:
//   - El código HTML con las rutas de los assets actualizadas
//
// Ejemplos:
//
//  1. HTML básico con etiquetas <link> y <img>:
//     html := `<link href="css/styles.css"><img src="images/logo.png">`
//     result := rewriteAssetUrls(html, "/static")
//     // Resultado: <link href="/static/styles.css"><img src="/static/logo.png">
//
//  2. HTML con rutas absolutas y relativas mixtas:
//     html := `<script src="/scripts/app.js"></script><img src="../images/header.jpg">`
//     result := rewriteAssetUrls(html, "https://cdn.example.com/assets")
//     // Resultado: <script src="https://cdn.example.com/assets/app.js"></script><img src="https://cdn.example.com/assets/header.jpg">
func rewriteAssetUrls(html string, newRoot string) string {
	// Compilar la expresión regular para encontrar tags HTML
	tagRe := regexp.MustCompile(`<[^>]+>`)

	// Procesar el HTML completo sustituyendo cada tag
	result := tagRe.ReplaceAllStringFunc(html, func(tag string) string {
		var rex *regexp.Regexp
		var linkType string
		var newPath string

		// Determinar si es un tag con href o src
		if strings.Contains(tag, "href=") {
			rex = regexp.MustCompile(`href="([^"]+)"`)
			linkType = "href"
		} else if strings.Contains(tag, "src=") {
			rex = regexp.MustCompile(`src="([^"]+)"`)
			linkType = "src"
		} else {
			return tag // Si no tiene href o src, devolver sin cambios
		}

		// Seleccionar el contenido de href o src
		match := rex.FindStringSubmatch(tag)
		if len(match) > 1 {
			oldHref := match[1]

			// Extraer el nombre del archivo
			file := filepath.Base(oldHref)

			// Nuevo path con separador /
			newPath = newRoot + "/" + file

			// Reemplazar con el nuevo path
			return rex.ReplaceAllString(tag, fmt.Sprintf(`%s="%s"`, linkType, newPath))
		}

		return tag // Si no hay match, devolver sin cambios
	})

	return result
}
