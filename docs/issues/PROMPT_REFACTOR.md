# Refactoring Plan for AssetMin and GoLite

## 1. AssetMin Refactor
**Goal**: Optimize template generation and integrate with TinyWasm updates.

- [ ] **Direct Output**: Modify `assetmin` to generate default templates directly into the output directory (minified), avoiding intermediate files in the source directory.
    - Currently, `assetmin` creates basic HTML files if they don't exist.
    - Change this to write the *minified* version directly to the output folder.
- [ ] **Public Update Method**: Expose a public method in `AssetMin` to receive notifications.
    - This method will be called when `tinywasm` updates `wasm_exec.js` or other assets.
    - It should trigger a re-minification/update of the output files.

## 2. GoLite Refactor (Next Step)
**Goal**: Wire everything together.

- [ ] **Integration**: Configure `golite` to connect `tinywasm`'s `OnWasmExecChange` to `assetmin`'s new update method.
- [ ] **Direct Injection**: Pass `wasm_exec.js` content directly if possible.
