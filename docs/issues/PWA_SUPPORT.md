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
