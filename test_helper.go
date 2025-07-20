package assetmin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/stretchr/testify/require"
)

// TestConcurrentFileProcessing is a reusable function to test concurrent file processing for both JS and CSS.
func (env *TestEnvironment) TestConcurrentFileProcessing(fileExtension string, fileCount int) {
	// Determine the file type and appropriate output path
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
		env.t.Fatalf("Unsupported file extension: %s", fileExtension)
	}

	// Create files with initial content
	fileNames := make([]string, fileCount)
	filePaths := make([]string, fileCount)
	fileContents := make([][]byte, fileCount)

	for i := range fileCount {
		fileNames[i] = fmt.Sprintf("file%d%s", i+1, fileExtension)
		filePaths[i] = filepath.Join(env.BaseDir, fileNames[i])

		// Generate appropriate content based on file type
		if fileExtension == ".js" {
			fileContents[i] = []byte(fmt.Sprintf("console.log('Content from %s file %d');", fileType, i+1))
		} else if fileExtension == ".css" {
			fileContents[i] = []byte(fmt.Sprintf(".test-class-%d { color: blue; content: \"Content from %s file %d\"; }", i+1, fileType, i+1))
		}
	}

	// Write initial files
	for i := range fileCount {
		require.NoError(env.t, os.WriteFile(filePaths[i], fileContents[i], 0644))
	}

	// Process files concurrently
	var wg sync.WaitGroup
	for i := range fileCount {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileNames[idx], fileExtension, filePaths[idx], "create"))
		}(i)
	}
	wg.Wait()

	// Verify the output file exists
	_, err := os.Stat(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("The output file was not created for %s", fileType))

	// Read the output file content
	content, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("Failed to read the output file for %s", fileType))

	// Verify that the content of all files is present
	contentStr := string(content)
	for i := range fileCount {
		expectedContent := fmt.Sprintf("Content from %s file %d", fileType, i+1)
		require.Contains(env.t, contentStr, expectedContent,
			fmt.Sprintf("The content of %s file %d is not present", fileType, i+1))
	}

	// Update all files with new content
	updatedContents := make([][]byte, fileCount)
	for i := range fileCount {
		// Generate updated content based on file type
		if fileExtension == ".js" {
			updatedContents[i] = []byte(fmt.Sprintf("console.log('Updated content from %s file %d');", fileType, i+1))
		} else if fileExtension == ".css" {
			updatedContents[i] = []byte(fmt.Sprintf(".test-class-%d { color: red; content: \"Updated content from %s file %d\"; }", i+1, fileType, i+1))
		}
		require.NoError(env.t, os.WriteFile(filePaths[i], updatedContents[i], 0644))
	}

	// Process the updated files concurrently
	wg = sync.WaitGroup{}
	for i := range fileCount {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileNames[idx], fileExtension, filePaths[idx], "write"))
		}(i)
	}
	wg.Wait()

	// Read the updated output file content
	updatedContent, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("Failed to read the updated output file for %s", fileType))
	updatedContentStr := string(updatedContent)

	// Verify that the updated content of all files is present
	for i := range fileCount {
		var expectedUpdatedContent string
		if fileExtension == ".js" {
			expectedUpdatedContent = fmt.Sprintf("Updated content from %s file %d", fileType, i+1)
		} else if fileExtension == ".css" {
			expectedUpdatedContent = fmt.Sprintf("content:\"Updated content from %s file %d\"", fileType, i+1)
		}
		require.Contains(env.t, updatedContentStr, expectedUpdatedContent,
			fmt.Sprintf("The updated content of %s file %d is not present", fileType, i+1))
	}

	// Verify that the original content is no longer present (no duplication)
	for i := range fileCount {
		var originalContent string
		if fileExtension == ".js" {
			originalContent = fmt.Sprintf("Content from %s file %d", fileType, i+1)
		} else if fileExtension == ".css" {
			originalContent = fmt.Sprintf("content:\"Content from %s file %d\"", fileType, i+1)
		}
		require.NotContains(env.t, updatedContentStr, originalContent,
			fmt.Sprintf("The original content of %s file %d should not be present", fileType, i+1))
	}
}

// TestFileCRUDOperations tests the complete CRUD cycle (create, write, remove) for a file
func (env *TestEnvironment) TestFileCRUDOperations(fileExtension string) {
	// Determine the file type and appropriate output path
	var outputPath string

	switch fileExtension {
	case ".js":
		outputPath = env.MainJsPath
	case ".css":
		outputPath = env.MainCssPath
	default:
		env.t.Fatalf("Unsupported file extension: %s", fileExtension)
	}

	// Create directories first
	env.CreatePublicDir()

	// 1. Create file with initial content
	fileName := fmt.Sprintf("script1%s", fileExtension)
	filePath := filepath.Join(env.BaseDir, fileName)
	var initialContent []byte

	if fileExtension == ".js" {
		initialContent = []byte("console.log('Initial content');")
	} else {
		initialContent = []byte(".test { color: blue; content: 'Initial content'; }")
	}

	require.NoError(env.t, os.WriteFile(filePath, initialContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "create"))

	// Verify that the output file was created with the initial content
	_, err := os.Stat(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("El archivo %s no fue creado", outputPath))
	initialOutputContent, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo %s", outputPath))
	require.Contains(env.t, string(initialOutputContent), "Initial content", "El contenido inicial no es el esperado")

	// 2. Update the file content
	var updatedContent []byte
	if fileExtension == ".js" {
		updatedContent = []byte("console.log('Updated content');")
	} else {
		updatedContent = []byte(".test { color: red; content: 'Updated content'; }")
	}
	require.NoError(env.t, os.WriteFile(filePath, updatedContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "write"))

	// Verify the content was updated and not duplicated
	updatedOutputContent, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo %s actualizado", outputPath))
	require.Contains(env.t, string(updatedOutputContent), "Updated content", "El contenido actualizado no está presente")
	require.NotContains(env.t, string(updatedOutputContent), "Initial content", "El contenido inicial no debería estar presente")

	// 3. Remove the file
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "remove"))

	// Verify the content was removed
	finalOutputContent, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo %s después de eliminar", outputPath))
	require.NotContains(env.t, string(finalOutputContent), "Updated content", "El contenido eliminado no debería estar presente")
}

