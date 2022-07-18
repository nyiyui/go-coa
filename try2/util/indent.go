package util

import "strings"

const IndentString = "\t"
const MultilineThreshold = 40

func Indent(s string) string {
	onlyLf := strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
	return IndentString + strings.ReplaceAll(onlyLf, "\n", "\n"+IndentString)
}
