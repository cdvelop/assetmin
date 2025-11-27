# Refactoring Plan V2 for AssetMin and GoLite

## 1. AssetMin Refactor
**Goal**: Replace direct content update with a notification-based rebuild mechanism.

- [ ] **Remove `UpdateAssetContent`**: This method is no longer needed as we won't push content directly.
- [ ] **Add `RefreshAsset(extension string)`**:
    - Accepts an extension (e.g., ".js", ".css").
    - Finds the corresponding handler (e.g., `mainJsHandler` for ".js").
    - Triggers a rebuild/write of that asset.
    - **Crucial**: Ensure that the rebuild process re-fetches the `GetRuntimeInitializerJS` content (which will be provided by `golite` -> `tinywasm`).
    - This method should NOT trigger a browser reload; it just updates the file on disk.

## 2. GoLite Refactor (Blocked on AssetMin)
**Goal**: Correctly wire `tinywasm` embedding and browser reload.

- [ ] **Configure `GetRuntimeInitializerJS`**:
    - In `AddSectionBUILD`, set `AssetConfig.GetRuntimeInitializerJS` to call `h.wasmHandler.JavascriptForInitializing()`.
    - This ensures `assetmin` pulls the latest `wasm_exec.js` content when building `script.js`.
- [ ] **Update `OnWasmExecChange`**:
    - In the callback:
        1.  Call `h.assetsHandler.RefreshAsset(".js")` to force `assetmin` to rebuild `script.js` (pulling the new WASM JS).
        2.  Call `h.browser.Reload()` to refresh the browser.
- [ ] **Update Test**:
    - Verify `script.js` contains the WASM JS.
    - Verify browser reload happens.
