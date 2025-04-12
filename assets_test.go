package assetmin

import (
	"fmt"
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

	t.Run("uc02_crud_operations", func(t *testing.T) {
		// En este caso probamos operaciones CRUD (Create, Read, Update, Delete) en archivos
		// Se espera que el contenido se actualice correctamente (sin duplicados) y
		// que el contenido sea eliminado cuando se elimina el archivo
		env := setupTestEnv("uc02_crud_operations", t)

		// Probar operaciones CRUD para archivos JS
		t.Run("js_file", func(t *testing.T) {
			env.TestFileCRUDOperations(".js")
		})

		// Probar operaciones CRUD para archivos CSS
		t.Run("css_file", func(t *testing.T) {
			env.TestFileCRUDOperations(".css")
		})

		env.CleanDirectory()
	})

	t.Run("uc03_concurrent_writes", func(t *testing.T) {
		// En este caso probamos el comportamiento de la librería cuando múltiples
		// archivos JS son escritos simultáneamente
		// Se espera que todos los contenidos se encuentren en web/public/main.js
		env := setupTestEnv("uc03_concurrent_writes", t)
		env.TestConcurrentFileProcessing(".js", 5)
		env.CleanDirectory()
	})

	t.Run("uc04_concurrent_writes_css", func(t *testing.T) {
		// En este caso probamos el comportamiento de la librería cuando múltiples
		// archivos CSS son escritos simultáneamente
		// Se espera que todos los contenidos se encuentren en web/public/main.css
		env := setupTestEnv("uc04_concurrent_writes_css", t)
		env.TestConcurrentFileProcessing(".css", 5)
		env.CleanDirectory()
	})

	t.Run("uc05_theme_priority", func(t *testing.T) {
		// En este caso probamos que el contenido de los archivos en la carpeta 'theme'
		// aparezcan antes que los archivos de la carpeta 'modulos' en el archivo de salida
		env := setupTestEnv("uc05_theme_priority", t)

		// Probar prioridad de theme para archivos JS
		t.Run("js_theme_priority", func(t *testing.T) {
			env.TestThemePriority(".js")
		})

		// Probar prioridad de theme para archivos CSS
		t.Run("css_theme_priority", func(t *testing.T) {
			env.TestThemePriority(".css")
		})

		env.CleanDirectory()
	})

	t.Run("uc06_event_based_compilation", func(t *testing.T) {
		// Este caso prueba el comportamiento cuando el archivo main.js/css ya existe:
		// - Si se recibe un evento 'create', no se debe actualizar el contenido del archivo main
		// - Solo se actualiza el contenido cuando se reciben eventos 'write' o 'delete'
		// Este comportamiento evita compilaciones innecesarias cuando ya existe una versión compilada
		env := setupTestEnv("uc06_event_based_compilation", t)

		// Configuramos WriteOnDisk=false para simular que el archivo main ya existe
		// y que no se debe escribir al disco hasta recibir un evento write/delete

		// Probar comportamiento para archivos JS
		t.Run("js_event_behavior", func(t *testing.T) {
			env.AssetsHandler.WriteOnDisk = false
			env.TestEventBasedCompilation(".js")
		})

		// Probar comportamiento para archivos CSS
		t.Run("css_event_behavior", func(t *testing.T) {
			env.AssetsHandler.WriteOnDisk = false
			env.TestEventBasedCompilation(".css")
		})

		env.CleanDirectory()
	})
}

// TestEventBasedCompilation prueba el comportamiento de compilación basado en eventos
// cuando el archivo main ya existe (WriteOnDisk=false):
// - Los eventos 'create' no deben actualizar el archivo main
// - Solo los eventos 'write' o 'delete' deben actualizar el archivo main
func (env *TestEnvironment) TestEventBasedCompilation(fileExtension string) {
	// Determinar los valores según la extensión del archivo
	var fileName, fileContent, expectedContent string
	var mainPath string

	if fileExtension == ".js" {
		fileName = "script1.js"
		fileContent = "console.log('Initial JS content');"
		expectedContent = "Initial JS content"
		mainPath = env.MainJsPath
	} else if fileExtension == ".css" {
		fileName = "style1.css"
		fileContent = "body { color: red; }"
		expectedContent = "body { color: red; }"
		mainPath = env.MainCssPath
	}

	// Crear el archivo de prueba
	filePath := filepath.Join(env.BaseDir, fileName)
	require.NoError(env.t, os.WriteFile(filePath, []byte(fileContent), 0644))

	// Evento CREATE: No debería actualizar el archivo main porque WriteOnDisk=false
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "create"))

	// Verificar que el archivo main NO fue creado todavía
	_, err := os.Stat(mainPath)
	require.Error(env.t, err, fmt.Sprintf("El archivo main%s no debería existir después de un evento create", fileExtension))

	// Evento WRITE: Ahora SÍ debería actualizar el archivo main
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "write"))

	// Verificar que ahora el archivo main existe y contiene el contenido esperado
	require.FileExists(env.t, mainPath, fmt.Sprintf("El archivo main%s debería existir después de un evento write", fileExtension))
	content, err := os.ReadFile(mainPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo main%s", fileExtension))
	require.Contains(env.t, string(content), expectedContent, fmt.Sprintf("El contenido del archivo main%s no es el esperado", fileExtension))

	// Evento DELETE: Debería actualizar el archivo main eliminando el contenido
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "delete"))

	// El archivo main debería seguir existiendo pero sin el contenido del archivo eliminado
	content, err = os.ReadFile(mainPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo main%s", fileExtension))
	require.NotContains(env.t, string(content), expectedContent, fmt.Sprintf("El contenido eliminado no debería estar en main%s", fileExtension))
}
