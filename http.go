package assetmin

import (
	"net/http"
)

// RegisterRoutes registers the HTTP handlers for all assets.
func (c *AssetMin) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(c.indexHtmlHandler.URLPath(), c.serveAsset(c.indexHtmlHandler))
	mux.HandleFunc(c.mainStyleCssHandler.URLPath(), c.serveAsset(c.mainStyleCssHandler))
	mux.HandleFunc(c.mainJsHandler.URLPath(), c.serveAsset(c.mainJsHandler))
	mux.HandleFunc(c.spriteSvgHandler.URLPath(), c.serveAsset(c.spriteSvgHandler))
	mux.HandleFunc(c.faviconSvgHandler.URLPath(), c.serveAsset(c.faviconSvgHandler))
}

func (c *AssetMin) serveAsset(asset *asset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := asset.GetMinifiedContent(c.min)
		if err != nil {
			http.Error(w, "Error getting minified content", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", asset.mediatype)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_, _ = w.Write(content)
	}
}
