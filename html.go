package assetmin

type htmlHandler struct {
	*fileHandler
}

func NewHtmlHandler(ac *AssetConfig) *fileHandler {
	h := NewFileHandler(htmlMainFileName, "text/html", ac, nil)

	hh := &htmlHandler{
		fileHandler: h,
	}
	// Configurar el procesador personalizado para manejar los módulos HTML
	h.customFileProcessor = hh.processModuleFile

	// Agregar el marcador de inicio de los módulos HTML
	h.contentOpen = append(h.contentOpen, &contentFile{
		path: "index-open.html",
		content: []byte(`<!DOCTYPE html>
<html lang="es">

<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no, viewport-fit=cover">
	<meta name="viewport">
	<meta name="mobile-web-app-capable" content="yes">
	<meta name="apple-mobile-web-app-capable" content="yes">
	<!-- posibles valores de contenido: predeterminado, negro o negro translúcido -->
	<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
	<link rel="icon" type="image/png" href="static/favicon.png">
	<link rel="StyleSheet" href="{{.StyleSheet}}">
	<meta name="JsonBootActions" content="{{.JsonBootActions}}">
	<title>{{.AppName}}-ver.{{.AppVersion}}</title>
	
</head>

<body>

	<nav class="menu-container">
		<ul class="navbar-container">
			<!-- <li class="navbar-item">
				<a href="#" class="navbar-link" name="home">
					<svg aria-hidden="true" focusable="false" class="fa-primary">
						<use xlink:href="#icon-home" />
					</svg>
					<span class="link-text">Home</span>
				</a>
			</li> -->
			{{.Menu}}
		</ul>
	</nav>
	<header>
		<div id="USER_NAME"><a href="#login" name="login" title="Cerrar Sesion">{{.UserName}}</a></div>
		<div id="user-desktop-messages">
			<H4 class="err">{{.Message}}</H4>
		</div>
		<h2 id="USER_AREA">{{.UserArea}}</h2>
	</header>

	<div id="user-mobile-messages">
		<H4 class="err">{{.Message}}</H4>
	</div>
	<div id="modal-window" onclick="modalHandler(event)">
		<H4>MODO DE VISUALIZACIÓN INCOMPATIBLE</H4>
	</div>

	<!-- Inicio de módulos HTML -->`),
	})

	// Agregar el marcador de cierre de los módulos HTML
	h.contentClose = append(h.contentClose, &contentFile{
		path: "index-close.html",
		content: []byte(`	<!-- Fin de módulos HTML -->
	<script src="{{.Script}}"></script>
</body>

</html>`),
	})

	return h
}

func (h *htmlHandler) processModuleFile(event string, f *contentFile) error {

	return nil
}
