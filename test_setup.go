package assetmin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupSimpleAssetTest configura un entorno mínimo para probar AssetMin sin un watcher real
func setupSimpleAssetTest(t *testing.T, cleanDir bool) (
	themeDir string,
	publicDir string,
	assetsHandler *AssetMin,
	mainJsPath string,
	styleCssPath string,
) {
	// Crear directorio real en lugar de uno temporal
	baseDir := filepath.Join(".", "test")
	themeDir = filepath.Join(baseDir, "theme")
	publicDir = filepath.Join(baseDir, "public")

	// Limpiar contenido si se solicita
	if cleanDir {
		if _, err := os.Stat(baseDir); err == nil {
			t.Log("Cleaning test directory content...")
			// Eliminar contenido pero mantener el directorio
			entries, err := os.ReadDir(baseDir)
			if err == nil {
				for _, entry := range entries {
					entryPath := filepath.Join(baseDir, entry.Name())
					os.RemoveAll(entryPath)
				}
			}
		}
	}

	// Crear directorios
	require.NoError(t, os.MkdirAll(themeDir, 0755))
	require.NoError(t, os.MkdirAll(publicDir, 0755))

	// Configurar rutas de salida
	mainJsPath = filepath.Join(publicDir, "main.js")
	styleCssPath = filepath.Join(publicDir, "style.css")

	// Crear configuración de assets con logging usando t.Log
	config := &Config{
		ThemeFolder:    func() string { return themeDir },
		WebFilesFolder: func() string { return publicDir },
		Print: func(messages ...any) {
			t.Log(messages...)
		},
		JavascriptForInitializing: func() (string, error) {
			return "", nil
		},
	}

	// Crear manejador de assets con escritura en disco habilitada
	assetsHandler = NewAssetMinify(config)
	assetsHandler.writeOnDisk = true

	return
}
