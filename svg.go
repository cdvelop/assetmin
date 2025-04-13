package assetmin

import (
	"bytes"
	"encoding/xml"
	"errors"
	"os"
	"strings"
)

type svgHandler struct {
	*fileHandler
	sprite *sprite
	icons  map[string]icon
}

func NewSvgHandler(WebFilesFolder string) *svgHandler {
	svgh := &svgHandler{
		fileHandler: NewFileHandler(svgMainFileName, "image/svg+xml", nil, WebFilesFolder),
		sprite:      &sprite{},
		icons:       make(map[string]icon),
	}

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
func (h *svgHandler) UpdateContent(filePath, event string, f *assetFile) error {

	contentStr := string(f.content)
	if strings.Contains(contentStr, "<svg") {
		return h.processSprite(contentStr, event)
	} else if strings.Contains(contentStr, "<symbol") {
		return h.processSymbol(contentStr, event)
	}
	return errors.New("contenido no reconocido")

}

func (h *svgHandler) processSprite(content, event string) error {

	if event == "remove" {
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

func (h *svgHandler) processSymbol(content, event string) error {
	var sym symbol
	if err := xml.Unmarshal([]byte(content), &sym); err != nil {
		return err
	}

	if event == "remove" {
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

// WriteContent implements the AssetHandler interface for svgHandler
func (h *svgHandler) WriteContent() error {
	// Implement SVG sprite generation and writing
	// This is a placeholder - you'll need to implement the actual SVG sprite generation
	var buf bytes.Buffer

	// Create the SVG sprite XML structure
	h.sprite.Xmlns = "http://www.w3.org/2000/svg"
	h.sprite.Role = "img"
	h.sprite.Hidden = "true"
	h.sprite.Focus = "false"
	h.sprite.Class = "svg-sprite"

	// Add all symbols to the sprite
	h.sprite.Defs.Symbols = make([]symbol, 0, len(h.icons))
	for _, icon := range h.icons {
		h.sprite.Defs.Symbols = append(h.sprite.Defs.Symbols, icon.symbol)
	}

	// Marshal the sprite to XML
	xmlData, err := xml.MarshalIndent(h.sprite, "", "  ")
	if err != nil {
		return err
	}

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.Write(xmlData)

	return FileWrite(h.outputPath, buf)
}
