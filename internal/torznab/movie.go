package torznab

import (
	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"
	"strconv"
)

// Maps a video resolution string to corresponding Torznab movie category IDs
func resolutionToMovieCategories(res string) []int {
	cats := []int{2000, 2070}
	switch {
	case res == "1080" || res == "720" || res == "1440":
		cats = append(cats, 2040) // Movies/HD
	case res == "2160":
		cats = append(cats, 2045) // Movies/UHD
	case res == "576" || res == "480" || res == "360":
		cats = append(cats, 2030) // Movies/SD
	}
	return cats
}

// Constructs a formatted torrent title for a movie based on the release and torrent data
func BuildMovieTitle(release api.Release, torrent api.Torrent) string {
	ruName := release.Name.Main
	latinName := release.Name.Latin

	yearStr := ""
	if release.Year != 0 {
		yearStr = strconv.Itoa(release.Year)
	}

	codecCleaned := ""
	if torrent.Codec.Value != "" {
		codecCleaned = cleanCodec(torrent.Codec.Value)
	}

	getValue := func(key string) string {
		switch key {
		case "title_latin_clean", "title_latin":
			return latinName
		case "title_ru":
			return ruName
		case "year":
			return yearStr
		case "type":
			return torrent.Type.Value
		case "quality":
			return torrent.Quality.Value
		case "codec":
			return codecCleaned
		case "season_episodes", "season", "episodes":
			return "" // ignore
		default:
			return ""
		}
	}

	return RenderTemplate(config.TORRENT_MOVIE_TITLE_TEMPLATE, getValue)
}
