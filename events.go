package assetmin

import (
	"bytes"
	"errors"
	"os"
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
		c.updateAsset(filePath, event, c.cssHandler, file)
		return c.cssHandler, nil

	case ".js":
		c.updateAsset(filePath, event, c.jsHandler, file)
		return c.jsHandler, nil
	}

	return nil, errors.New("UpdateFileContentInMemory extension: " + extension + " not found " + filePath)
}

// assetHandlerFiles ej &jsHandler, &cssHandler
func (c *AssetMin) updateAsset(filePath, event string, assetHandler *fileHandler, newFile *assetFile) {

	filesToUpdate := &assetHandler.moduleFiles

	if strings.Contains(filePath, c.ThemeFolder()) {
		filesToUpdate = &assetHandler.themeFiles
	}

	if event == "remove" {
		if idx := c.findFileIndex(*filesToUpdate, filePath); idx != -1 {
			*filesToUpdate = append((*filesToUpdate)[:idx], (*filesToUpdate)[idx+1:]...)
		}
	} else {
		if idx := c.findFileIndex(*filesToUpdate, filePath); idx != -1 {
			(*filesToUpdate)[idx] = newFile
		} else {
			*filesToUpdate = append(*filesToUpdate, newFile)
		}
	}
}

func (c *AssetMin) findFileIndex(files []*assetFile, filePath string) int {
	for i, f := range files {
		if f.path == filePath {
			return i
		}
	}
	return -1
}

// event: create, remove, write, rename
func (c *AssetMin) NewFileEvent(fileName, extension, filePath, event string) error {
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

	// Enable disk writing on first write or create event
	if (event == "write" || event == "create") && !c.WriteOnDisk {
		c.WriteOnDisk = true
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
	}
}

func (c *AssetMin) startCodeJS() (out string, err error) {
	out = "'use strict';"

	js, err := c.JavascriptForInitializing() // wasm js code
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
