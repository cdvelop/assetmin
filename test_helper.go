package assetmin

import (
	"fmt"
	"os"
	"path/filepath"
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
