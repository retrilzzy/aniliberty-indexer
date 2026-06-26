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

	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", server.HandleRequest)

	displayHost := config.HOST
	if displayHost == "" {
		displayHost = "0.0.0.0"
	}

	apiKeyState := "disabled"
	if config.API_KEY != "" {
		apiKeyState = "enabled"
	}

	log.Printf(`
╔═══════════════════════════════════════════
║  AniLiberty Indexer
║
║  Server:         http://%s:%s
║  API Key:        %s
║  Upstream proxy: %s
╚═══════════════════════════════════════════
`, displayHost, config.PORT, apiKeyState, config.PingProxy())

	if err := http.ListenAndServe(config.HOST+":"+config.PORT, nil); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}
