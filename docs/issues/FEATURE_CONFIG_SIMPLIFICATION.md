# Feature: Config Simplification (PRIORITY 1)

## Overview
Simplify `Config` struct by removing `ThemeFolder` and changing `OutputDir` to string.

## Breaking Changes

### 1. Remove ThemeFolder
**Reason**: Components are registered dynamically via `NewFileEvent`, folder structure is irrelevant to assetmin.

```go
// BEFORE
type Config struct {
    ThemeFolder func() string  // REMOVE
    OutputDir   func() string
    // ...
}

// AFTER
type Config struct {
    OutputDir string  // Changed to string
    // ...
}
```

### 2. OutputDir as String
**Reason**: Output directory doesn't change at runtime. Simpler API.

```go
// BEFORE
OutputDir func() string  // eg: func() { return "web/public" }

// AFTER  
OutputDir string  // eg: "web/public"
```

### 3. Remove EnsureOutputDirectoryExists from NewAssetMin
**Reason**: With `MemoryMode` default, directory may not be needed. Create only in `DiskMode`.

```go
// REMOVE from NewAssetMin():
// c.EnsureOutputDirectoryExists()

// KEEP method but call only when switching to DiskMode
func (c *AssetMin) EnsureOutputDirectoryExists() { ... }
```

### 4. Remove NotifyIfOutputFilesExist
**Reason**: With `MemoryMode` default and no ThemeFolder, checking disk files at startup is irrelevant.

```go
// REMOVE entirely:
// - NotifyIfOutputFilesExist() method
// - notifyMeIfOutputFileExists field from asset struct
// - All callback registrations
```

### 5. Remove themeFolder from asset struct
**Reason**: No longer needed without ThemeFolder config.

```go
// BEFORE
type asset struct {
    themeFolder string  // REMOVE
    // ...
}

// AFTER
type asset struct {
    // themeFolder removed
    // ...
}
```

## New Workflow

1. **Single directory**: All output goes to `OutputDir`
2. **No duplication**: User places custom files directly in `OutputDir`
3. **Registration via events**: All content arrives via `NewFileEvent`
4. **filePath = source**: `NewFileEvent.filePath` is the source file path
5. **Ignore existing files**: Files in `OutputDir` at startup are ignored unless registered

## Directory Structure Example

```
src/components/          <- Source files (watched by devwatch)
├── button.css
├── header.js
└── icon.svg

web/public/              <- OutputDir (output only)
├── index.html           <- Generated OR user-edited
├── favicon.svg          <- User provides
├── style.css            <- Generated (merged CSS)
├── script.js            <- Generated (merged JS)
└── icons.svg            <- Generated (merged SVG sprite)
```

## Code Changes Required

### assetmin.go

**Remove:**
```go
// In NewAssetMin():
c.EnsureOutputDirectoryExists()  // Remove call
c.NotifyIfOutputFilesExist()     // Remove call
// Remove WriteOnDisk auto-enable based on file existence
```

**Change Config:**
```go
type Config struct {
    OutputDir               string                 // Changed from func() string
    Logger                  func(message ...any)
    GetRuntimeInitializerJS func() (string, error)
    AppName                 string
    AssetsURLPrefix         string                 // New: for HTTP routes
}
```

**Remove methods:**
```go
// Remove entirely:
func (c *AssetMin) NotifyIfOutputFilesExist() { ... }
func (c *AssetMin) MainInputFileRelativePath() string { ... }  // Uses themeFolder
```

### asset.go

**Remove field:**
```go
type asset struct {
    fileOutputName string
    outputPath     string
    mediatype      string
    initCode       func() (string, error)
    // themeFolder    string  <- REMOVE
    // notifyMeIfOutputFileExists func(content string)  <- REMOVE

    contentOpen   []*contentFile
    contentMiddle []*contentFile
    contentClose  []*contentFile
}
```

**Update newAssetFile:**
```go
func newAssetFile(outputName, mediaType string, ac *Config, initCode func() (string, error)) *asset {
    handler := &asset{
        fileOutputName: outputName,
        outputPath:     filepath.Join(ac.OutputDir, outputName),  // Direct string
        mediatype:      mediaType,
        initCode:       initCode,
        // themeFolder removed
        contentOpen:    []*contentFile{},
        contentMiddle:  []*contentFile{},
        contentClose:   []*contentFile{},
        // notifyMeIfOutputFileExists removed
    }
    return handler
}
```

**Update UpdateContent - remove themeFolder checks:**
```go
func (h *asset) UpdateContent(filePath, event string, f *contentFile) (err error) {
    // REMOVE: All checks for h.themeFolder
    // REMOVE: strings.Contains(filePath, h.themeFolder)
    
    // SIMPLIFIED: All files go to contentMiddle unless HTML complete doc
    filesToUpdate := &h.contentMiddle
    
    // Keep HTML complete document check
    if strings.HasSuffix(h.fileOutputName, ".html") && strings.HasSuffix(filePath, ".html") {
        if isCompleteHtmlDocument(string(f.content)) {
            return nil  // Ignore complete HTML docs
        }
    }
    // ... rest of logic
}
```

### html.go

**Update NewHtmlHandler:**
```go
func NewHtmlHandler(ac *Config, outputName, cssName, jsName string) *asset {
    af := newAssetFile(outputName, "text/html", ac, nil)
    // Remove: af.notifyMeIfOutputFileExists = hh.notifyMeIfOutputFileExists
    // ...
}
```

**Remove:**
```go
// Remove notifyMeIfOutputFileExists method from htmlHandler
```

### svg.go
No changes needed - doesn't use themeFolder.

### htmlGenerator.go
**Remove entirely** or gut completely - methods write to disk which violates MemoryMode.

## Files to Modify

| File | Changes |
|------|---------|
| `assetmin.go` | Remove ThemeFolder, change OutputDir type, remove NotifyIfOutputFilesExist |
| `asset.go` | Remove themeFolder field, notifyMeIfOutputFileExists, update newAssetFile |
| `html.go` | Remove notification callback, simplify handler |
| `htmlGenerator.go` | Remove or mark for later removal |
| `events.go` | Update OutputDir access from function to string |
| All tests | Update config initialization |

## Migration for Consumers

### Before
```go
config := &assetmin.Config{
    ThemeFolder: func() string { return "web/theme" },
    OutputDir:   func() string { return "web/public" },
}
am := assetmin.NewAssetMin(config)
```

### After
```go
config := &assetmin.Config{
    OutputDir: "web/public",
}
am := assetmin.NewAssetMin(config)
```

## Impact on Other Features

| Feature | Impact |
|---------|--------|
| Asset Caching | No themeFolder checks in UpdateContent |
| HTTP Routes | Uses `OutputDir` string directly |
| Template Refactor | No theme folder logic to remove |
| TUI Handler | No impact |
| SSR | No impact |
