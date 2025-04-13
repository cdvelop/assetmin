package assetmin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHtmlModulesIntegration verifica que la funcionalidad de integración de módulos HTML
// funciona correctamente, utilizando contentOpen para la apertura del HTML,
// contentMiddle para los módulos HTML y contentClose para el cierre.
func TestHtmlModulesIntegration(t *testing.T) {
	t.Run("uc09_html_modules_integration", func(t *testing.T) {
		env := setupTestEnv("uc09_html_modules_integration", t)
		env.CreatePublicDir() // Aseguramos que el directorio público existe

		// Crear un directorio de prueba para módulos HTML
		testModulesDir := filepath.Join(env.BaseDir, "modules")
		require.NoError(t, os.MkdirAll(testModulesDir, 0755))

		// Crear archivos de módulos HTML individuales
		modulePaths := createTestHtmlModules(t, testModulesDir)

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
		assert.True(t, strings.Contains(htmlContent, "<!DOCTYPE html>"), "Debería contener la etiqueta DOCTYPE")

		// Verificar que contenga los módulos HTML
		assert.True(t, strings.Contains(htmlContent, "Módulo de prueba 1"), "Debería contener el módulo 1")
		assert.True(t, strings.Contains(htmlContent, "Módulo de prueba 2"), "Debería contener el módulo 2")

		// Verificar que contenga la etiqueta de cierre
		assert.True(t, strings.Contains(htmlContent, "</html>"), "Debería contener la etiqueta de cierre HTML")

		// Probar eliminar un módulo
		require.NoError(t, env.AssetsHandler.NewFileEvent("modulo1.html", ".html", modulePaths[0], "remove"))

		// Verificar que el HTML actualizado no contiene el módulo eliminado
		content, err = os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err, "Debería poder leer el archivo HTML actualizado")
		htmlContent = string(content)

		// El módulo eliminado no debería estar presente
		assert.False(t, strings.Contains(htmlContent, "Módulo de prueba 1"), "No debería contener el módulo eliminado")
		assert.True(t, strings.Contains(htmlContent, "Módulo de prueba 2"), "Debería seguir conteniendo el módulo 2")

		env.CleanDirectory()
	})
}

// TestHtmlStructure prueba que la estructura HTML sigue el formato esperado con secciones open, middle y close
func TestHtmlStructure(t *testing.T) {
	t.Run("uc10_html_structure", func(t *testing.T) {
		env := setupTestEnv("uc10_html_structure", t)
		env.CreatePublicDir()

		// Acceder directamente al manejador HTML
		htmlHandler := env.AssetsHandler.indexHtmlHandler

		// Verificar que contentOpen tiene el contenido adecuado
		require.GreaterOrEqual(t, len(htmlHandler.contentOpen), 1, "El manejador HTML debería tener contentOpen")

		// Verificar que contentOpen contiene la etiqueta de apertura HTML
		found := false
		for _, cf := range htmlHandler.contentOpen {
			content := string(cf.content)
			if strings.Contains(content, "<!DOCTYPE html>") {
				found = true
				break
			}
		}
		assert.True(t, found, "ContentOpen debería contener la etiqueta DOCTYPE HTML")

		// Crear un módulo de prueba
		moduleContent := `<div class="test-module">
    <h2>Módulo de Prueba</h2>
    <p>Este es un módulo HTML de prueba para verificar la estructura.</p>
</div>`
		moduleFile := &contentFile{
			path:    "test-module.html",
			content: []byte(moduleContent),
		}

		// Agregar el módulo al manejador sin escribir en disco
		require.NoError(t, htmlHandler.customFileProcessor("create", moduleFile))

		// En lugar de escribir en disco y leer el archivo, verificamos directamente
		// el contenido en memoria combinando contentOpen + contentMiddle + contentClose
		var htmlContent string

		// Añadir contentOpen
		for _, cf := range htmlHandler.contentOpen {
			htmlContent += string(cf.content)
		}

		// Añadir el contenido del módulo
		for _, cf := range htmlHandler.contentMiddle {
			htmlContent += string(cf.content)
		}

		// Añadir contentClose
		for _, cf := range htmlHandler.contentClose {
			htmlContent += string(cf.content)
		}

		// Verificar que tiene la estructura de apertura y cierre apropiada
		assert.Contains(t, htmlContent, "<!DOCTYPE html>", "Debería contener la etiqueta DOCTYPE")
		assert.Contains(t, htmlContent, "</html>", "Debería contener la etiqueta de cierre HTML")
		assert.Contains(t, htmlContent, "Módulo de Prueba", "Debería contener el módulo de prueba")

		// Verificar que los marcadores de inicio y fin de módulos están presentes
		assert.Contains(t, htmlContent, "<!-- Inicio de módulos HTML -->", "Debería contener el marcador de inicio de módulos")
		assert.Contains(t, htmlContent, "<!-- Fin de módulos HTML -->", "Debería contener el marcador de fin de módulos")

		env.CleanDirectory()
	})
}

// Función auxiliar para crear archivos de módulos HTML de prueba
func createTestHtmlModules(t *testing.T, dir string) []string {
	modules := []struct {
		name    string
		content string
	}{
		{
			name: "modulo1.html",
			content: `<div class="modulo-1">
    <h2>Módulo de prueba 1</h2>
    <p>Este es el contenido del módulo 1</p>
</div>`,
		},
		{
			name: "modulo2.html",
			content: `<div class="modulo-2">
    <h2>Módulo de prueba 2</h2>
    <p>Este es el contenido del módulo 2</p>
</div>`,
		},
	}

	var paths []string
	for _, module := range modules {
		path := filepath.Join(dir, module.name)
		require.NoError(t, os.WriteFile(path, []byte(module.content), 0644))
		paths = append(paths, path)
	}
	return paths
}
