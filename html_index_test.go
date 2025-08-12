package assetmin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHtmlIndexIntegration verifica que al modificar un módulo HTML en un subdirectorio (por ejemplo, theme/index.html),
// el contenido del archivo principal index.html en public se actualiza correctamente, sin duplicar el contenido del módulo.
func TestHtmlIndexIntegration(t *testing.T) {
	t.Run("uc11_html_index_update_without_duplication", func(t *testing.T) {
		env := setupTestEnv("uc11_html_index_update_without_duplication", t)

		// Crear el directorio público y el index.html inicial
		env.CreatePublicDir()
		initialIndex := `<!doctype html>
<html>
<head><title>Test Index</title></head>
<body>
<main>
<!-- MODULES_PLACEHOLDER -->
</main>
</body>
</html>`
		require.NoError(t, os.WriteFile(env.MainHtmlPath, []byte(initialIndex), 0644))

		// Crear el subdirectorio theme y el archivo theme/index.html
		themeDir := filepath.Join(env.PublicDir, "theme")
		require.NoError(t, os.MkdirAll(themeDir, 0755))
		themeIndexPath := filepath.Join(themeDir, "index.html")
		themeContent := `<div class=\"theme-index\">Theme Index Content</div>`
		require.NoError(t, os.WriteFile(themeIndexPath, []byte(themeContent), 0644))

		// Procesar el nuevo archivo de módulo (simula evento de creación o modificación)
		require.NoError(t, env.AssetsHandler.NewFileEvent("theme/index.html", ".html", themeIndexPath, "create"))

		// Leer el index.html principal actualizado
		content, err := os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err)
		htmlContent := string(content)

		// El contenido del módulo debe estar presente solo una vez
		assert.Equal(t, 1, strings.Count(htmlContent, "Theme Index Content"), "El contenido del módulo debe aparecer solo una vez en index.html")

		// Simular una modificación del archivo theme/index.html
		updatedThemeContent := `<div class=\"theme-index\">Theme Index Content UPDATED</div>`
		require.NoError(t, os.WriteFile(themeIndexPath, []byte(updatedThemeContent), 0644))
		require.NoError(t, env.AssetsHandler.NewFileEvent("theme/index.html", ".html", themeIndexPath, "modify"))

		// Leer el index.html principal actualizado nuevamente
		content, err = os.ReadFile(env.MainHtmlPath)
		require.NoError(t, err)
		htmlContent = string(content)

		// El contenido actualizado debe estar presente solo una vez
		assert.Equal(t, 1, strings.Count(htmlContent, "Theme Index Content UPDATED"), "El contenido actualizado debe aparecer solo una vez en index.html")
		assert.NotContains(t, htmlContent, "Theme Index Content</div>\n<div class=\"theme-index\">Theme Index Content UPDATED", "No debe duplicar el contenido del módulo")
	})
}
