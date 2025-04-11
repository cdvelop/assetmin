package assetmin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetScenario(t *testing.T) {

	t.Run("uc01_empty_directory", func(t *testing.T) {
		// en este caso se espera que la libreria pueda crear el archivo el el directorio web/public/main.js
		// si el archivo no existe se considerara un error, la libreria de be ser capas de crear el directorio de trabajo web/public
		env := setupScenario("uc01_empty_directory", t)
		// 1. Create JS file and verify output
		jsFileName := "script1.js"
		jsFilePath := filepath.Join(env.ModulesDir, jsFileName)
		jsContent := []byte("console.log('Hello from JS');")

		t.Logf("Creating JS file: %s", jsFilePath)
		require.NoError(t, os.WriteFile(jsFilePath, jsContent, 0644))

		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "create"))

		env.CleanDirectory()
	})
}
