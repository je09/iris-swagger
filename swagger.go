package swagger

import (
	"html/template"
	"os"
	"strings"

	"golang.org/x/net/webdav"

	"github.com/kataras/iris/v12"
	"github.com/swaggo/swag"
)

// Configurator represents a configuration setter.
type Configurator interface {
	Configure(*Config)
}

// ConfiguratorFunc implements the Configuration as a function type.
type ConfiguratorFunc func(*Config)

// Configure calls itself and modifies the default config.
func (fn ConfiguratorFunc) Configure(config *Config) {
	fn(config)
}

// Config stores swagger configuration variables.
type Config struct {
	// The URL pointing to API definition (normally swagger.json or swagger.yaml).
	// Default is `doc.json`.
	URL string
	// The prefix url which this swagger ui is registered on.
	// Defaults to "/swagger". It can be a "." too.
	Prefix       string
	DeepLinking  bool
	DocExpansion string
	DomID        string
}

// Configure completes the Configurator interface.
// It allows to pass a Config as it is and override any option.
func (c Config) Configure(config *Config) {
	config.URL = c.URL
	config.Prefix = c.Prefix
	config.DeepLinking = c.DeepLinking
	config.DocExpansion = c.DocExpansion
	config.DomID = c.DomID
}

// URL presents the URL pointing to API definition (normally swagger.json or swagger.yaml).
func URL(url string) ConfiguratorFunc {
	return func(c *Config) {
		c.URL = url
	}
}

// Prefix presents the URL prefix of this swagger UI (normally "/swagger" or ".").
func Prefix(prefix string) ConfiguratorFunc {
	return func(c *Config) {
		c.Prefix = prefix
	}
}

// DocExpansion list, full, none.
func DocExpansion(docExpansion string) ConfiguratorFunc {
	return func(c *Config) {
		c.DocExpansion = docExpansion
	}
}

// DomID #swagger-ui.
func DomID(domID string) ConfiguratorFunc {
	return func(c *Config) {
		c.DomID = domID
	}
}

// DeepLinking set the swagger deeplinking configuration.
func DeepLinking(deepLinking bool) ConfiguratorFunc {
	return func(c *Config) {
		c.DeepLinking = deepLinking
	}
}

// Handler wraps the webdav http handler into an Iris Handler one.
//
// Usage:
//
//	swaggerUI := swagger.Handler(swaggerFiles.Handler,
//	 swagger.URL("http://localhost:8080/swagger/doc.json"), // The url pointing to API definition))
//	 swagger.DeepLinking(true),
//	 swagger.Prefix("/swagger"),
//	)
//	app.Get("/swagger", swaggerUI)
//	app.Get("/swagger/{any:path}", swaggerUI)
//
// OR
//
//	swaggerUI := swagger.Handler(swaggerFiles.Handler, swagger.Config{
//	 URL: ...,
//	 Prefix: ...,
//	 DeepLinking: ...,
//	 DocExpansion: ...,
//	 DomID: ...,
//	}
func Handler(h *webdav.Handler, configurators ...Configurator) iris.Handler {
	config := &Config{
		URL:          "doc.json",
		DeepLinking:  true,
		DocExpansion: "list",
		DomID:        "#swagger-ui",
		Prefix:       "/swagger",
	}

	for _, c := range configurators {
		c.Configure(config)
	}

	if prefix := config.Prefix; prefix != "" && prefix != "." {
		h.Prefix = prefix
	} else {
		config.Prefix = "." // relative files and don't touch the webdav one's (index swagger will not work without index.html).
	}

	handler := func(ctx iris.Context) {
		path := strings.TrimPrefix(ctx.Path(), config.Prefix)
		if sufIdx := strings.LastIndexByte(path, '.'); sufIdx > 0 {
			suffix := path[sufIdx:]
			switch suffix {
			case ".html":
				ctx.ContentType("text/html; charset=utf-8")
			case ".css":
				ctx.ContentType("text/css; charset=utf-8")
			case ".js":
				ctx.ContentType("application/javascript")
			case ".json":
				ctx.ContentType("application/json")
			}
		}

		switch path {
		case "", "/", "/index.html":
			ctx.ContentType("text/html; charset=utf-8")
			err := indexTmpl.Execute(ctx, config)
			if err != nil {
				ctx.Application().Logger().Errorf("swagger: %v", err)
				ctx.StopWithStatus(iris.StatusInternalServerError)
				return
			}
		case "/doc.json":
			doc, err := swag.ReadDoc()
			if err != nil {
				ctx.Application().Logger().Errorf("swagger: %v", err)
				ctx.StopWithStatus(iris.StatusInternalServerError)
				return
			}
			ctx.WriteString(doc)
		default:
			h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
		}
	}

	return handler
}

// DisablingHandler turns handler off
// if specified environment variable passed.
func DisablingHandler(h *webdav.Handler, envName string, configurators ...Configurator) iris.Handler {
	eFlag := os.Getenv(envName)
	if eFlag != "" {
		return func(ctx iris.Context) {
			// Simulate behavior when route unspecified and
			// return 404 HTTP code
			ctx.NotFound()
		}
	}

	return Handler(h, configurators...)
}

// Disable turns handler off if called.
func Disable(h *webdav.Handler, configurators ...Configurator) iris.Handler {
	return func(ctx iris.Context) {
		// Simulate behavior when route unspecified and
		// return 404 HTTP code
		ctx.NotFound()
	}
}

var indexTmpl = template.Must(template.New("swagger_index.html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="{{.Prefix}}/swagger-ui.css" />
  <link rel="icon" type="image/png" href="{{.Prefix}}/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="{{.Prefix}}/favicon-16x16.png" sizes="16x16" />
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
<div id="swagger-ui"></div>
<script src="{{.Prefix}}/swagger-ui-bundle.js" charset="UTF-8"></script>
<script src="{{.Prefix}}/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
<script>
window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "{{.URL}}",
    deepLinking: {{.DeepLinking}},
    docExpansion: "{{.DocExpansion}}",
    dom_id: "{{.DomID}}",
    validatorUrl: null,
    persistAuthorization: true,
    tryItOutEnabled: true,
    displayRequestDuration: true,
    requestSnippetsEnabled: true,
    filter: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout"
  })
}
</script>
</body>
</html>
`))
