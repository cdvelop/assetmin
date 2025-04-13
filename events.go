package assetmin

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
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
		err := c.mainJsHandler.UpdateContent(filePath, event, file)
		return c.mainJsHandler, err

	case ".svg":
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
		c.Print("Skipping output file:", filePath)
		return nil
	}

	c.mu.Lock()         // Lock the mutex at the beginning
	defer c.mu.Unlock() // Ensure mutex is unlocked when the function returns

	var e = "NewFileEvent " + extension + " " + event
	if filePath == "" {
		return errors.New(e + "filePath is empty")
	}

	c.Print("Asset", extension, event, "...", filePath)

	// Increase sleep duration significantly to allow file system operations (like write after rename) to settle
	// fail when time is < 10ms
	time.Sleep(20 * time.Millisecond) // Increased from 10ms

	// read file content from filePath
	content, err := os.ReadFile(filePath)
	if err != nil {
		return errors.New(e + err.Error())
	}

	fh, err := c.UpdateFileContentInMemory(filePath, extension, event, content)
	if err != nil {
		return errors.New(e + err.Error())
	}
	// Check event type and file existence to determine if we should write to disk
	if !c.WriteOnDisk {
		// If the file doesn't exist, create it regardless of event type
		if _, err := os.Stat(fh.outputPath); os.IsNotExist(err) {
			c.WriteOnDisk = true
		} else if err == nil {
			// File exists, only update on write or delete events
			if event == "write" || event == "remove" || event == "delete" {
				c.WriteOnDisk = true
			}
		}
	}

	if !c.WriteOnDisk {
		return nil
	}

	// Process content into a buffer
	var buf bytes.Buffer
	fh.WriteContent(&buf)

	// Minify the content
	var minifiedBuf bytes.Buffer
	if err := c.min.Minify(fh.mediatype, &minifiedBuf, &buf); err != nil {
		return errors.New(e + " Minification error: " + err.Error())
	}

	// Write to disk
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
		c.indexHtmlHandler.outputPath,
	}
}

func (c *AssetMin) startCodeJS() (out string, err error) {
	out = "'use strict';"

	js, err := c.GetRuntimeInitializerJS() // wasm js code
	if err != nil {
		return "", errors.New("startCodeJS " + err.Error())
	}
	out += js

	return
}

// clear memory files
func (f *asset) ClearMemoryFiles() {
	f.contentOpen = []*contentFile{}
	f.contentMiddle = []*contentFile{}
	f.contentClose = []*contentFile{}
}

// isOutputPath checks if the given file path matches any of our output paths
func (c *AssetMin) isOutputPath(filePath string) bool {
	// Normalize paths for cross-platform comparison
	normalizedFilePath := filepath.Clean(filePath)
	cssOutputPath := filepath.Clean(c.mainStyleCssHandler.outputPath)
	jsOutputPath := filepath.Clean(c.mainJsHandler.outputPath)
	svgOutputPath := filepath.Clean(c.spriteSvgHandler.outputPath)
	htmlHandlerOutputPath := filepath.Clean(c.indexHtmlHandler.outputPath)

	return normalizedFilePath == cssOutputPath ||
		normalizedFilePath == jsOutputPath ||
		normalizedFilePath == svgOutputPath ||
		normalizedFilePath == htmlHandlerOutputPath

}
