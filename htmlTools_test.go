package assetmin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseExistingHtmlContent(t *testing.T) {
	t.Run("with_placeholder", func(t *testing.T) {
		html := `<!doctype html>
<html>
<head>
    <title>Test</title>
</head>
<body>
    <header>Header</header>
    <!-- MODULES_PLACEHOLDER -->
    <footer>Footer</footer>
    <script src="app.js"></script>
</body>
</html>`

		open, close := parseExistingHtmlContent(html)

		assert.Contains(t, open, "<header>Header</header>")
		assert.Contains(t, close, "<footer>Footer</footer>")
		assert.Contains(t, close, "<script src=\"app.js\"></script>")
	})

	t.Run("with_main_tag", func(t *testing.T) {
		html := `<!doctype html>
<html>
<head>
    <title>Test</title>
</head>
<body>
    <header>Header</header>
    <main>
        <div>Content</div>
    </main>
    <footer>Footer</footer>
    <script src="app.js"></script>
</body>
</html>`

		open, close := parseExistingHtmlContent(html)

		assert.Contains(t, open, "<main>")
		assert.Contains(t, close, "</main>")
		assert.Contains(t, close, "<footer>Footer</footer>")
	})

	t.Run("with_script_tag", func(t *testing.T) {
		html := `<!doctype html>
<html>
<head>
    <title>Test</title>
</head>
<body>
    <header>Header</header>
    <div>Content</div>
    <script src="app.js"></script>
</body>
</html>`

		open, close := parseExistingHtmlContent(html)

		assert.Contains(t, open, "<div>Content</div>")
		assert.Contains(t, close, "<script src=\"app.js\"></script>")
		assert.NotContains(t, open, "<script")
	})

	t.Run("only_body_tag", func(t *testing.T) {
		html := `<!doctype html>
<html>
<head>
    <title>Test</title>
</head>
<body>
    <header>Header</header>
    <div>Content</div>
</body>
</html>`

		open, close := parseExistingHtmlContent(html)

		assert.Contains(t, open, "<div>Content</div>")
		assert.Contains(t, close, "</body>")
		assert.Contains(t, close, "</html>")
	})

	t.Run("complex_body_structure", func(t *testing.T) {
		html := `<!DOCTYPE html>
<html lang="es">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link rel="StyleSheet" href="style.css">
	<title>App Title</title>
</head>
<body>
	<nav class="menu-container">
		<ul class="navbar-container">
			<li class="navbar-item">
				<a href="#" class="navbar-link">Home</a>
			</li>
		</ul>
	</nav>
	<header>
		<div id="USER_NAME"><a href="#login">Username</a></div>
		<h2 id="USER_AREA">User Area</h2>
	</header>
	<div id="user-mobile-messages">
		<h4 class="err">Message</h4>
	</div>

	{{.Modules}}

	<script src="app.js"></script>
</body>
</html>`

		open, close := parseExistingHtmlContent(html)

		// Verificar que el contenido se dividió correctamente en el marcador {{.Modules}}
		assert.Contains(t, open, "<div id=\"user-mobile-messages\">")
		assert.Contains(t, open, "<h4 class=\"err\">Message</h4>")
		assert.Contains(t, open, `<div id="user-mobile-messages">
		<h4 class="err">Message</h4>
	</div>`)
		assert.Contains(t, close, "<script src=\"app.js\"></script>")

		// Verificar que la división fue exacta alrededor del marcador
		assert.True(t, strings.HasSuffix(strings.TrimSpace(open), "</div>"))
		assert.True(t, strings.HasPrefix(strings.TrimSpace(close), "<script"))
	})
}
