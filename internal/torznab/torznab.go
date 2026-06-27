package torznab

import (
	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"

	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
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

var wordToNum = map[string]int{
	"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5,
	"sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
	"i": 1, "ii": 2, "iii": 3, "iv": 4, "v": 5,
	"vi": 6, "vii": 7, "viii": 8, "ix": 9, "x": 10,
}

func parseSeasonNum(s string) (int, bool) {
	s = strings.ToLower(strings.Trim(s, ".,:;[]()"))
	s = strings.TrimPrefix(s, "tv-")
	if val, ok := wordToNum[s]; ok {
		return val, true
	}
	s = strings.TrimRight(s, "stndrdth")
	if val, err := strconv.Atoi(s); err == nil && val > 0 && val < 20 {
		return val, true
	}
	return 0, false
}

func isSeasonWord(w string) bool {
	return strings.ToLower(strings.TrimRight(w, ".,:;")) == "season"
}

// Extracts season number from a title
func extractSeason(title string) int {
	title = strings.ToLower(title)
	words := strings.Fields(title)
	for i, w := range words {
		if isSeasonWord(w) {
			if i+1 < len(words) {
				if val, ok := parseSeasonNum(words[i+1]); ok {
					return val
				}
			}
			if i > 0 {
				if val, ok := parseSeasonNum(words[i-1]); ok {
					return val
				}
			}
		}
	}
	if len(words) > 0 {
		if val, ok := parseSeasonNum(words[len(words)-1]); ok {
			return val
		}
	}
	return 1
}

// Removes season markers and trailing digits from the Latin title
func cleanTitleSeason(title string) string {
	original := title
	words := strings.Fields(title)
	var newWords []string

	skipNext := false
	foundExplicitSeason := false
	for i := 0; i < len(words); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		wLower := strings.ToLower(words[i])

		if isSeasonWord(wLower) {
			if i+1 < len(words) {
				if _, ok := parseSeasonNum(strings.ToLower(words[i+1])); ok {
					skipNext = true
					foundExplicitSeason = true
					continue
				}
			}
			if i > 0 {
				if _, ok := parseSeasonNum(strings.ToLower(words[i-1])); ok {
					if len(newWords) > 0 {
						newWords = newWords[:len(newWords)-1]
					}
					foundExplicitSeason = true
					continue
				}
			}
		}
		newWords = append(newWords, words[i])
	}

	if !foundExplicitSeason && len(newWords) > 0 {
		lastWord := newWords[len(newWords)-1]
		if val, ok := parseSeasonNum(strings.ToLower(lastWord)); ok && val > 1 {
			if len(newWords) > 1 && strings.ToLower(newWords[len(newWords)-2]) == "act" {
			} else {
				newWords = newWords[:len(newWords)-1]
				if len(newWords) > 0 && strings.ToLower(newWords[len(newWords)-1]) == "part" {
					newWords = newWords[:len(newWords)-1]
				}
			}
		}
	}

	res := strings.Join(newWords, " ")
	res = strings.TrimRight(res, " -:/,;[({")
	res = strings.ReplaceAll(res, " .", ".")
	res = strings.ReplaceAll(res, " ,", ",")
	if res == "" {
		return original
	}
	return res
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
	if len(epRange) > 0 && epRange[0] >= '0' && epRange[0] <= '9' {
		if strings.Contains(epRange, "-") {
			eps := strings.SplitN(epRange, "-", 2)
			if len(eps) == 2 {
				return fmt.Sprintf("E%s-E%s", padZero(eps[0]), padZero(eps[1]))
			}
		} else {
			return fmt.Sprintf("E%s", padZero(epRange))
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

// Builds the torrent title for the Torznab feed
func buildTitle(release api.Release, torrent api.Torrent) string {
	ruName := release.Name.Main
	latinName := release.Name.Latin

	season := extractSeason(latinName)
	if season == 1 && extractSeason(ruName) > 1 {
		season = extractSeason(ruName)
	}

	cleanLatinName := cleanTitleSeason(latinName)

	yearStr := ""
	if release.Year != 0 {
		yearStr = strconv.Itoa(release.Year)
	}

	codecCleaned := ""
	if torrent.Codec.Value != "" {
		codecCleaned = cleanCodec(torrent.Codec.Value)
	}

	episodes := extractEpisodes(torrent.Label)

	seasonEpisodes := fmt.Sprintf("S%02d", season)
	if episodes != "" {
		seasonEpisodes += episodes
	}

	getValue := func(key string) string {
		switch key {
		case "title_latin_clean":
			return cleanLatinName
		case "season_episodes":
			return seasonEpisodes
		case "title_latin":
			return latinName
		case "title_ru":
			return ruName
		case "season":
			return strconv.Itoa(season)
		case "episodes":
			return episodes
		case "year":
			return yearStr
		case "type":
			return torrent.Type.Value
		case "quality":
			return torrent.Quality.Value
		case "codec":
			return codecCleaned
		default:
			return ""
		}
	}

	return RenderTemplate(config.TORRENT_TITLE_TEMPLATE, getValue)
}

// Writes a single <torznab:attr> element
func writeTorznabAttr(sb *strings.Builder, name string, value interface{}) {
	sb.WriteString("<torznab:attr name=\"")
	sb.WriteString(name)
	sb.WriteString("\" value=\"")
	switch v := value.(type) {
	case string:
		sb.WriteString(v)
	case int:
		sb.WriteString(strconv.Itoa(v))
	case int64:
		sb.WriteString(strconv.FormatInt(v, 10))
	default:
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	sb.WriteString("\" />\n")
}

// Returns Torznab category IDs for the given resolution string
func resolutionToCategories(res string) []int {
	cats := []int{5000, 5070}
	switch {
	case res == "1080" || res == "720" || res == "1440":
		cats = append(cats, 5040)
	case res == "2160":
		cats = append(cats, 5045)
	case res == "576" || res == "480" || res == "360":
		cats = append(cats, 5030)
	}
	return cats
}

// Builds the Torznab RSS feed XML from a list of items
func BuildTorznabXML(items []api.CombinedItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:torznab="http://torznab.com/schemas/2015/feed">
<channel>
<title>AniLiberty</title>
<link>%s</link>
`, escapeXML(config.ANILIBERTY_SITE)))

	for _, item := range items {
		release := item.Release
		torrent := item.Torrent

		title := buildTitle(release, torrent)
		hash := torrent.Hash

		releaseAlias := release.Alias
		if releaseAlias == "" && release.ID != 0 {
			releaseAlias = strconv.Itoa(release.ID)
		}

		detailsUrl := config.ANILIBERTY_SITE + "/anime/releases/release/" + releaseAlias
		downloadUrl := config.ANILIBERTY_API + "/anime/torrents/" + hash + "/file"
		magnetUrl := torrent.Magnet

		pubDate := time.Now().UTC().Format(time.RFC1123)
		if torrent.UpdatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, torrent.UpdatedAt); err == nil {
				pubDate = parsed.UTC().Format(time.RFC1123)
			}
		}

		var descParts []string
		if release.Name.Main != "" {
			descParts = append(descParts, release.Name.Main)
		}
		if release.Name.Latin != "" {
			descParts = append(descParts, release.Name.Latin)
		}
		if release.Description != "" {
			descParts = append(descParts, release.Description)
		}
		description := strings.Join(descParts, " / ")

		res := qualityToResolution(torrent.Quality.Value)
		cats := resolutionToCategories(res)

		var catXml strings.Builder
		for _, cat := range cats {
			catXml.WriteString(fmt.Sprintf("<category>%d</category>\n", cat))
		}

		sb.WriteString(fmt.Sprintf(`<item>
<title>%s</title>
<guid isPermaLink="false">%s</guid>
<link>%s</link>
<comments>%s</comments>
<pubDate>%s</pubDate>
<size>%d</size>
<description>%s</description>
%s<enclosure url="%s" length="%d" type="application/x-bittorrent" />
`,
			escapeXML(title),
			escapeXML(hash),
			escapeXML(detailsUrl),
			escapeXML(detailsUrl),
			escapeXML(pubDate),
			torrent.Size,
			escapeXML(description),
			catXml.String(),
			escapeXML(downloadUrl),
			torrent.Size,
		))

		if magnetUrl != "" {
			writeTorznabAttr(&sb, "magneturl", escapeXML(magnetUrl))
		}
		if hash != "" {
			writeTorznabAttr(&sb, "infohash", escapeXML(hash))
		}

		for _, cat := range cats {
			writeTorznabAttr(&sb, "category", cat)
		}
		writeTorznabAttr(&sb, "seeders", torrent.Seeders)
		writeTorznabAttr(&sb, "peers", torrent.Seeders+torrent.Leechers)
		writeTorznabAttr(&sb, "grabs", torrent.CompletedTimes)
		writeTorznabAttr(&sb, "minimumratio", 1)
		writeTorznabAttr(&sb, "minimumseedtime", 0)
		writeTorznabAttr(&sb, "downloadvolumefactor", 0)
		writeTorznabAttr(&sb, "uploadvolumefactor", 1)

		if res != "" {
			writeTorznabAttr(&sb, "resolution", res)
		}

		posterUrl := release.Poster.Optimized.Src
		if posterUrl == "" {
			posterUrl = release.Poster.Src
		}
		if posterUrl != "" {
			writeTorznabAttr(&sb, "poster", escapeXML(posterUrl))
		}

		sb.WriteString("</item>\n")
	}

	sb.WriteString("</channel>\n</rss>")
	return sb.String()
}

// Returns the indexer capabilities XML
func BuildCapsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<caps>
<server version="1.0" title="AniLiberty Indexer" />
<limits max="50" default="25" />
<searching>
<search available="yes" supportedParams="q" />
<tv-search available="yes" supportedParams="q" />
<movie-search available="yes" supportedParams="q" />
</searching>
<categories>
<category id="5000" name="TV">
<subcat id="5030" name="TV/SD" />
<subcat id="5040" name="TV/HD" />
<subcat id="5045" name="TV/UHD" />
<subcat id="5070" name="TV/Anime" />
</category>
</categories>
</caps>`
}
