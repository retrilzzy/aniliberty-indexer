package main

import (
	"aniliberty-indexer/internal/api"
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

	tvRelease := api.Release{
		Name: struct {
			Main  string `json:"main"`
			Latin string `json:"english"`
		}{
			Main:  "Цикл Историй: Второй Сезон",
			Latin: "Monogatari Series: Second Season",
		},
		Year: 2013,
	}
	tvTorrent := api.Torrent{
		Type: struct {
			Value string `json:"value"`
		}{Value: "WEBRip"},
		Quality: struct {
			Value string `json:"value"`
		}{Value: "1080p"},
		Codec: struct {
			Value string `json:"value"`
		}{Value: "x264"},
		Label: "[E01-E23]",
	}
	exampleTVTitle := torznab.BuildTVTitle(tvRelease, tvTorrent)

	movieRelease := api.Release{
		Name: struct {
			Main  string `json:"main"`
			Latin string `json:"english"`
		}{
			Main:  "Врата Штейна: Зона загрузки дежавю",
			Latin: "Gekijouban Steins;Gate: Fuka Ryouiki no Deja vu",
		},
		Year: 2013,
		Type: struct {
			Value string `json:"value"`
		}{Value: "MOVIE"},
	}
	movieTorrent := api.Torrent{
		Type: struct {
			Value string `json:"value"`
		}{Value: "BDRip"},
		Quality: struct {
			Value string `json:"value"`
		}{Value: "720p"},
		Codec: struct {
			Value string `json:"value"`
		}{Value: "x265"},
	}
	exampleMovieTitle := torznab.BuildMovieTitle(movieRelease, movieTorrent)

	log.Printf(`
╔═══
║  AniLiberty Indexer
║
║  Server:         http://%s:%s
║  API Key:        %s
║  Upstream proxy: %s
║
║  TV Template:    %s
║  TV Example:     %s
║
║  Movie Template: %s
║  Movie Example:  %s
╚═══
`, displayHost, config.PORT, apiKeyState, config.PingProxy(), config.TORRENT_TITLE_TEMPLATE, exampleTVTitle, config.TORRENT_MOVIE_TITLE_TEMPLATE, exampleMovieTitle)

	if err := http.ListenAndServe(config.HOST+":"+config.PORT, nil); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
