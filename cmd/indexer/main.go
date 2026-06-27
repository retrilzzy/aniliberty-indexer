/**
 * AniLiberty Indexer
 *
 * Lightweight Torznab-compatible indexer that bridges
 * AniLiberty (AniLibria) API with Prowlarr and Sonarr.
 */
package main

import (
	"aniliberty-indexer/internal/config"
	"aniliberty-indexer/internal/server"
	"aniliberty-indexer/internal/torznab"

	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", server.HandleRequest)

	displayHost := config.HOST

	apiKeyState := "disabled"
	if config.API_KEY != "" {
		apiKeyState = "enabled"
	}

	getValue := func(key string) string {
		switch key {
		case "title_latin_clean":
			return "Monogatari Series"
		case "title_latin":
			return "Monogatari Series: Second Season"
		case "title_ru":
			return "Цикл Историй: Второй Сезон"
		case "season":
			return "2"
		case "episodes":
			return "E01-E23"
		case "season_episodes":
			return "S02E01-E23"
		case "year":
			return "2013"
		case "type":
			return "WEBRip"
		case "quality":
			return "1080p"
		case "codec":
			return "x264"
		default:
			return ""
		}
	}
	exampleTitle := torznab.RenderTemplate(config.TORRENT_TITLE_TEMPLATE, getValue)

	log.Printf(`
╔═══
║  AniLiberty Indexer
║
║  Server:         http://%s:%s
║  API Key:        %s
║  Upstream proxy: %s
║
║  Title Template: %s
║  Title Example:  %s
╚═══
`, displayHost, config.PORT, apiKeyState, config.PingProxy(), config.TORRENT_TITLE_TEMPLATE, exampleTitle)

	if err := http.ListenAndServe(config.HOST+":"+config.PORT, nil); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
