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
}
