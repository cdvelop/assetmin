# Feature: HTTP Routes and Work Modes

## Overview
Add HTTP serving capability to AssetMin with configurable work modes (Memory/Disk).

## Related Documents
- [FEATURE_CONFIG_SIMPLIFICATION.md](FEATURE_CONFIG_SIMPLIFICATION.md) - Prerequisite: Config changes
- [FEATURE_ASSET_CACHING.md](FEATURE_ASSET_CACHING.md) - Prerequisite: cache system
- [FEATURE_TUI_HANDLER.md](FEATURE_TUI_HANDLER.md) - Mode toggle via TUI

## Requirements

### 1. Work Modes
Two mutually exclusive modes controlled globally:

```go
type WorkMode int

const (
    MemoryMode WorkMode = iota  // Serve from memory cache (default)
    DiskMode                     // Write to disk + serve from cache
)
```

**Behavior:**
- `MemoryMode`: Assets served via HTTP from minified cache. No disk writes.
- `DiskMode`: On content change, write to disk AND update cache. HTTP serves from cache.

**Mode Toggle:**
- Switching `MemoryMode → DiskMode`: Immediately write all cached content to disk.
- Switching `DiskMode → MemoryMode`: Stop writing to disk, continue serving from cache.

### 2. Configuration Changes

```go
type Config struct {
    OutputDir               string                 // Single output directory
    Logger                  func(message ...any)
    GetRuntimeInitializerJS func() (string, error)
    AppName                 string
    
    // AssetsURLPrefix is the URL prefix for static assets (CSS, JS, SVG).
    // Example: "/assets/" or "/static/"
    // index.html is ALWAYS served at "/" regardless of this prefix.
    // Default: "" (assets served from root: /style.css, /script.js)
    AssetsURLPrefix string
}
```

### 3. HTTP Route Registration

New method on `AssetMin`:

```go
func (c *AssetMin) RegisterRoutes(mux *http.ServeMux)
```

**Route mapping:**
| Asset | Route (no prefix) | Route (prefix="/assets/") |
|-------|-------------------|---------------------------|
| index.html | `GET /` | `GET /` |
| style.css | `GET /style.css` | `GET /assets/style.css` |
| script.js | `GET /script.js` | `GET /assets/script.js` |
| icons.svg | `GET /icons.svg` | `GET /assets/icons.svg` |
| favicon.svg | `GET /favicon.svg` | `GET /assets/favicon.svg` |

**Note:** `index.html` NEVER uses prefix to avoid ugly `/assets/index.html` URLs.

### 4. ServeHTTP on asset

Each `asset` implements HTTP serving:

```go
func (h *asset) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

**Logic:**
1. Check if `cachedMinified` is valid
2. If valid: serve cached content with correct `Content-Type`
3. If invalid: regenerate cache, then serve
4. Set appropriate headers: `Content-Type`, `Cache-Control`

### 5. URL Path Generation

New field on `asset`:

```go
type asset struct {
    // ... existing fields ...
    urlPath string  // HTTP route path, e.g., "/assets/style.css" or "/style.css"
}
```

Computed at initialization based on `AssetsURLPrefix` and `fileOutputName`.

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `http.go` | Create | `RegisterRoutes`, HTTP helpers |
| `asset.go` | Modify | Add `urlPath`, `ServeHTTP` |
| `assetmin.go` | Modify | Add `WorkMode`, remove `WriteOnDisk` |
| `events.go` | Modify | Update `UnobservedFiles`, refactor mode logic |

## UnobservedFiles Behavior

Only truly generated/merged files should be unobserved:

```go
func (c *AssetMin) UnobservedFiles() []string {
    return []string{
        c.mainStyleCssHandler.outputPath,  // Merged CSS
        c.mainJsHandler.outputPath,        // Merged JS  
        c.spriteSvgHandler.outputPath,     // Merged SVG sprite
    }
}
```

**Excluded from unobserved** (user may edit):
- `index.html` - User adds links, meta tags
- `favicon.svg` - User provides custom favicon

## API Changes

### New Public Methods on AssetMin
```go
func (c *AssetMin) RegisterRoutes(mux *http.ServeMux)
func (c *AssetMin) SetWorkMode(mode WorkMode)
func (c *AssetMin) GetWorkMode() WorkMode
```

### New Public Methods on asset (internal use)
```go
func (h *asset) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (h *asset) URLPath() string
```

## Template Updates Required
See [FEATURE_TEMPLATE_REFACTOR.md](FEATURE_TEMPLATE_REFACTOR.md) for dynamic URL handling in templates.

## Breaking Changes
- `WriteOnDisk bool` field **removed** completely. Replaced by `WorkMode`.
- `ThemeFolder` field **removed**. See [FEATURE_CONFIG_SIMPLIFICATION.md](FEATURE_CONFIG_SIMPLIFICATION.md).
- `OutputDir` changed from `func() string` to `string`.
- `processAndWrite` renamed to `processAsset`.
- Default is `MemoryMode` (memory-only, HTTP serving).
- Mode switching is **manual only** (no auto-switch on events).
