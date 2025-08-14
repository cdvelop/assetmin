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

		// Probar que el código de inicialización JS aparezca al principio
		t.Run("js_init_code_priority", func(t *testing.T) {
			env.TestJSInitCodePriority()
		})

		env.CleanDirectory()
	})

	t.Run("uc06_event_based_compilation", func(t *testing.T) {
		// Este caso prueba el comportamiento cuando el archivo main.js/css ya existe:
		// - Si se recibe un evento 'create', no se debe actualizar el contenido del archivo main
		// - Solo se actualiza el contenido cuando se reciben eventos 'write' o 'delete'
		// Este comportamiento evita compilaciones innecesarias cuando ya existe una versión compilada
		env := setupTestEnv("uc06_event_based_compilation", t)

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
// cuando WriteOnDisk=false (modo InitialRegistration):
// - Los eventos 'create' NO deben escribir a disco, solo almacenar en memoria
// - Solo los eventos 'write' o 'delete' deben habilitar escritura a disco
// - Esto permite que InitialRegistration cargue todo en memoria antes de compilar
func (env *TestEnvironment) TestEventBasedCompilation(fileExtension string) {
	// Determinar los valores según la extensión del archivo
	var fileName, fileContent, expectedContent string
	var mainPath string
	var initialContent string

	if fileExtension == ".js" {
		fileName = "script1.js"
		fileContent = "console.log('Initial JS content');"
		expectedContent = "Initial JS content"
		mainPath = env.MainJsPath
		initialContent = "var existingContent = 'compiled-content';"
	} else if fileExtension == ".css" {
		fileName = "style1.css"
		fileContent = "body { color: red; }"
		// Usar un término de búsqueda sin depender de los espacios
		expectedContent = "body{color:red}"
		mainPath = env.MainCssPath
		initialContent = ".existing-content { color: blue; }"
	}

	// Primera parte: Probar comportamiento cuando el archivo main ya existe

	// Crear un archivo main inicial con contenido existente
	require.NoError(env.t, os.MkdirAll(filepath.Dir(mainPath), 0755))
	require.NoError(env.t, os.WriteFile(mainPath, []byte(initialContent), 0644))

	// Crear el archivo de prueba
	filePath := filepath.Join(env.BaseDir, fileName)
	require.NoError(env.t, os.WriteFile(filePath, []byte(fileContent), 0644))

	// Evento CREATE: No debería actualizar el archivo main porque ya existe una compilacion previa
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "create"))

	// Verificar que el archivo main conserva el contenido original
	content, err := os.ReadFile(mainPath)
	require.NoError(env.t, err, "No se pudo leer el archivo main existente")
	require.Equal(env.t, initialContent, string(content), "El contenido del archivo main no debería cambiar tras un evento create")
	require.NotContains(env.t, string(content), expectedContent, "El contenido nuevo no debería estar en el archivo main")

	// Evento WRITE: Ahora SÍ debería actualizar el archivo main
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "write"))

	// Verificar que ahora el archivo main contiene el contenido esperado
	content, err = os.ReadFile(mainPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo main%s", fileExtension))
	require.Contains(env.t, string(content), expectedContent, fmt.Sprintf("El contenido del archivo main%s no es el esperado", fileExtension))

	// Evento DELETE: Debería actualizar el archivo main eliminando el contenido
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName, fileExtension, filePath, "delete"))

	// El archivo main debería seguir existiendo pero sin el contenido del archivo eliminado
	content, err = os.ReadFile(mainPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo main%s", fileExtension))
	require.NotContains(env.t, string(content), expectedContent, fmt.Sprintf("El contenido eliminado no debería estar en main%s", fileExtension))

	// Segunda parte: Probar comportamiento cuando el archivo main NO existe inicialmente
	env.AssetsHandler.WriteOnDisk = false // volver a desactivar emulando el inicio de la libreria
	// Eliminar el archivo main
	require.NoError(env.t, os.Remove(mainPath))

	// Crear un nuevo archivo de prueba
	fileName2 := fmt.Sprintf("new_file%s", fileExtension)
	filePath2 := filepath.Join(env.BaseDir, fileName2)
	var fileContent2 string

	if fileExtension == ".js" {
		fileContent2 = "var newContent = 'when-main-doesnt-exist';"
	} else {
		fileContent2 = ".new-content { background-color: yellow; }"
	}

	require.NoError(env.t, os.WriteFile(filePath2, []byte(fileContent2), 0644))

	// Evento CREATE: Con WriteOnDisk=false, NO debe crear el archivo main (solo almacenar en memoria)
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName2, fileExtension, filePath2, "create"))

	// Verificar que el archivo main NO fue creado aún (WriteOnDisk=false)
	_, err = os.Stat(mainPath)
	require.True(env.t, os.IsNotExist(err), fmt.Sprintf("El archivo main%s NO debería existir después de un evento create con WriteOnDisk=false", fileExtension))

	// Evento WRITE: Ahora SÍ debería crear el archivo main con todo el contenido en memoria
	require.NoError(env.t, os.WriteFile(filePath2, []byte(fileContent2), 0644)) // Asegurar que el archivo existe
	require.NoError(env.t, env.AssetsHandler.NewFileEvent(fileName2, fileExtension, filePath2, "write"))

	// Verificar que el archivo main fue creado con el contenido esperado
	require.FileExists(env.t, mainPath, fmt.Sprintf("El archivo main%s debería existir después de un evento write", fileExtension))
	content, err = os.ReadFile(mainPath)
	require.NoError(env.t, err, fmt.Sprintf("No se pudo leer el archivo main%s", fileExtension))

	// Verificar el contenido apropiado según la extensión
	if fileExtension == ".js" {
		require.Contains(env.t, string(content), "newContent", fmt.Sprintf("El contenido del archivo main%s no es el esperado", fileExtension))
	} else {
		require.Contains(env.t, string(content), "new-content", fmt.Sprintf("El contenido del archivo main%s no es el esperado", fileExtension))
	}
}
