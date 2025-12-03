package assetmin

func NewSvgHandler(ac *Config, outputName string) *asset {
	svgh := newAssetFile(outputName, "image/svg+xml", ac, nil)

	// Add the open tags to contentOpen
	svgh.contentOpen = append(svgh.contentOpen, &contentFile{
		path: "sprite-open.svg",
		content: []byte(`<svg class="sprite-icons" xmlns="http://www.w3.org/2000/svg" role="img" aria-hidden="true" focusable="false">
		<defs>`),
	})

	// Add the closing tags to contentClose
	svgh.contentClose = append(svgh.contentClose, &contentFile{
		path: "sprite-close.svg",
		content: []byte(`		</defs>
	</svg>`),
	})

	return svgh
}

// NewFaviconSvgHandler creates a handler for favicon.svg that simply minifies and copies the file
// without sprite wrapping. This handler processes standalone SVG files like favicon.svg
func NewFaviconSvgHandler(ac *Config, outputName string) *asset {
	return newAssetFile(outputName, "image/svg+xml", ac, nil)
}
