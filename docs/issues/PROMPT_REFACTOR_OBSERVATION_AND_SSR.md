# SUPERSEDED - See Active Feature Documents

> **Note**: This document has been superseded and split into focused feature documents.

## Migrated To:
- File Observation Logic → [FEATURE_HTTP_ROUTES_WORK_MODES.md](FEATURE_HTTP_ROUTES_WORK_MODES.md)
- SSR Support → [SSR_IMPLEMENTATION_DETAILS.md](SSR_IMPLEMENTATION_DETAILS.md)

## Original Context (Archived)
`assetmin` instructs the watcher to ignore output files via `UnobservedFiles`. The `isOutputPath` check in `NewFileEvent` prevents infinite loops.

**Current behavior preserved**: `isOutputPath()` already handles loop prevention correctly. No changes needed to observation logic.
