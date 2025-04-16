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
}
