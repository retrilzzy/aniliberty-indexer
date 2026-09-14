package torznab

import (
	"html"
	"strings"
)

// Escapes special XML characters
func escapeXML(str string) string {
	return html.EscapeString(str)
}

// Maps a quality string to a numeric resolution for category mapping
func qualityToResolution(quality string) string {
	if quality == "" {
		return ""
	}
	q := strings.ToLower(quality)
	if strings.Contains(q, "4k") || strings.Contains(q, "2160") {
		return "2160"
	}
	if strings.Contains(q, "2k") {
		return "1440"
	}
	if strings.Contains(q, "1080") {
		return "1080"
	}
	if strings.Contains(q, "720") {
		return "720"
	}
	if strings.Contains(q, "576") {
		return "576"
	}
	if strings.Contains(q, "480") {
		return "480"
	}
	if strings.Contains(q, "360") {
		return "360"
	}
	return ""
}

// Normalizes a codec string to x264 or x265
func cleanCodec(codec string) string {
	c := strings.ToLower(codec)
	if strings.Contains(c, "x265") || strings.Contains(c, "hevc") {
		return "x265"
	}
	if strings.Contains(c, "x264") || strings.Contains(c, "h264") || strings.Contains(c, "avc") {
		return "x264"
	}
	return codec
}

// Pads a string with leading zeros
func padZero(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

// Extracts episode range or single episode from the last bracket pair in the label
func extractEpisodes(label string) string {
	lastOpen := strings.LastIndex(label, "[")
	lastClose := strings.LastIndex(label, "]")
	if lastOpen == -1 || lastClose == -1 || lastOpen > lastClose {
		return ""
	}

	content := label[lastOpen+1 : lastClose]
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return ""
	}

	epRange := parts[0]
	cleanEp := strings.TrimPrefix(strings.TrimPrefix(epRange, "E"), "e")
	if len(cleanEp) > 0 && cleanEp[0] >= '0' && cleanEp[0] <= '9' {
		if strings.Contains(cleanEp, "-") {
			eps := strings.SplitN(cleanEp, "-", 2)
			if len(eps) == 2 {
				e0 := strings.TrimPrefix(strings.TrimPrefix(eps[0], "E"), "e")
				e1 := strings.TrimPrefix(strings.TrimPrefix(eps[1], "E"), "e")
				return "E" + padZero(e0) + "-E" + padZero(e1)
			}
		} else {
			return "E" + padZero(cleanEp)
		}
	}
	return ""
}

// Cleans up empty or malformed bracket sequences and spaces
func cleanFinalTitle(s string) string {
	s = strings.Join(strings.Fields(s), " ")

	for i := 0; i < 2; i++ {
		oldLen := len(s)
		s = strings.ReplaceAll(s, "(, ", "(")
		s = strings.ReplaceAll(s, "[, ", "[")
		s = strings.ReplaceAll(s, "{, ", "{")
		s = strings.ReplaceAll(s, ", )", ")")
		s = strings.ReplaceAll(s, ", ]", "]")
		s = strings.ReplaceAll(s, ", }", "}")
		s = strings.ReplaceAll(s, "( )", "")
		s = strings.ReplaceAll(s, "[ ]", "")
		s = strings.ReplaceAll(s, "{ }", "")
		s = strings.ReplaceAll(s, "()", "")
		s = strings.ReplaceAll(s, "[]", "")
		s = strings.ReplaceAll(s, "{}", "")
		s = strings.ReplaceAll(s, "[ ", "[")
		s = strings.ReplaceAll(s, " ]", "]")
		s = strings.ReplaceAll(s, "( ", "(")
		s = strings.ReplaceAll(s, " )", ")")
		s = strings.ReplaceAll(s, " ,", ",")

		s = strings.Join(strings.Fields(s), " ")
		if len(s) == oldLen {
			break
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, ",")
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimSuffix(s, ",")
	s = strings.TrimSuffix(s, "-")
	return strings.Join(strings.Fields(s), " ")
}

// Replaces placeholders in the template with corresponding values
func RenderTemplate(tpl string, getValue func(string) string) string {
	var sb strings.Builder
	inBrace := false
	var currentPlaceholder strings.Builder

	for _, char := range tpl {
		if char == '{' {
			inBrace = true
			currentPlaceholder.Reset()
		} else if char == '}' {
			inBrace = false
			sb.WriteString(getValue(currentPlaceholder.String()))
		} else {
			if inBrace {
				currentPlaceholder.WriteRune(char)
			} else {
				sb.WriteRune(char)
			}
		}
	}
	return cleanFinalTitle(sb.String())
}
