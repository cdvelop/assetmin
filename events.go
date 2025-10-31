package assetmin

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (c *AssetMin) UpdateFileContentInMemory(filePath, extension, event string, content []byte) (*asset, error) {
	file := &contentFile{
		path:    filePath,
		content: content,
	}

	switch extension {
	case ".css":
		err := c.mainStyleCssHandler.UpdateContent(filePath, event, file)
		return c.mainStyleCssHandler, err

	case ".js":
		// Remove a leading "use strict" directive from incoming files to avoid
		// duplicating the directive which we add globally in startCodeJS.
		file.content = stripLeadingUseStrict(file.content)
		err := c.mainJsHandler.UpdateContent(filePath, event, file)
		return c.mainJsHandler, err

	case ".svg":
		// Check if it's the favicon file
		if filepath.Base(filePath) == c.svgFaviconFileName {
			err := c.faviconSvgHandler.UpdateContent(filePath, event, file)
			return c.faviconSvgHandler, err
		}
		// Otherwise treat as sprite icon
		err := c.spriteSvgHandler.UpdateContent(filePath, event, file)
		return c.spriteSvgHandler, err

	case ".html":
		err := c.indexHtmlHandler.UpdateContent(filePath, event, file)
		return c.indexHtmlHandler, err
	}

	return nil, errors.New("UpdateFileContentInMemory extension: " + extension + " not found " + filePath)
}

// event: create, remove, write, rename
func (c *AssetMin) NewFileEvent(fileName, extension, filePath, event string) error {
	// Check if filePath matches any of our output paths to avoid infinite recursion
	if c.isOutputPath(filePath) {
		//c.writeMessage("Skipping output file:", filePath)
		return nil
	}

	c.mu.Lock()         // Lock the mutex at the beginning
	defer c.mu.Unlock() // Ensure mutex is unlocked when the function returns

	var e = "NewFileEvent " + extension + " " + event
	if filePath == "" {
		return errors.New(e + "filePath is empty")
	}

	c.writeMessage(extension, event, "...", filePath)

	// Debug trace: log current WriteOnDisk state with timestamp
	// c.writeMessage("DEBUG [", time.Now().Format("15:04:05.000"), "] WriteOnDisk=", c.WriteOnDisk, "for event:", event)

	// Increase sleep duration significantly to allow file system operations (like write after rename) to settle
	// fail when time is < 10ms
	time.Sleep(20 * time.Millisecond) // Increased from 10ms

	var content []byte
	var err error

	// For delete/remove events, we don't need to read file content since file no longer exists
	if event == "remove" || event == "delete" {
		// c.writeMessage("DEBUG processing delete event, skipping file read")
		content = []byte{} // Empty content for delete events
	} else {
		// read file content from filePath for other events
		content, err = os.ReadFile(filePath)
		if err != nil {
			return errors.New(e + err.Error())
		}
	}

	fh, err := c.UpdateFileContentInMemory(filePath, extension, event, content) // Update contentMiddle
	if err != nil {
		return errors.New(e + err.Error())
	}

	// Log handler and memory state
	if fh != nil {
		// report counts of content arrays if available
		var memInfo string
		memInfo = "hasContentInMemory="
		if fh.hasContentInMemory() {
			memInfo += "true"
		} else {
			memInfo += "false"
		}
		// Debug: show detailed memory state
		// c.writeMessage("DEBUG handlerOutput=", fh.outputPath, memInfo)
		// c.writeMessage("DEBUG memory state - open:", len(fh.contentOpen), "middle:", len(fh.contentMiddle), "close:", len(fh.contentClose))

		// Log file paths in memory (omitted to reduce debug noise)
	}

	// Check event type and file existence to determine if we should write to disk
	if !c.WriteOnDisk {
		// Only enable writing for write/delete events, never for create events
		// This ensures that:
		// 1. During InitialRegistration: create events only store in memory
		// 2. After InitialRegistration: write events enable compilation with all memory content
		// 3. Post-deletion scenarios: the first write (not create) will trigger compilation
		if event == "write" || event == "remove" || event == "delete" {
			c.WriteOnDisk = true
		}
		// Create events are always memory-only when WriteOnDisk=false
		// This prevents premature writing during InitialRegistration
	}

	if !c.WriteOnDisk {
		return nil
	}

	// c.writeMessage("DEBUG proceeding to write to disk; WriteOnDisk=", c.WriteOnDisk)

	// Process content into a buffer
	var buf bytes.Buffer
	fh.WriteContent(&buf)

	// Debug: log content counts and preview
	bufLen := buf.Len()
	previewLen := bufLen
	if previewLen > 100 {
		previewLen = 100
	}
	// c.writeMessage("DEBUG contentOpen=", len(fh.contentOpen), "contentMiddle=", len(fh.contentMiddle), "contentClose=", len(fh.contentClose))
	// c.writeMessage("DEBUG raw buffer size=", bufLen, "preview:", string(buf.Bytes()[:previewLen]))

	// Minify the content
	var minifiedBuf bytes.Buffer
	if err := c.min.Minify(fh.mediatype, &minifiedBuf, &buf); err != nil {
		return errors.New(e + " Minification error: " + err.Error())
	}

	// Write to disk
	// c.writeMessage("DEBUG outputPath=", fh.outputPath, "minifiedSize=", minifiedBuf.Len())
	minifiedLen := minifiedBuf.Len()
	contentPreviewLen := minifiedLen
	if contentPreviewLen > 200 {
		contentPreviewLen = 200
	}
	// c.writeMessage("DEBUG writing content:", string(minifiedBuf.Bytes()[:contentPreviewLen]))
	if err := FileWrite(fh.outputPath, minifiedBuf); err != nil {
		return errors.New(e + " File write error: " + err.Error())
	}

	return nil
}

