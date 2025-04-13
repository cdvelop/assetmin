package assetmin

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type htmlHandler struct {
	*fileHandler

	appName    string // ej: "myapp"
	appVersion string // ej: "1.0.0"

	index *html.Node

	// files []*file
	buf bytes.Buffer

	head            *html.Node
	deleteHeadNodes []*html.Node
	body            *html.Node
	deleteBodyNodes []*html.Node
}

func NewHtmlHandler(ac *AssetConfig) *htmlHandler {
	h := &htmlHandler{
		fileHandler: NewFileHandler(htmlMainFileName, "text/html", ac),
		appName:     "myapp",
		appVersion:  "1.0.0",
	}
	return h

}

type file struct {
	index   int
	path    string
	content []byte
}

type module struct {
	filePath string
	content  []byte
}

func (h *htmlHandler) Title() string {
	return h.appName + "-v" + h.appVersion
}

func createLinkNode(href string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: "link",
		Attr: []html.Attribute{
			{Key: "rel", Val: "stylesheet"},
			{Key: "href", Val: href},
		},
	}
}

func createScriptNode(src string) *html.Node {
	return &html.Node{
		Type: html.ElementNode,
		Data: "script",
		Attr: []html.Attribute{
			{Key: "src", Val: src},
		},
	}
}

func (h *htmlHandler) modifyAttributes(n *html.Node) {
	// fmt.Printf("N DATA: %s => %s\n", n.Data, n.Attr)
	// Verificamos si el nodo es un elemento HTML
	if n.Type == html.ElementNode {
		switch n.Data {

		case "title":
			// fmt.Println("**TITLE NODE FOUND:", n.Data, n.Attr)
			// Agregar el nuevo título como un nodo de texto
			n.FirstChild.Data = h.Title()

		case "body":
			// Marcamos que hemos llegado al final del head y empezamos el body
			// fmt.Println("**END HEAD BODY START:", n.Data)
			// Añadimos el script principal al final del body
			n.AppendChild(createScriptNode(jsMainFileName))
		case "head":
			h.head = n
		case "link":
			var stylesheetType, externalAsset bool
			// fmt.Println("LINK NODE FOUND:", n.Data, "n.Attr:", n.Attr)

			// Analizamos los atributos del nodo link
			for i, a := range n.Attr {
				if a.Key == "rel" && a.Val == "stylesheet" {
					stylesheetType = true
				}
				if strings.HasPrefix(a.Val, "http://") || strings.HasPrefix(a.Val, "https://") {
					externalAsset = true
				}
				// fmt.Printf("key: %s value: %s\n", a.Key, a.Val)

				// Si es un enlace a un favicon, lo modificamos para que apunte a la carpeta de assets
				if strings.HasPrefix(a.Val, "favicon") {
					// fmt.Println("=> favicon found:", a.Val)
					n.Attr[i].Val = a.Val
				}
			}
			// fmt.Println("stylesheetType:", stylesheetType, "externalAsset:", externalAsset, n.Attr)

			// Si es una hoja de estilos interna, la eliminamos
			if h.head != nil && stylesheetType && !externalAsset {
				// fmt.Println("agregando nodo a eliminar:", n.Attr)
				h.deleteHeadNodes = append(h.deleteHeadNodes, n)
			}
		}
	}

	// Procesamos recursivamente todos los hijos del nodo actual
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		h.modifyAttributes(c)
	}
}

// event: create, remove, write, rename
func (h *htmlHandler) UpdateContent(filePath, event string, f *assetFile) error {

	var e = "processContent Html " + event
	if len(f.content) == 0 {
		return nil
	}

	fmt.Println("Compilando HTML..." + filePath)

	var err error
	if event == "create" || event == "update" {
		h.head = nil
		h.index, err = html.Parse(bytes.NewReader(f.content))
		if err != nil {
			return errors.New(e + err.Error())
		}

		// Modificar atributos
		h.modifyAttributes(h.index)

		// delete head nodes
		if h.head != nil {
			for _, n := range h.deleteHeadNodes {
				// fmt.Println("=> eliminando nodo:", n.Attr)
				h.head.RemoveChild(n)
			}

			// Añadimos el enlace al CSS principal al final del head
			h.head.AppendChild(createLinkNode(cssMainFileName))
		}

	}

	return nil
}
