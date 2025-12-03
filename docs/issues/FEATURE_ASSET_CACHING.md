# Feature: Asset Caching System

## Overview
Add minified content caching to `asset` struct to avoid re-minification on every HTTP request.

## Related Documents
- [FEATURE_HTTP_ROUTES_WORK_MODES.md](FEATURE_HTTP_ROUTES_WORK_MODES.md) - Uses this cache for HTTP serving

## Requirements

### 1. Cache Fields on asset

```go
type asset struct {
    // ... existing fields ...
    
    cachedMinified []byte  // Minified content ready to serve
    cacheValid     bool    // True if cache matches current content
}
```

### 2. Cache Invalidation

Cache becomes invalid when:
- Any `contentOpen`, `contentMiddle`, or `contentClose` changes via `UpdateContent()`
- `initCode` function returns different content
- Manual invalidation via `InvalidateCache()`

### 3. Cache Regeneration

```go
func (h *asset) RegenerateCache(minifier *minify.M) error
```

**Logic:**
1. Call `WriteContent()` to buffer
2. Minify using provided minifier
3. Store result in `cachedMinified`
4. Set `cacheValid = true`

### 4. Integration with processAsset

Rename `processAndWrite` → `processAsset` (reflects conditional write).

```go
func (c *AssetMin) processAsset(fh *asset, context string) error {
    // 1. Always regenerate cache
    if err := fh.RegenerateCache(c.min); err != nil {
        return err
    }
    
    // 2. Write to disk only if DiskMode
    if c.workMode == DiskMode {
        return FileWrite(fh.outputPath, bytes.NewBuffer(fh.cachedMinified))
    }
    return nil
}
```

**Current flow:**
```
NewFileEvent → UpdateContent → processAndWrite → Minify → Write to disk
```

**New flow:**
```
NewFileEvent → UpdateContent (invalidate cache) → processAsset → RegenerateCache → (if DiskMode) Write
```

### 5. HTTP Serving from Cache

```go
func (h *asset) GetMinifiedContent(minifier *minify.M) ([]byte, error) {
    if !h.cacheValid {
        if err := h.RegenerateCache(minifier); err != nil {
            return nil, err
        }
    }
    return h.cachedMinified, nil
}
```

## Files to Modify

| File | Changes |
|------|---------|
| `asset.go` | Add cache fields, `RegenerateCache`, `InvalidateCache`, `GetMinifiedContent` |
| `events.go` | Rename `processAndWrite` → `processAsset`, invalidate on `UpdateContent` |

## Cache Lifecycle

```
┌─────────────────┐
│  Content Change │
│  (UpdateContent)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ cacheValid=false│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  processAsset   │
│  called         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│RegenerateCache  │
│ - WriteContent  │
│ - Minify        │
│ - Store cache   │
│ - cacheValid=T  │
└────────┬────────┘
         │
    ┌────┴────┐
    │DiskMode?│
    └────┬────┘
    Yes  │  No
    ▼    │   ▼
┌───────┐│┌──────┐
│Write  │││Done  │
│to disk│││      │
└───────┘│└──────┘
         │
         ▼
┌─────────────────┐
│  HTTP Request   │
│  Serve from     │
│  cachedMinified │
└─────────────────┘
```

## Memory Considerations
- Cache increases memory usage per asset
- Typical sizes: CSS ~10KB, JS ~50KB, HTML ~5KB, SVG ~20KB
- Total additional memory: ~100KB typical, acceptable for dev server

## Concurrency
- Create simple concurrency test first
- If race conditions detected, add `sync.RWMutex` to `asset`
- RWMutex allows concurrent HTTP reads, exclusive cache regeneration
