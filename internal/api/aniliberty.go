package api

import (
	"aniliberty-indexer/internal/config"

	"fmt"
	"log"
	"net/url"
	"sync"
)

// Fetches the latest torrents from AniLiberty with pagination
func FetchLatestTorrents(limit int, page int) ([]CombinedItem, error) {
	apiUrl := fmt.Sprintf("%s/anime/torrents?page=%d&limit=%d", config.ANILIBERTY_API, page, limit)
	var data LatestTorrentsResponse
	if err := FetchJSON(apiUrl, &data); err != nil {
		return nil, err
	}

	items := make([]CombinedItem, 0, len(data.Data))
	for _, item := range data.Data {
		items = append(items, CombinedItem{
			Release: item.Release,
			Torrent: item.Torrent,
		})
	}
	return items, nil
}

// Searches releases by query, then fetches their torrents concurrently
func SearchTorrents(query string, limit int) ([]CombinedItem, error) {
	searchUrl := fmt.Sprintf("%s/app/search/releases?query=%s", config.ANILIBERTY_API, url.QueryEscape(query))
	var releases []Release
	if err := FetchJSON(searchUrl, &releases); err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return []CombinedItem{}, nil
	}

	var results []CombinedItem
	concurrency := config.SEARCH_CONCURRENCY

	for i := 0; i < len(releases) && len(results) < limit; i += concurrency {
		end := i + concurrency
		if end > len(releases) {
			end = len(releases)
		}
		batch := releases[i:end]

		var wg sync.WaitGroup
		var mu sync.Mutex
		torrentsMap := make(map[int][]Torrent)

		for idx, release := range batch {
			wg.Add(1)
			go func(index int, rel Release) {
				defer wg.Done()
				torrentsUrl := fmt.Sprintf("%s/anime/torrents/release/%d", config.ANILIBERTY_API, rel.ID)
				var torrents []Torrent
				if err := FetchJSON(torrentsUrl, &torrents); err != nil {
					log.Printf("Failed to fetch torrents for release %d: %v\n", rel.ID, err)
					return
				}
				mu.Lock()
				torrentsMap[index] = torrents
				mu.Unlock()
			}(idx, release)
		}
		wg.Wait()

		for idx, release := range batch {
			if torrents, ok := torrentsMap[idx]; ok {
				for _, t := range torrents {
					results = append(results, CombinedItem{
						Release: release,
						Torrent: t,
					})
				}
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
