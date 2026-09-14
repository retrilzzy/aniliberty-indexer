package torznab

import (
	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"
	"fmt"
	"strconv"
	"strings"
)

var wordToNum = map[string]int{
	"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5,
	"sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
	"i": 1, "ii": 2, "iii": 3, "iv": 4, "v": 5,
	"vi": 6, "vii": 7, "viii": 8, "ix": 9, "x": 10,
}

// Attempts to parse a string into a season number, handling both numeric and word-based inputs
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

// Checks if a given word indicates a season
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

// Maps a video resolution string to corresponding Torznab TV category IDs
func resolutionToTVCategories(res string) []int {
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

// Constructs a formatted torrent title for a TV series based on the release and torrent data
func BuildTVTitle(release api.Release, torrent api.Torrent) string {
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

	seasonStr := fmt.Sprintf("S%02d", season)
	seasonEpisodes := seasonStr
	if episodes != "" {
		if strings.Contains(episodes, "-") {
			parts := strings.SplitN(episodes, "-", 2)
			seasonEpisodes = seasonStr + parts[0] + "-" + seasonStr + parts[1]
		} else {
			seasonEpisodes = seasonStr + episodes
		}
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
