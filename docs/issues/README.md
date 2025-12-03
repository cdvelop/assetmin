# AssetMin Feature Documentation

This directory contains specifications for pending and in-progress features.

## ⚠️ Breaking Changes Summary
- `ThemeFolder` field **removed**
- `OutputDir` changed from `func() string` to `string`
- `WriteOnDisk bool` replaced with `WorkMode`
- `NotifyIfOutputFilesExist()` **removed**
- `notifyMeIfOutputFileExists` field **removed** from asset
- `themeFolder` field **removed** from asset
- `processAndWrite` renamed to `processAsset`
- `EnsureOutputDirectoryExists()` only called in `DiskMode`
- Default mode: `MemoryMode` (no disk writes, manual mode switch only)

## Active Features (Implementation Order)

| # | Feature | Document |
|---|---------|----------|
| 1 | **Config Simplification** | [FEATURE_CONFIG_SIMPLIFICATION.md](FEATURE_CONFIG_SIMPLIFICATION.md) |
| 2 | Asset Caching | [FEATURE_ASSET_CACHING.md](FEATURE_ASSET_CACHING.md) |
| 3 | HTTP Routes & Work Modes | [FEATURE_HTTP_ROUTES_WORK_MODES.md](FEATURE_HTTP_ROUTES_WORK_MODES.md) |
| 4 | Template System Refactor | [FEATURE_TEMPLATE_REFACTOR.md](FEATURE_TEMPLATE_REFACTOR.md) |
| 5 | TUI Integration | [FEATURE_TUI_HANDLER.md](FEATURE_TUI_HANDLER.md) |
| 6 | SSR Support | [SSR_IMPLEMENTATION_DETAILS.md](SSR_IMPLEMENTATION_DETAILS.md) |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| ThemeFolder | Remove | Components registered via NewFileEvent |
| OutputDir | `string` type | Doesn't change at runtime |
| WriteOnDisk | Remove → WorkMode | Clean break |
| Mode switch | Manual only | No auto-switch, predictable |
| processAndWrite | Rename to `processAsset` | Conditional disk write |
| EnsureOutputDir | DiskMode only | Don't create dirs in MemoryMode |
| NotifyIfOutputFilesExist | Remove | Irrelevant with MemoryMode |
| notifyMeIfOutputFileExists | Remove | No longer needed |
| filePath in NewFileEvent | Source path | OutputDir is for output only |
| Files in OutputDir at startup | Ignore | Must register via NewFileEvent |
| Minifier access | Pass as parameter | No state duplication |
| Concurrency | Test first | Add sync if needed |
| UnobservedFiles | style.css, script.js, icons.svg only | index.html, favicon.svg user-editable |
| Mode naming | MEMORY/DISK | Clear terminology |
| Cache regeneration | Eager (in processAsset) | Avoid HTTP latency |
| Templates | Keep embed.FS | Easier to edit |
| faviconSvgHandler | Keep | Consistent minification |

## Archived/Superseded

| Document | Notes |
|----------|-------|
| PROMPT_REFACTOR.md | Merged into active features |
| PROMPT_REFACTOR_V2.md | Merged into active features |
| PROMPT_REFACTOR_OBSERVATION_AND_SSR.md | Split into active features |
