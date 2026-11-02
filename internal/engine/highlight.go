package engine

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// htmlCode colours YAML during generation. HTML stays self-contained and the
// browser copies the original source from textContent, including CRLF and tabs.
func htmlCode(language, source string) string {
	if language != "yaml" || source == "" {
		return escapedCode(source)
	}
	iterator, err := lexers.Get("yaml").Tokenise(&chroma.TokeniseOptions{State: "root"}, source)
	if err != nil {
		return escapedCode(source)
	}
	var output, original strings.Builder
	for token := iterator(); token != chroma.EOF; token = iterator() {
		original.WriteString(token.Value)
		value := escapedCode(token.Value)
		if class := yamlTokenClass(token.Type); class != "" {
			fmt.Fprintf(&output, `<span class="syntax-%s">%s</span>`, class, value)
		} else {
			output.WriteString(value)
		}
	}
	// A lexer must never alter an example. Keep safe, lossless plain text if an
	// upstream lexer encounters input it cannot round-trip.
	if original.String() != source {
		return escapedCode(source)
	}
	return output.String()
}

func yamlTokenClass(token chroma.TokenType) string {
	switch {
	case token == chroma.NameTag:
		return "key"
	case token == chroma.CommentPreproc:
		return "meta"
	case token.InCategory(chroma.Comment):
		return "comment"
	case token.InCategory(chroma.Keyword):
		return "constant"
	case token.InSubCategory(chroma.LiteralNumber):
		return "number"
	case token.InSubCategory(chroma.LiteralString), token == chroma.Literal, token == chroma.LiteralDate:
		return "string"
	default:
		return ""
	}
}
