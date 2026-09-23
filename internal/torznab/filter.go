package torznab

import (
	"strconv"
	"strings"

	"aniliberty-indexer/internal/api"
)

type FilterOptions struct {
	Categories []int
	Season     int
	Episode    int
}

// Filters combined items by category, season, and episode
func FilterItems(items []api.CombinedItem, opts FilterOptions) []api.CombinedItem {
	if len(opts.Categories) == 0 && opts.Season <= 0 && opts.Episode <= 0 {
		return items
	}

	filtered := make([]api.CombinedItem, 0, len(items))
	for _, item := range items {
		if itemMatches(item, opts) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func itemMatches(item api.CombinedItem, opts FilterOptions) bool {
	isMovie := item.Release.Type.Value == "MOVIE"
	res := qualityToResolution(item.Torrent.Quality.Value)

	var itemCats []int
	if isMovie {
		itemCats = resolutionToMovieCategories(res)
	} else {
		itemCats = resolutionToTVCategories(res)
	}

	// Filter by category
	if len(opts.Categories) > 0 {
		catMatch := false
		for _, reqCat := range opts.Categories {
			for _, itemCat := range itemCats {
				if reqCat == itemCat {
					catMatch = true
					break
				}
			}
			if catMatch {
				break
			}
		}
		if !catMatch {
			return false
		}
	}

	// Filter by season
	if opts.Season > 0 {
		if isMovie {
			return false
		}
		season := extractSeason(item.Release.Name.Latin)
		if season == 1 && extractSeason(item.Release.Name.Main) > 1 {
			season = extractSeason(item.Release.Name.Main)
		}
		if season != opts.Season {
			return false
		}
	}

	// Filter by episode
	if opts.Episode > 0 {
		if isMovie {
			return false
		}
		epStr := extractEpisodes(item.Torrent.Label)
		if epStr != "" {
			if strings.Contains(epStr, "-") {
				parts := strings.SplitN(epStr, "-", 2)
				startEp, err1 := strconv.Atoi(strings.TrimPrefix(parts[0], "E"))
				endEp, err2 := strconv.Atoi(strings.TrimPrefix(parts[1], "E"))
				if err1 == nil && err2 == nil {
					if opts.Episode < startEp || opts.Episode > endEp {
						return false
					}
				}
			} else {
				epNum, err := strconv.Atoi(strings.TrimPrefix(epStr, "E"))
				if err == nil && epNum != opts.Episode {
					return false
				}
			}
		}
	}

	return true
}
