package assetmin

func NewSvgHandler(ac *AssetConfig) *asset {
	svgh := NewFileHandler(svgMainFileName, "image/svg+xml", ac, nil)

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
