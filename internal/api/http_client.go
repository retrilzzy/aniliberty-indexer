package api

import (
	"aniliberty-indexer/internal/config"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Request to AniLiberty API; deserializes JSON into target
func FetchJSON(requestUrl string, target interface{}) error {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AniLiberty-Indexer/1.0")

	res, err := config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(res.Body)
		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200]
		}
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, bodyStr)
	}

	return json.NewDecoder(res.Body).Decode(target)
}

// Relays a binary file from AniLiberty upstream to the client
func RelayTorrentFile(upstreamUrl string, w http.ResponseWriter) {
	req, err := http.NewRequest("GET", upstreamUrl, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upstream error: %v", err), http.StatusBadGateway)
		return
	}
	req.Header.Set("User-Agent", "AniLiberty-Indexer/1.0")

	res, err := config.HTTPClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upstream error: %v", err), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/x-bittorrent"
	}
	w.Header().Set("Content-Type", contentType)

	if cd := res.Header.Get("Content-Disposition"); cd != "" {
		w.Header().Set("Content-Disposition", cd)
	}

	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		log.Printf("[relay] copy error for %s: %v\n", upstreamUrl, err)
	}
}
