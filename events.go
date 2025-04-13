package assetmin

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func (c *AssetMin) UpdateFileContentInMemory(filePath, extension, event string, content []byte) (*fileHandler, error) {
	file := &assetFile{
		path:    filePath,
		content: content,
	}

	switch extension {
	case ".css":
		c.cssHandler.UpdateContent(filePath, event, file)
		return c.cssHandler, nil

	case ".js":
		c.jsHandler.UpdateContent(filePath, event, file)
		return c.jsHandler, nil

	case ".svg":
		c.svgHandler.UpdateContent(filePath, event, file)
		return c.svgHandler.fileHandler, nil

	case ".html":
		c.htmlHandler.UpdateContent(filePath, event, file)
		return c.htmlHandler.fileHandler, nil
	}

	return nil, errors.New("UpdateFileContentInMemory extension: " + extension + " not found " + filePath)
}

// assetHandlerFiles ej &jsHandler, &cssHandler
func (fh *fileHandler) UpdateContent(filePath, event string, newFile *assetFile) {
	// This function is now handled by the UpdateContent method in each handler
	// Keeping it here for backward compatibility
	filesToUpdate := &fh.moduleFiles

	// Corregir la lógica para identificar correctamente archivos de tema
	// Verificamos si el path contiene "theme" en cualquier parte de la ruta
	if strings.Contains(filePath, fh.themeFolder) {
		filesToUpdate = &fh.themeFiles
	}
	if event == "remove" || event == "delete" {
		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			*filesToUpdate = slices.Delete((*filesToUpdate), idx, idx+1)
		}
	} else {
		if idx := findFileIndex(*filesToUpdate, filePath); idx != -1 {
			(*filesToUpdate)[idx] = newFile
		} else {
			*filesToUpdate = append(*filesToUpdate, newFile)
		}
	}
}

func findFileIndex(files []*assetFile, filePath string) int {
	for i, f := range files {
		if f.path == filePath {
			return i
		}
	}
	return -1
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
	// c.Print("DEBUG:", "writing "+extension+" to disk...")

	var buf bytes.Buffer

	if fh.startCode != nil {
		startCode, err := fh.startCode()
		if err != nil {
			return errors.New(e + err.Error())
		}
		buf.WriteString(startCode)
	}

	// Write theme files first
	for _, f := range fh.themeFiles {
		buf.Write(f.content)
		buf.WriteString("\n") // Add newline between files
	}

	// Then write module files
	for _, f := range fh.moduleFiles {
		buf.Write(f.content)
		buf.WriteString("\n") // Add newline between files
	}

	outputPath := fh.outputPath
	// No need to check directories again, they were created in initialization

	//  Minify and write the buffer
	var minifiedBuf bytes.Buffer
	if err := c.min.Minify(fh.mediatype, &minifiedBuf, &buf); err != nil {
		return errors.New(e + err.Error())
	}
	// c.Print("debug", "writing to disk (minified):", minifiedBuf.String())
	if err := FileWrite(outputPath, minifiedBuf); err != nil {
		return errors.New(e + err.Error())
	}

	return nil
}

func (c *AssetMin) UnobservedFiles() []string {
	// Return the full path of the output files to ignore
	return []string{
		c.cssHandler.outputPath,
		c.jsHandler.outputPath,
		c.svgHandler.outputPath,
		c.htmlHandler.outputPath,
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
func (f *fileHandler) ClearMemoryFiles() {
	f.themeFiles = []*assetFile{}
	f.moduleFiles = []*assetFile{}
}

// isOutputPath checks if the given file path matches any of our output paths
func (c *AssetMin) isOutputPath(filePath string) bool {
	// Normalize paths for cross-platform comparison
	normalizedFilePath := filepath.Clean(filePath)
	cssOutputPath := filepath.Clean(c.cssHandler.outputPath)
	jsOutputPath := filepath.Clean(c.jsHandler.outputPath)
	svgOutputPath := filepath.Clean(c.svgHandler.outputPath)
	htmlHandlerOutputPath := filepath.Clean(c.htmlHandler.outputPath)

	return normalizedFilePath == cssOutputPath ||
		normalizedFilePath == jsOutputPath ||
		normalizedFilePath == svgOutputPath ||
		normalizedFilePath == htmlHandlerOutputPath

}