// TestThemePriority tests that files in 'theme' folder appear before files in 'modules' folder
func (env *TestEnvironment) TestThemePriority(fileExtension string) {
	// Determine the file type and appropriate output path
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
		env.t.Fatalf("Unsupported file extension: %s", fileExtension)
	}

	env.CreateModulesDir()
	env.CreateThemeDir()

	// 1. Create file in modules directory first
	modulesFileName := fmt.Sprintf("modules-file%s", fileExtension)
	modulesFilePath := filepath.Join(env.ModulesDir, modulesFileName)

	var modulesContent []byte
	var themeContent []byte

	if fileExtension == ".js" {
		modulesContent = []byte("console.log('Content from modules');")
		themeContent = []byte("console.log('Content from theme');")
	} else {
		modulesContent = []byte(".modules { color: blue; content: 'Content from modules'; }")
		themeContent = []byte(".theme { color: red; content: 'Content from theme'; }")
	}

	require.NoError(env.t, os.WriteFile(modulesFilePath, modulesContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(modulesFileName, fileExtension, modulesFilePath, "create"))

	// 2. Create file in theme directory second
	themeFileName := fmt.Sprintf("theme-file%s", fileExtension)
	themeFilePath := filepath.Join(env.ThemeDir, themeFileName)

	require.NoError(env.t, os.WriteFile(themeFilePath, themeContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(themeFileName, fileExtension, themeFilePath, "create"))

	// Verify the output file exists
	_, err := os.Stat(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("The output file was not created for %s", fileType))

	// Read the output file content
	content, err := os.ReadFile(outputPath)
	require.NoError(env.t, err, fmt.Sprintf("Failed to read the output file for %s", fileType))
	contentStr := string(content)

	// Verify that both contents are present
	require.Contains(env.t, contentStr, "Content from modules", "El contenido de modules no está presente")
	require.Contains(env.t, contentStr, "Content from theme", "El contenido de theme no está presente")

	// Verify that theme content appears before modules content
	themeIndex := strings.Index(contentStr, "Content from theme")
	modulesIndex := strings.Index(contentStr, "Content from modules")
	require.Less(env.t, themeIndex, modulesIndex,
		fmt.Sprintf("The theme content should appear before the modules content in %s", fileType))
}

// TestJSInitCodePriority verifica que el código de inicialización JS aparece al principio del archivo generado
func (env *TestEnvironment) TestJSInitCodePriority() {
	// Crear directorios necesarios
	env.CreateModulesDir()
	env.CreateThemeDir()
	env.CreatePublicDir()

	// 1. Crear archivo en theme
	themeFileName := "theme-file.js"
	themeFilePath := filepath.Join(env.ThemeDir, themeFileName)
	themeContent := []byte("console.log('Theme content');")

	require.NoError(env.t, os.WriteFile(themeFilePath, themeContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(themeFileName, ".js", themeFilePath, "create"))

	// 2. Crear archivo en modules
	modulesFileName := "modules-file.js"
	modulesFilePath := filepath.Join(env.ModulesDir, modulesFileName)
	modulesContent := []byte("console.log('Modules content');")

	require.NoError(env.t, os.WriteFile(modulesFilePath, modulesContent, 0644))
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(modulesFileName, ".js", modulesFilePath, "create"))

	// Verificar que el archivo se haya generado
	_, err := os.Stat(env.MainJsPath)
	require.NoError(env.t, err, "El archivo main.js no fue creado")

	// Leer contenido del archivo generado
	content, err := os.ReadFile(env.MainJsPath)
	require.NoError(env.t, err, "No se pudo leer el archivo main.js")
	contentStr := string(content)

	// Verificar que todos los contenidos estén presentes
	require.Contains(env.t, contentStr, "use strict", "El contenido 'use strict' debería estar presente")
	require.Contains(env.t, contentStr, "WebAssembly.Memory", "El código de inicialización de WebAssembly debería estar presente")
	require.Contains(env.t, contentStr, "Theme content", "El contenido de theme debería estar presente")
	require.Contains(env.t, contentStr, "Modules content", "El contenido de modules debería estar presente")

	// Verificar que el código de inicialización aparece antes que el contenido de theme y modules
	initCodeIndex := strings.Index(contentStr, "use strict")
	wasmCodeIndex := strings.Index(contentStr, "WebAssembly")
	themeIndex := strings.Index(contentStr, "Theme content")
	modulesIndex := strings.Index(contentStr, "Modules content")

	// El código de inicialización debe aparecer antes que todo
	require.Less(env.t, initCodeIndex, themeIndex, "El código de inicialización 'use strict' debe aparecer antes que el contenido de theme")
	require.Less(env.t, wasmCodeIndex, themeIndex, "El código de inicialización WebAssembly debe aparecer antes que el contenido de theme")

	// Verificar la secuencia correcta de importancia: initCode -> wasmCode -> theme -> modules
	require.True(env.t, initCodeIndex < wasmCodeIndex && wasmCodeIndex < themeIndex && themeIndex < modulesIndex,
		"El orden de importancia debe ser: código de inicialización -> código WebAssembly -> contenido de theme -> contenido de módulos")
}
