# AssetMin Roadmap

This document outlines planned features and improvements for AssetMin.

## Completed Features ✅

- ✅ Config Simplification (removed `ThemeFolder`, simplified `OutputDir`)
- ✅ Asset Caching System (in-memory minified content cache)
- ✅ HTTP Routes & Work Modes (MemoryMode/DiskMode)
- ✅ Thread-safe operations with mutex locks

## Planned Features

### Template System Refactor

**Status**: Planned  
**Priority**: Medium  
**Document**: [issues/FEATURE_TEMPLATE_REFACTOR.md](issues/FEATURE_TEMPLATE_REFACTOR.md)

Improve the HTML template system to:
- Use `embed.FS` for easier template editing
- Support dynamic URL path injection
- Better separation between template and generation logic

### TUI Integration

**Status**: Planned  
**Priority**: Low  
**Document**: [issues/FEATURE_TUI_HANDLER.md](issues/FEATURE_TUI_HANDLER.md)

Add Terminal UI for:
- Real-time asset monitoring
- Work mode toggling
- Build statistics
- File event visualization

### SSR Support

**Status**: Planned  
**Priority**: Low  
**Document**: [issues/SSR_IMPLEMENTATION_DETAILS.md](issues/SSR_IMPLEMENTATION_DETAILS.md)

Enhanced server-side rendering capabilities:
- Component-based SSR
- Streaming HTML generation
- Better integration with Go templates

### PWA Support

**Status**: Planned  
**Priority**: Low  
**Document**: [issues/PWA_SUPPORT.md](issues/PWA_SUPPORT.md)

Progressive Web App features:
- Service worker generation
- Manifest file generation
- Offline support
- Cache strategies

## Feature Requests

Have an idea for AssetMin? Please:

1. Check existing [issues](https://github.com/tinywasm/assetmin/issues)
2. Open a new issue with the `feature-request` label
3. Describe your use case and proposed solution

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to AssetMin.
