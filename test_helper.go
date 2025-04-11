package assetmin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/stretchr/testify/require"
)

// TestConcurrentFileProcessing es una función reutilizable para probar el procesamiento
// concurrente de archivos tanto para JS como para CSS.
func (env *TestEnvironment) TestConcurrentFileProcessing(fileExtension string, fileCount int) {
	// Determinar el tipo de archivo y la ruta de salida adecuada
	var outputPath string
	var fileType string

	switch fileExtension {
	case ".js":
		outputPath = env.MainJsPath
		fileType = "JS"
	case ".css":
		outputPath = env.MainCssPath
		fileType = "CSS"
	default:
		env.t.Fatalf("Extensión de archivo no soportada: %s", fileExtension)
	}

	// Crear archivos con contenido inicial
	fileNames := make([]string, fileCount)
	filePaths := make([]string, fileCount)
	fileContents := make([][]byte, fileCount)

	for i := range fileCount {
		fileNames[i] = fmt.Sprintf("file%d%s", i+1, fileExtension)
		filePaths[i] = filepath.Join(env.BaseDir, fileNames[i])
		fileContents[i] = []byte(fmt.Sprintf("console.log('Content from %s file %d');", fileType, i+1))
	}

	// Escribir archivos iniciales
	for i := range fileCount {
		require.NoError(env.t, os.WriteFile(filePaths[i], fileContents[i], 0644))
	}

	// Procesar archivos concurrentemente
	var wg sync.WaitGroup
	for i := range fileCount {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileNames[idx], fileExtension, filePaths[idx], "create"))
		}(i)
	}
	wg.Wait()

	// Verificar que el archivo de salida existe
	_, err := os.Stat(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("El archivo de salida no fue creado para %s", fileType))

	// Leer contenido del archivo de salida
	content, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo de salida para %s", fileType))

	// Verificar que el contenido de todos los archivos está presente
	contentStr := string(content)
	for i := range fileCount {
		expectedContent := fmt.Sprintf("Content from %s file %d", fileType, i+1)
		require.Contains(env.t, contentStr, expectedContent,
			fmt.Sprintf("El contenido del archivo %s %d no está presente", fileType, i+1))
	}

	// Actualizar todos los archivos con nuevo contenido
	updatedContents := make([][]byte, fileCount)
	for i := range fileCount {
		updatedContents[i] = []byte(fmt.Sprintf("console.log('Updated content from %s file %d');", fileType, i+1))
		require.NoError(env.t, os.WriteFile(filePaths[i], updatedContents[i], 0644))
	}

	// Procesar los archivos actualizados concurrentemente
	wg = sync.WaitGroup{}
	for i := range fileCount {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileNames[idx], fileExtension, filePaths[idx], "write"))
		}(i)
	}
	wg.Wait()

	// Leer el contenido actualizado del archivo de salida
	updatedContent, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo de salida actualizado para %s", fileType))
	updatedContentStr := string(updatedContent)

	// Verificar que el contenido actualizado de todos los archivos está presente
	for i := range fileCount {
		expectedUpdatedContent := fmt.Sprintf("Updated content from %s file %d", fileType, i+1)
		require.Contains(env.t, updatedContentStr, expectedUpdatedContent,
			fmt.Sprintf("El contenido actualizado del archivo %s %d no está presente", fileType, i+1))
	}

	// Verificar que el contenido original ya no está presente (no hay duplicación)
	for i := range fileCount {
		originalContent := fmt.Sprintf("Content from %s file %d", fileType, i+1)
		require.NotContains(env.t, updatedContentStr, originalContent,
			fmt.Sprintf("El contenido original del archivo %s %d no debería estar presente", fileType, i+1))
	}
}
