package assetmin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Función auxiliar para crear archivos de módulos HTML de prueba
func createTestHtmlModules(t *testing.T, dir string) []string {
	moduleTemplate := `<div class="module-%d">
    <h2>Test Module %d</h2>
    <p>module %d content</p>
</div>`
	var paths []string
	for i := range 2 {
		moduleNumber := i + 1
		content := fmt.Sprintf(moduleTemplate, moduleNumber, moduleNumber, moduleNumber)
		path := filepath.Join(dir, fmt.Sprintf("module%d.html", moduleNumber))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		paths = append(paths, path)
	}
	return paths
}

// TestHtmlModulesIntegration verifica que la funcionalidad de integración de módulos HTML
// funciona correctamente, utilizando contentOpen para la apertura del HTML,
// contentMiddle para los módulos HTML y contentClose para el cierre.
func TestHtmlModulesIntegration(t *testing.T) {
	t.Run("uc09_html_modules_integration_without_index", func(t *testing.T) {
		// este test verifica que actualicen modulos html cunando no existe un documento html
		// principal. En este caso, el archivo index.html no existe y se espera que se genere uno por defecto

		env := setupTestEnv("uc09_html_modules_integration_without_index", t)

		// Crear un directorio de prueba para módulos HTML
		env.CreateModulesDir()

		// Crear archivos de módulos HTML individuales
		modulePaths := createTestHtmlModules(t, env.ModulesDir)

		// Procesar cada archivo de módulo
		for _, modulePath := range modulePaths {
			moduleName := filepath.Base(modulePath)
			require.NoError(t, env.AssetsHandler.NewFileEvent(moduleName, ".html", modulePath, "create"))
		}

		// Verificar que el archivo HTML principal fue creado
		require.FileExists(t, env.MainHtmlPath, "El archivo index.html debería haberse creado")

		// Leer el archivo generado
		content, err := os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML generado")

		// Verificar la estructura del contenido
		htmlContent := string(content)

		// Verificar que contenga la etiqueta de apertura HTML
		assert.True(t, strings.Contains(htmlContent, "<!doctype html>"), "Debería contener la etiqueta doctype")

		// Verificar que contenga los módulos HTML
		assert.True(t, strings.Contains(htmlContent, "Test Module 1"), "Debería contener el módulo 1")
		assert.True(t, strings.Contains(htmlContent, "Test Module 2"), "Debería contener el módulo 2")

		// Verificar que contenga la etiqueta de cierre
		assert.True(t, strings.Contains(htmlContent, "</html>"), "Debería contener la etiqueta de cierre HTML")

		// Probar eliminar un módulo 1
		require.NoError(t, env.AssetsHandler.NewFileEvent("module1.html", ".html", modulePaths[0], "remove"))

		// Verificar que el HTML actualizado no contiene el módulo eliminado
		content, err = os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML actualizado")
		htmlContent = string(content)

		// El módulo eliminado no debería estar presente
		assert.False(t, strings.Contains(htmlContent, "Test Module 1"), "No debería contener el módulo 1 eliminado")
		assert.True(t, strings.Contains(htmlContent, "Test Module 2"), "Debería seguir conteniendo el módulo 2")

		env.CleanDirectory()
	})
	t.Run("uc10_html_modules_integration_with_index", func(t *testing.T) {
		// este test verifica que cuando existe un index.html en el directorio de salida
		// y se agregan nuevos módulos HTML, estos se deben integrar respetando la estructura existente

		// Crear un index.html existente con una estructura definida
		existingHtml := `<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Existing Index</title>
    <link rel="stylesheet" href="style.css" type="text/css" />
</head>
<body>
    <!-- Contenido existente -->
    <header>
        <h1>My Existing Page</h1>
    </header>
    <main>
        <!-- MODULES_PLACEHOLDER -->
    </main>
    <footer>
        <p>Existing footer</p>
    </footer>
    <script src="main.js" type="text/javascript"></script>
</body>
</html>`

		// Crear un contentFile para el HTML existente
		indexPath := filepath.Join(".", "test", "uc10_html_modules_integration_with_index", "web", "public", "index.html")
		indexFile := &contentFile{
			path:    indexPath,
			content: []byte(existingHtml),
		}

		// Pasar el contentFile a setupTestEnv para que lo escriba antes de inicializar NewAssetMin
		env := setupTestEnv("uc10_html_modules_integration_with_index", t, indexFile)

		// Primero limpiamos el directorio para asegurarnos de partir de cero
		// env.CleanDirectory() - Eliminamos esta limpieza ya que ahora queremos mantener el index.html

		// Crear el directorio público donde estará el index.html
		env.CreatePublicDir()

		// Verificar que el index.html existe
		require.FileExists(t, env.MainHtmlPath, "El archivo index.html debe existir antes de iniciar la prueba")

		// Crear un directorio de prueba para módulos HTML
		env.CreateModulesDir()

		// Crear archivos de módulos HTML individuales
		modulePaths := createTestHtmlModules(t, env.ModulesDir)

		// Procesar solo el primer módulo
		modulePath := modulePaths[0]
		moduleName := filepath.Base(modulePath)
		require.NoError(t, env.AssetsHandler.NewFileEvent(moduleName, ".html", modulePath, "create"))

		// Verificar que el archivo HTML existe (no debería haberse eliminado)
		require.FileExists(t, env.MainHtmlPath, "El archivo index.html debe seguir existiendo")

		// Leer el archivo actualizado
		content, err := os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML actualizado")

		// Verificar la estructura del contenido
		htmlContent := string(content)

		// Verificar que se conservan los elementos originales
		assert.True(t, strings.Contains(htmlContent, "<title>Existing Index</title>"), "Debe conservar la estructura del head")
		assert.True(t, strings.Contains(htmlContent, "<h1>My Existing Page</h1>"), "Debe conservar el header existente")
		assert.True(t, strings.Contains(htmlContent, "<p>Existing footer</p>"), "Debe conservar el footer existente")

		// Verificar que contiene el módulo HTML añadido
		assert.True(t, strings.Contains(htmlContent, "Test Module 1"), "Debe contener el módulo añadido")

		// Ahora procesamos el segundo módulo
		modulePath = modulePaths[1]
		moduleName = filepath.Base(modulePath)
		require.NoError(t, env.AssetsHandler.NewFileEvent(moduleName, ".html", modulePath, "create"))

		// Leer el archivo actualizado de nuevo
		content, err = os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML re-actualizado")
		htmlContent = string(content)

		// Verificar que ambos módulos están presentes
		assert.True(t, strings.Contains(htmlContent, "Test Module 1"), "Debe contener el primer módulo")
		assert.True(t, strings.Contains(htmlContent, "Test Module 2"), "Debe contener el segundo módulo")

		// Verificar que se conserva la estructura original
		assert.True(t, strings.Contains(htmlContent, "<header>"), "Debe conservar las etiquetas estructurales")
		assert.True(t, strings.Contains(htmlContent, "</footer>"), "Debe conservar las etiquetas de cierre")

		// env.CleanDirectory()
	})
	t.Run("uc11_html_template_should_be_ignored", func(t *testing.T) {
		// Este test verifica que cuando existe un archivo template.html con estructura HTML completa
		// (que comienza con <!doctype html> y termina con </body></html>), este NO debe ser tratado
		// como un módulo HTML sino que debe ser ignorado para evitar duplicación de contenido

		env := setupTestEnv("uc11_html_template_should_be_ignored", t)

		// Crear un directorio de prueba para módulos HTML
		env.CreateModulesDir()

		// Crear un archivo template.html con estructura HTML completa
		// Este archivo NO debería incluirse como módulo en el index.html
		templateContent := `<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Template Page</title>
    <link rel="stylesheet" href="style.css" type="text/css" />
</head>
<body>
    <header>
        <h1>Template Header</h1>
    </header>
    <main>
        <p>This is a complete template file</p>
    </main>
    <footer>
        <p>Template Footer</p>
    </footer>
    <script src="main.js" type="text/javascript"></script>
</body>
</html>`

		templatePath := filepath.Join(env.ModulesDir, "template.html")
		require.NoError(t, os.WriteFile(templatePath, []byte(templateContent), 0644))

		// Crear un módulo HTML normal (fragmento sin estructura completa)
		moduleContent := `<div class="module-test">
    <h2>Test Module</h2>
    <p>This is a normal module fragment</p>
</div>`

		modulePath := filepath.Join(env.ModulesDir, "module1.html")
		require.NoError(t, os.WriteFile(modulePath, []byte(moduleContent), 0644))

		// Procesar el archivo template.html (que debería ser ignorado)
		require.NoError(t, env.AssetsHandler.NewFileEvent("template.html", ".html", templatePath, "create"))

		// Procesar el módulo normal
		require.NoError(t, env.AssetsHandler.NewFileEvent("module1.html", ".html", modulePath, "create"))

		// Verificar que el archivo HTML principal fue creado
		require.FileExists(t, env.MainHtmlPath, "El archivo index.html debería haberse creado")

		// Leer el archivo generado
		content, err := os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML generado")

		htmlContent := string(content)

		// Verificar que contiene el módulo normal
		assert.True(t, strings.Contains(htmlContent, "Test Module"), "Debería contener el módulo normal")
		assert.True(t, strings.Contains(htmlContent, "This is a normal module fragment"), "Debería contener el contenido del módulo normal")

		// VERIFICAR QUE EL TEMPLATE.HTML NO SE INCLUYE COMO MÓDULO
		// No debe haber duplicación de estructura HTML
		assert.False(t, strings.Contains(htmlContent, "Template Header"), "NO debería contener el contenido del template.html")
		assert.False(t, strings.Contains(htmlContent, "This is a complete template file"), "NO debería contener el contenido del template.html")
		assert.False(t, strings.Contains(htmlContent, "Template Footer"), "NO debería contener el footer del template.html")

		// Verificar que el HTML generado tiene la estructura correcta (solo una vez)
		doctypeCount := strings.Count(htmlContent, "<!doctype html>")
		assert.Equal(t, 1, doctypeCount, "Solo debería haber un <!doctype html>")

		htmlOpenCount := strings.Count(htmlContent, "<html>")
		assert.Equal(t, 1, htmlOpenCount, "Solo debería haber una etiqueta <html>")

		bodyCloseCount := strings.Count(htmlContent, "</body>")
		assert.Equal(t, 1, bodyCloseCount, "Solo debería haber una etiqueta de cierre </body>")

		htmlCloseCount := strings.Count(htmlContent, "</html>")
		assert.Equal(t, 1, htmlCloseCount, "Solo debería haber una etiqueta de cierre </html>")

		t.Logf("HTML Content:\n%s", htmlContent)

		env.CleanDirectory()
	})
}
