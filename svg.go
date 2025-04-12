package assetmin

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type svgHandler struct {
	fileOutputName string // eg:  "sprite.svg"
	outputPath     string // eg:  "public/sprite.svg"
	mediatype      string // eg:  "image/svg+xml"
	sprite         *sprite
	icons          map[string]icon
}

func NewSvgHandler(WebFilesFolder string) *svgHandler {
	svgh := &svgHandler{
		fileOutputName: "sprite.svg",
		mediatype:      "image/svg+xml",
		sprite:         &sprite{},
		icons:          make(map[string]icon),
	}

	svgh.outputPath = filepath.Join(WebFilesFolder, svgh.fileOutputName)

	return svgh
}

type icon struct {
	content string
	symbol  symbol
}

type sprite struct {
	XMLName xml.Name `xml:"svg"`
	Xmlns   string   `xml:"xmlns,attr"`
	Role    string   `xml:"role,attr"`
	Hidden  string   `xml:"aria-hidden,attr"`
	Focus   string   `xml:"focusable,attr"`
	Class   string   `xml:"class,attr"`
	Defs    defs     `xml:"defs"`
}

type defs struct {
	Symbols []symbol `xml:"symbol"`
}

type symbol struct {
	XMLName xml.Name `xml:"symbol"`
	ID      string   `xml:"id,attr"`
	ViewBox string   `xml:"viewBox,attr"`
	Path    path     `xml:"path"`
}

type path struct {
	Fill string `xml:"fill,attr"`
	D    string `xml:"d,attr"`
}

// event: create, remove, write, rename
func (h *svgHandler) processContent(content []byte, action string) error {
	contentStr := string(content)
	if strings.Contains(contentStr, "<svg") {
		return h.processSprite(contentStr, action)
	} else if strings.Contains(contentStr, "<symbol") {
		return h.processSymbol(contentStr, action)
	}
	return errors.New("contenido no reconocido")
}

func (h *svgHandler) processSprite(content, action string) error {

	if action == "delete" {
		return os.Remove(h.outputPath)
	}

	// var sp sprite
	if err := xml.Unmarshal([]byte(content), h.sprite); err != nil {
		return err
	}
	// for _, symbol := range h.sprite.Defs.Symbols {
	// 	h.symbols[symbol.ID] = symbolToString(symbol)
	// }
	return nil
}

func (h *svgHandler) processSymbol(content, action string) error {
	var sym symbol
	if err := xml.Unmarshal([]byte(content), &sym); err != nil {
		return err
	}

	if action == "delete" {
		delete(h.icons, sym.ID)
		// h.writeToFile()
		return nil
	}

	h.icons[sym.ID] = icon{
		content: content,
		symbol:  sym,
	}

	// fmt.Println("simbolo", sym.ID, sym.ViewBox)

	return nil
}
