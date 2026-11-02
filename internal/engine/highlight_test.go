package engine

import (
	"html"
	"regexp"
	"strings"
	"testing"
)

func TestHTMLYAMLHighlighting(t *testing.T) {
	source := "# platform settings\napiVersion: example.org/v1\nmetadata:\n  name: team-atlas\nspec:\n  replicas: 2\n  enabled: true\n"
	w := &writer{config: Config{Formatter: "html"}}
	w.code("yaml", source)
	output := w.buf.String()
	for _, class := range []string{"syntax-key", "syntax-string", "syntax-number", "syntax-constant", "syntax-comment"} {
		if !strings.Contains(output, `class="`+class+`"`) {
			t.Errorf("YAML has no %s tokens", class)
		}
	}
}

func TestHTMLCodePreservesTextAndEscapesMarkup(t *testing.T) {
	for _, source := range []string{
		"name: atlas\n",
		"unicode: λ 🐈\r\n\t# preserve CRLF and tabs\r\n",
		"script: |\n  {{ .spec.region }}\n  echo 'literal \\n'\n  </script><img src=x onerror=alert(1)>\n",
		"defaults: &defaults {enabled: true, count: 0, unset: null}\ncopy: *defaults\nquoted: 'x: # not a comment' # comment\n",
		"unterminated: [\"still safe",
		"",
	} {
		for _, language := range []string{"yaml", "text", "gotemplate"} {
			t.Run(language+"/"+source, func(t *testing.T) {
				w := &writer{config: Config{Formatter: "html"}}
				w.code(language, source)
				output := w.buf.String()
				code := regexp.MustCompile(`<pre><code[^>]*>([\s\S]*?)</code></pre>`).FindStringSubmatch(output)
				if len(code) != 2 {
					t.Fatal("code element missing")
				}
				text := html.UnescapeString(regexp.MustCompile(`<span class="syntax-[a-z]+">|</span>`).ReplaceAllString(code[1], ""))
				if text != source {
					t.Fatalf("code text changed: got %q, want %q", text, source)
				}
				if strings.Contains(output, "<img") || strings.Contains(output, "</script>") {
					t.Fatal("source markup was not escaped")
				}
				if language != "yaml" && strings.Contains(code[1], "<span") {
					t.Fatal("highlighting changed a non-YAML language")
				}
			})
		}
	}
}
