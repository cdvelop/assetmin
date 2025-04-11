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
		// si el archivo no existe se considerara un error, la libreria debe ser capas de crear el directorio de trabajo web/public
		env := setupTestEnv("uc01_empty_directory", t)
		// 1. Create JS file and verify output
		jsFileName := "script1.js"
		jsFilePath := filepath.Join(env.BaseDir, jsFileName)
		jsContent := []byte("console.log('Hello from JS');")

		require.NoError(t, os.WriteFile(jsFilePath, jsContent, 0644))
		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "create"))

		// Verificar que el archivo main.js fue creado correctamente
		_, err := os.Stat(env.MainJsPath)
		require.NoError(t, err, "El archivo main.js no fue creado")
		require.FileExists(t, env.MainJsPath, "El archivo main.js no existe")

		// Verificar que el contenido fue escrito correctamente
		content, err := os.ReadFile(env.MainJsPath)
		require.NoError(t, err, "No se pudo leer el archivo main.js")
		require.Contains(t, string(content), "Hello from JS", "El contenido del archivo main.js no es el esperado")

		env.CleanDirectory()

	})

	t.Run("uc02_existing_directory", func(t *testing.T) {
		// en este caso los directorios ya existen y se crea un archivo JS que luego se edita
		// se espera que el contenido no esté duplicado en la salida web/public/main.js
		env := setupTestEnv("uc02_existing_directory", t)

		// Crear directorios primero
		env.CreatePublicDir()

		// 1. Crear archivo JS inicial
		jsFileName := "script1.js"
		jsFilePath := filepath.Join(env.BaseDir, jsFileName)
		initialContent := []byte("console.log('Initial content');")

		require.NoError(t, os.WriteFile(jsFilePath, initialContent, 0644))
		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "create"))

		// Verificar que el archivo main.js fue creado con el contenido inicial
		_, err := os.Stat(env.MainJsPath)
		require.NoError(t, err, "El archivo main.js no fue creado")
		initialMainContent, err := os.ReadFile(env.MainJsPath)
		require.NoError(t, err, "No se pudo leer el archivo main.js")
		require.Contains(t, string(initialMainContent), "Initial content", "El contenido inicial no es el esperado")

		// 2. Editar el mismo archivo JS
		updatedContent := []byte("console.log('Updated content');")
		require.NoError(t, os.WriteFile(jsFilePath, updatedContent, 0644))
		require.NoError(t, env.AssetsHandler.NewFileEvent(jsFileName, ".js", jsFilePath, "write"))

		// Verificar que el contenido se actualizó y no está duplicado
		updatedMainContent, err := os.ReadFile(env.MainJsPath)
		require.NoError(t, err, "No se pudo leer el archivo main.js actualizado")
		require.Contains(t, string(updatedMainContent), "Updated content", "El contenido actualizado no está presente")
		require.NotContains(t, string(updatedMainContent), "Initial content", "El contenido inicial no debería estar presente")

		env.CleanDirectory()
	})

	t.Run("uc03_concurrent_writes", func(t *testing.T) {
		// En este caso probamos el comportamiento de la librería cuando múltiples
		// archivos JS son escritos simultáneamente
		// Se espera que todos los contenidos se encuentren en web/public/main.js
		env := setupTestEnv("uc03_concurrent_writes", t)
		env.TestConcurrentFileProcessing(".js", 5)
	})

	t.Run("uc04_concurrent_writes_css", func(t *testing.T) {
		// En este caso probamos el comportamiento de la librería cuando múltiples
		// archivos CSS son escritos simultáneamente
		// Se espera que todos los contenidos se encuentren en web/public/main.css
		env := setupTestEnv("uc04_concurrent_writes_css", t)
		env.TestConcurrentFileProcessing(".css", 5)
	})
}
