# Feature: TUI Handler Integration

## Overview
Implement `HandlerExecution` interface to allow AssetMin work mode toggle via TUI (devtui).

## Related Documents
- [FEATURE_HTTP_ROUTES_WORK_MODES.md](FEATURE_HTTP_ROUTES_WORK_MODES.md) - Defines work modes

## Interface Definition

```go
// HandlerExecution defines interface for executable actions in TUI
type HandlerExecution interface {
    Name() string                   // Identifier: "ASSETS"
    Label() string                  // Display label: current mode description
    Execute(progress chan<- string) // Toggle mode + report via progress
}
```

## Mode Naming Decision

**Chosen**: `"MEMORY"` / `"DISK"`

**Justification**:
- "SSR" is a specific feature (Server-Side Rendering), not a mode of operation
- "MEMORY" accurately describes where assets are served from
- "DISK" clearly indicates file system writes
- Avoids confusion: SSR can work in both modes (memory for dev, disk for deploy)
- Consistent with common terminology (in-memory cache vs disk persistence)

**Display format**: `"Asset Output: MEMORY"` / `"Asset Output: DISK"`

## Implementation on AssetMin

### File: `tui.go`

```go
package assetmin

// Name returns the handler identifier for TUI
func (c *AssetMin) Name() string {
    return "ASSETS"
}

// Label returns current mode description for TUI display
func (c *AssetMin) Label() string {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.workMode == MemoryMode {
        return "Asset Output: MEMORY"
    }
    return "Asset Output: DISK"
}

// Execute toggles between work modes and reports progress
// NOTE: Caller (devtui) owns the channel and will close it
func (c *AssetMin) Execute(progress chan<- string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.workMode == MemoryMode {
        c.workMode = DiskMode
        c.writeAllCacheToDisk(progress)
        progress <- "Switched to DISK mode - assets written to disk"
    } else {
        c.workMode = MemoryMode
        progress <- "Switched to MEMORY mode - serving from cache"
    }
}
```

### Helper Method

```go
func (c *AssetMin) writeAllCacheToDisk(progress chan<- string) {
    assets := []*asset{
        c.indexHtmlHandler,
        c.mainStyleCssHandler,
        c.mainJsHandler,
        c.spriteSvgHandler,
        c.faviconSvgHandler,
    }
    
    for _, a := range assets {
        if a.cacheValid && len(a.cachedMinified) > 0 {
            if err := FileWrite(a.outputPath, bytes.NewBuffer(a.cachedMinified)); err != nil {
                progress <- "Error writing " + a.fileOutputName + ": " + err.Error()
            } else {
                progress <- "Written: " + a.fileOutputName
            }
        }
    }
}
```

## TUI Display

In devtui, the handler appears as:

```
┌──────────────────────────────────────┐
│ [A] ASSETS: Asset Output: MEMORY    │
└──────────────────────────────────────┘
```

Pressing `A` toggles:

```
┌──────────────────────────────────────┐
│ [A] ASSETS: Asset Output: DISK      │
│                                      │
│ Written: style.css                   │
│ Written: script.js                   │
│ Written: icons.svg                   │
│ Written: index.html                  │
│ Written: favicon.svg                 │
│ Switched to DISK mode                │
└──────────────────────────────────────┘
```

## Files to Create

| File | Description |
|------|-------------|
| `tui.go` | `Name()`, `Label()`, `Execute()` methods |

## Integration Notes

1. devtui discovers handlers via interface check
2. AssetMin registers automatically when passed to devtui
3. No explicit registration needed - interface satisfaction is enough
