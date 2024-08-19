package markup

import "strings"

var replacer = strings.NewReplacer(
	"-", "\\-",
	"_", "\\_",
	"*", "\\*",
	"[", "\\[",
	"]", "\\]",
	"(", "\\(",
	")", "\\)",
	"{", "\\{",
	"}", "\\}",
	"~", "\\~",
	"`", "\\`",
	">", "\\>",
	"#", "\\#",
	"+", "\\+",
	"=", "\\=",
	"|", "\\|",
	"!", "\\!",
	".", "\\.",
)

func MarkdownEscape(text string) string {
	return replacer.Replace(text)
}
