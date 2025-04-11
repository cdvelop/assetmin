package assetmin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

		// Crear directorios primero
		env.CreatePublicDir()

		// Crear 5 archivos JS diferentes para procesamiento concurrente
		jsFiles := make([]string, 5)
		jsPaths := make([]string, 5)
		jsContents := make([][]byte, 5)

		for i := range 5 {
			jsFiles[i] = fmt.Sprintf("script%d.js", i+1)
			jsPaths[i] = filepath.Join(env.BaseDir, jsFiles[i])
			jsContents[i] = []byte(fmt.Sprintf("console.log('Content from JS file %d');", i+1))
		}

		// Escribir archivos iniciales
		for i := range 5 {
			require.NoError(t, os.WriteFile(jsPaths[i], jsContents[i], 0644))
		}

		// Procesar archivos concurrentemente
		var wg sync.WaitGroup
		for i := range 5 {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				require.NoError(t, env.AssetsHandler.NewFileEvent(jsFiles[idx], ".js", jsPaths[idx], "create"))
			}(i)
		}
		wg.Wait()

		// Verificar que el archivo main.js existe
		_, err := os.Stat(env.MainJsPath)
		require.NoError(t, err, "El archivo main.js no fue creado")

		// Leer contenido del archivo main.js
		content, err := os.ReadFile(env.MainJsPath)
		require.NoError(t, err, "No se pudo leer el archivo main.js")

		// Verificar que el contenido de todos los archivos está presente
		contentStr := string(content)
		for i := range 5 {
			expectedContent := fmt.Sprintf("Content from JS file %d", i+1)
			require.Contains(t, contentStr, expectedContent,
				fmt.Sprintf("El contenido del archivo JS %d no está presente", i+1))
		}

		// Limpiar directorio al finalizar
		// env.CleanDirectory()
	})
}
