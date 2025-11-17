# PWA Support for AssetMin

This document outlines a plan to add Progressive Web App (PWA) support to the `assetmin` library. It explores different alternatives, their pros and cons, and provides a recommendation on the best approach.

## 1. Introduction to PWA

A Progressive Web App (PWA) is a web application that uses modern web capabilities to provide a user experience similar to that of a native app. Key features of a PWA include:

- **Reliability:** Load instantly and never show a "no internet connection" screen.
- **Fast:** Respond quickly to user interactions with smooth animations.
- **Engaging:** Feel like a natural app on the device, with an immersive user experience.

To achieve this, PWAs require a few key components:

- **Service Worker:** A script that runs in the background, separate from the web page, and handles network requests, caching, and push notifications.
- **Web App Manifest:** A JSON file that provides information about the application, such as its name, icons, and colors.
- **HTTPS:** PWAs must be served over a secure connection.

## 2. PWA Integration with `assetmin`

The `assetmin` library is responsible for bundling and minifying web assets. To add PWA support, we need to extend its functionality to generate and manage the PWA-specific files (`manifest.json` and `sw.js`).

The integration would involve the following steps:

1. **Generate `manifest.json`:** Create a new handler that generates the `manifest.json` file based on a user-provided configuration.
2. **Generate `sw.js`:** Create a service worker file. This can be a simple, pre-defined template or a more complex, configurable one.
3. **Inject PWA elements into `index.html`:** The `html.go` file needs to be modified to include:
    - A link to the `manifest.json` file in the `<head>` section.
    - A script to register the service worker in the `<body>` section.

## 3. Implementation Alternatives

There are two main alternatives for implementing PWA support in `assetmin`:

### Alternative A: Integrated PWA Handler

In this approach, the PWA logic would be integrated directly into the existing `assetmin` package.

- **Pros:**
    - **Simplicity:** All the logic would be in one place, making it easier to maintain.
    - **Tight Integration:** The PWA handler would have direct access to the `assetmin` core, allowing for a seamless integration.
- **Cons:**
    - **Increased Complexity:** Adding PWA logic to the core package could make it more complex and harder to understand.
    - **Less Flexibility:** It might be harder to reuse the PWA logic in other projects.

### Alternative B: Separate PWA Package

In this approach, the PWA logic would be implemented in a separate package (e.g., `assetminpwa`).

- **Pros:**
    - **Modularity:** The PWA logic would be decoupled from the core `assetmin` package, making it easier to maintain and test.
    - **Reusability:** The PWA package could be used in other projects, independently of `assetmin`.
- **Cons:**
    - **Increased Complexity:** It would require managing an additional package and its dependencies.
    - **Potential for Code Duplication:** Some code might need to be duplicated between the two packages.

## 4. Recommendation

After evaluating the pros and cons of each alternative, the recommended approach is to **create a separate PWA package** (`assetminpwa`). This will provide a more modular and reusable solution, while keeping the core `assetmin` package focused on its primary responsibility of asset bundling and minification.

The `assetminpwa` package would provide a PWA handler that can be used in conjunction with the `assetmin` library to add PWA support to any Go web application.

## 5. Next Steps

The next steps to implement PWA support are:

1. Create a new Go package named `assetminpwa`.
2. Implement a PWA handler in the new package that generates the `manifest.json` and `sw.js` files.
3. Modify the `html.go` file in the `assetmin` package to inject the PWA elements into the `index.html` file.
4. Add a new example to the `assetmin` documentation that demonstrates how to use the `assetminpwa` package to add PWA support to a web application.

## 6. Detailed Implementation Plan

This section provides a detailed overview of the files and methods that will be modified to implement PWA support.

### 6.1. New Package: `assetminpwa`

A new package `assetminpwa` will be created to house the PWA-specific logic. This package will contain two main components:

- **`manifest.go`:** This file will define a `Manifest` struct that represents the `manifest.json` file. It will also contain a function to generate the JSON output.
- **`serviceworker.go`:** This file will contain a function to generate a default `sw.js` file. The service worker will be pre-configured to cache the main assets (`main.js`, `style.css`, `index.html`).

### 6.2. Modifications to `assetmin`

The following files in the `assetmin` package will be modified:

#### `assetmin.go`

- **`AssetConfig` struct:** A new field `PWAConfig *assetminpwa.Config` will be added to this struct. This will allow users to enable and configure PWA support.
- **`NewAssetMin` function:** This function will be updated to check for the `PWAConfig`. If it's not nil, it will create new asset handlers for `manifest.json` and `sw.js`.

#### `html.go`

- **`NewHtmlHandler` function:** This is the core of the integration. The function will be modified to:
    1. Check if PWA is enabled in the `AssetConfig`.
    2. If enabled, it will add the following to the `contentOpen` slice, which corresponds to the `<head>` of the HTML document:
        ```html
        <link rel="manifest" href="manifest.json">
        <meta name="theme-color" content="black">
        ```
    3. It will also append a script to the `contentClose` slice to register the service worker just before the closing `</body>` tag:
        ```html
        <script>
            if ('serviceWorker' in navigator) {
                window.addEventListener('load', () => {
                    navigator.serviceWorker.register('/sw.js').then(registration => {
                        console.log('SW registered: ', registration);
                    }).catch(registrationError => {
                        console.log('SW registration failed: ', registrationError);
                    });
                });
            }
        </script>
        ```

### 6.3. Architectural Alignment

This approach aligns with the existing architecture by:

- **Leveraging the existing asset handling mechanism:** The `manifest.json` and `sw.js` files will be treated as regular assets, managed by the `asset` struct.
- **Maintaining separation of concerns:** The PWA logic is encapsulated in its own package, `assetminpwa`, keeping the core `assetmin` library clean.
- **Providing a consistent configuration experience:** Users will enable PWA support through the familiar `AssetConfig` struct.

This detailed plan ensures that the PWA implementation will be clean, maintainable, and well-integrated with the existing `assetmin` architecture.