func (c *AssetMin) UnobservedFiles() []string {
	// Return the full path of the output files to ignore
	return []string{
		c.mainStyleCssHandler.outputPath,
		c.mainJsHandler.outputPath,
		c.spriteSvgHandler.outputPath,
		c.faviconSvgHandler.outputPath,
		c.indexHtmlHandler.outputPath,
	}
}

func (c *AssetMin) startCodeJS() (out string, err error) {
	out = "'use strict';"

	js, err := c.GetRuntimeInitializerJS() // wasm js code
	if err != nil {
		return "", errors.New("startCodeJS " + err.Error())
	}

	// Remove any leading 'use strict' in the initializer to avoid duplication.
	// The initializer comes from GetRuntimeInitializerJS and doesn't go through
	// UpdateFileContentInMemory, so we need to clean it here.
	clean := stripLeadingUseStrict([]byte(js))
	out += string(clean)

	return
}

// clear memory files
func (f *asset) ClearMemoryFiles() {
	f.contentOpen = []*contentFile{}
	f.contentMiddle = []*contentFile{}
	f.contentClose = []*contentFile{}
}

// hasContentInMemory checks if the asset has any content stored in memory
func (f *asset) hasContentInMemory() bool {
	return len(f.contentOpen) > 0 || len(f.contentMiddle) > 0 || len(f.contentClose) > 0
}

// isOutputPath checks if the given file path matches any of our output paths
func (c *AssetMin) isOutputPath(filePath string) bool {
	// Normalize paths for cross-platform comparison
	normalizedFilePath := filepath.Clean(filePath)
	cssOutputPath := filepath.Clean(c.mainStyleCssHandler.outputPath)
	jsOutputPath := filepath.Clean(c.mainJsHandler.outputPath)
	svgOutputPath := filepath.Clean(c.spriteSvgHandler.outputPath)
	faviconOutputPath := filepath.Clean(c.faviconSvgHandler.outputPath)
	htmlHandlerOutputPath := filepath.Clean(c.indexHtmlHandler.outputPath)

	// Case-sensitive comparison first
	if normalizedFilePath == cssOutputPath ||
		normalizedFilePath == jsOutputPath ||
		normalizedFilePath == svgOutputPath ||
		normalizedFilePath == faviconOutputPath ||
		normalizedFilePath == htmlHandlerOutputPath {
		return true
	}

	// Case-insensitive comparison for cross-platform compatibility
	normalizedFilePathLower := strings.ToLower(normalizedFilePath)
	cssOutputPathLower := strings.ToLower(cssOutputPath)
	jsOutputPathLower := strings.ToLower(jsOutputPath)
	svgOutputPathLower := strings.ToLower(svgOutputPath)
	faviconOutputPathLower := strings.ToLower(faviconOutputPath)
	htmlHandlerOutputPathLower := strings.ToLower(htmlHandlerOutputPath)

	return normalizedFilePathLower == cssOutputPathLower ||
		normalizedFilePathLower == jsOutputPathLower ||
		normalizedFilePathLower == svgOutputPathLower ||
		normalizedFilePathLower == faviconOutputPathLower ||
		normalizedFilePathLower == htmlHandlerOutputPathLower
}
