# Archived Feature Documents

This directory contains feature planning documents for features that have been successfully implemented.

These documents are kept for historical reference and to understand the design decisions that went into the current implementation.

## Archived Documents

### FEATURE_CONFIG_SIMPLIFICATION.md
**Status**: ✅ Implemented  
**Implementation**: See [`assetmin.go`](../../../assetmin.go) and [`asset.go`](../../../asset.go)

Key changes implemented:
- Removed `ThemeFolder` field
- Changed `OutputDir` from `func() string` to `string`
- Removed `WriteOnDisk bool`, replaced with `WorkMode`
- Removed `NotifyIfOutputFilesExist()` method

### FEATURE_ASSET_CACHING.md
**Status**: ✅ Implemented  
**Implementation**: See [`asset.go`](../../../asset.go)

Key features implemented:
- In-memory cache for minified content
- `RegenerateCache()` method
- `GetMinifiedContent()` with double-checked locking
- Thread-safe cache access with RWMutex

### FEATURE_HTTP_ROUTES_WORK_MODES.md
**Status**: ✅ Implemented  
**Implementation**: See [`http.go`](../../../http.go) and [`assetmin.go`](../../../assetmin.go)

Key features implemented:
- `MemoryMode` and `DiskMode` work modes
- `RegisterRoutes()` for HTTP serving
- Configurable `AssetsURLPrefix`
- URL path generation for assets

## Current Documentation

For up-to-date API documentation, see:
- [API Documentation](../../API.md)
- [Roadmap](../../ROADMAP.md)
- [Active Features](../README.md)
