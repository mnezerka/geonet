package utils

// https://wondernote.org/color-palettes-for-web-digital-blog-graphic-design-with-hexadecimal-codes/
func GetDarkPastelColor(i int) string {
	colors := []string{
		"#0065a2",
		"#ff828b",
		"#ffa23a",
		"#c05780",
		"#00a5e3",
	}
	if i >= 0 && i < len(colors) {
		return colors[i]
	}
	return ""
}
