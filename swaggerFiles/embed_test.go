package swaggerFiles

import "testing"

func TestEmbeddedFiles(t *testing.T) {
	for _, name := range []string{"/swagger-ui-bundle.js", "/swagger-ui-standalone-preset.js", "/swagger-ui.css", "/oauth2-redirect.html", "/oauth2-redirect.js"} {
		if b, err := ReadFile(name); err != nil || len(b) == 0 {
			t.Errorf("%s: %v", name, err)
		}
	}
}
