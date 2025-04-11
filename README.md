# 📦 AssetMin

> 🚀 A lightweight and efficient web asset packager and minifier for Go applications

AssetMin is a simple yet powerful tool that bundles and minifies your JavaScript and CSS files into single optimized output files, improving your web application's performance.
## 🛠️ Primary Use Case

AssetMin is primarily used in the [GoDEV](https://github.com/cdvelop/godev) framework for developing full stack projects with Go. It provides an efficient solution for managing and optimizing web assets, ensuring seamless integration into your Go workflow. Whether you're building a small project or a large-scale application, AssetMin simplifies the bundling and minification of your JavaScript and CSS files with minimal effort.

## ✨ Features

- 🔄 **Live Asset Processing** - Automatically processes files when they are created or modified
- 🗜️ **Minification** - Reduces file size by removing unnecessary characters
- 🔌 **Concurrency Support** - Thread-safe operation for multiple file processing
- 📁 **Directory Structure** - Organizes files from theme and module directories
- 🛠️ **Simple API** - Easy to integrate into your Go application

## 📥 Installation

```go
import "github.com/cdvelop/assetmin"
```

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	
	"github.com/cdvelop/assetmin"
)

func main() {
	// Create configuration
	config := &assetmin.AssetConfig{
		// Define theme folder path
		ThemeFolder: func() string { 
			return "./web/theme" 
		},
		
		// Define public folder for output files
		WebFilesFolder: func() string { 
			return "./web/public" 
		},
		
		// Define print function
		Print: func(messages ...any) {
			fmt.Println(messages...)
		},
		
		// Optional JavaScript initialization code
		JavascriptForInitializing: func() (string, error) {
			return "console.log('App initialized!');", nil
		},
	}
	
	// Initialize AssetMin
	handler := assetmin.NewAssetMin(config)
	
	// Process a new JavaScript file
	handler.NewFileEvent("script.js", ".js", "./path/to/script.js", "create")
	
	// Process a new CSS file
	handler.NewFileEvent("styles.css", ".css", "./path/to/styles.css", "create")
	
	// Files will be bundled and minified into:
	// - ./web/public/main.js
	// - ./web/public/style.css
}
```

## 🔄 How It Works

1. 📁 You define theme and output directories
2. 📝 Create or modify JS/CSS files
3. 🔄 Call `NewFileEvent` when files change
4. 📦 AssetMin processes and bundles your files
5. 🚀 Output is saved to your public directory as minified files

## 📋 API Reference

### `NewAssetMin(config *Config) *AssetMin`

Creates a new instance of the AssetMin handler.

### `NewFileEvent(filename, extension, filepath, event string) error`

Processes a file event (create/write).

## 🛠️ Configuration Options

| Option | Description |
|--------|-------------|
| `ThemeFolder` | Function that returns the path to your theme directory |
| `WebFilesFolder` | Function that returns the path to your public output directory |
| `Print` | Function for logging messages |
| `JavascriptForInitializing` | Function that returns initialization JavaScript code |

## 🤝 Contributing

Contributions are welcome! Feel free to submit a Pull Request.

## 📄 License

This project is licensed under the [MIT] License.
