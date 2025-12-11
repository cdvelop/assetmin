# AssetMin Feature Planning

This directory contains detailed specifications for planned features.

## ✅ Completed Features

The following features have been successfully implemented:

### 1. Config Simplification
- ✅ Removed `ThemeFolder` field
- ✅ Changed `OutputDir` from `func() string` to `string`
- ✅ Removed `WriteOnDisk bool`, replaced with `WorkMode`
- ✅ Removed `NotifyIfOutputFilesExist()` method
- ✅ Simplified asset initialization

**See**: Current implementation in [`assetmin.go`](../../assetmin.go)

### 2. Asset Caching System
- ✅ Added `cachedMinified` and `cacheValid` fields to asset struct
- ✅ Implemented `RegenerateCache()` method
- ✅ Implemented `GetMinifiedContent()` with double-checked locking
- ✅ Cache invalidation on content updates
- ✅ Thread-safe cache access with RWMutex

**See**: Current implementation in [`asset.go`](../../asset.go)

### 3. HTTP Routes & Work Modes
- ✅ Implemented `MemoryMode` and `DiskMode`
- ✅ Added `RegisterRoutes()` method
- ✅ HTTP serving from cached content
- ✅ Configurable `AssetsURLPrefix`
- ✅ URL path generation for assets
- ✅ `SetWorkMode()` and `GetWorkMode()` methods

**See**: Current implementation in [`http.go`](../../http.go) and [`assetmin.go`](../../assetmin.go)

## 📋 Planned Features

### 4. Template System Refactor
**Status**: Planned  
**Priority**: Medium  
**Document**: [FEATURE_TEMPLATE_REFACTOR.md](FEATURE_TEMPLATE_REFACTOR.md)

**Goals**:
- Use `embed.FS` for easier template editing
- Support dynamic URL path injection in templates
- Better separation between template and generation logic
- Improve HTML generation flexibility

### 5. TUI Integration
**Status**: Planned  
**Priority**: Low  
**Document**: [FEATURE_TUI_HANDLER.md](FEATURE_TUI_HANDLER.md)

**Goals**:
- Real-time asset monitoring dashboard
- Work mode toggling via TUI
- Build statistics and metrics
- File event visualization
- Interactive debugging

### 6. SSR Support
**Status**: Planned  
**Priority**: Low  
**Document**: [SSR_IMPLEMENTATION_DETAILS.md](SSR_IMPLEMENTATION_DETAILS.md)

**Goals**:
- Enhanced server-side rendering capabilities
- Component-based SSR
- Streaming HTML generation
- Better integration with Go templates

### 7. PWA Support
**Status**: Planned  
**Priority**: Low  
**Document**: [PWA_SUPPORT.md](PWA_SUPPORT.md)

**Goals**:
- Service worker generation
- Web app manifest generation
- Offline support
- Cache strategies for PWAs

## 🗂️ Archived Documents

The following documents have been superseded by implemented features:

| Document | Status | Notes |
|----------|--------|-------|
| [FEATURE_CONFIG_SIMPLIFICATION.md](FEATURE_CONFIG_SIMPLIFICATION.md) | ✅ Implemented | Config is now simplified |
| [FEATURE_ASSET_CACHING.md](FEATURE_ASSET_CACHING.md) | ✅ Implemented | Caching system is complete |
| [FEATURE_HTTP_ROUTES_WORK_MODES.md](FEATURE_HTTP_ROUTES_WORK_MODES.md) | ✅ Implemented | HTTP serving is functional |
| [PROMPT_REFACTOR.md](PROMPT_REFACTOR.md) | 🗄️ Archived | Merged into completed features |
| [PROMPT_REFACTOR_V2.md](PROMPT_REFACTOR_V2.md) | 🗄️ Archived | Merged into completed features |
| [PROMPT_REFACTOR_OBSERVATION_AND_SSR.md](PROMPT_REFACTOR_OBSERVATION_AND_SSR.md) | 🗄️ Archived | Split into active features |

## 📖 Documentation

For current API documentation, see:
- [API Documentation](../API.md) - Complete API reference
- [Roadmap](../ROADMAP.md) - High-level feature roadmap
- [Main README](../../README.md) - Project overview

## 🤝 Contributing

When proposing new features:

1. Create a new document in this directory: `FEATURE_[NAME].md`
2. Follow the existing document structure:
   - Overview
   - Requirements
   - Implementation details
   - Files to modify
   - Breaking changes (if any)
3. Link related documents
4. Update this README with the new feature

See [CONTRIBUTING.md](../CONTRIBUTING.md) for general contribution guidelines.
