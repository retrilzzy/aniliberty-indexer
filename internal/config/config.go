package config

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Global config
var (
	HOST            = "0.0.0.0"
	PORT            = "3649"
	ANILIBERTY_API  = "https://anilibria.top/api/v1"
	ANILIBERTY_SITE = "https://anilibria.top"
	API_KEY         = ""

	// Routes outgoing requests to AniLiberty through an HTTP/SOCKS5 proxy
	UPSTREAM_PROXY = ""

	// Defines the pattern for formatting torrent titles
	TORRENT_TITLE_TEMPLATE = "[AniLiberty] {title_latin_clean} - S{season} [RUS][{type} {quality} {codec}] ({year})"

	// Defines the pattern for formatting movie torrent titles
	TORRENT_MOVIE_TITLE_TEMPLATE = "[AniLiberty] {title_latin_clean} ({year}) [RUS][{type} {quality} {codec}]"

	DEFAULT_LIMIT      = 25
	MAX_LIMIT          = 50
	SEARCH_CONCURRENCY = 5
)

// Map of placeholders supported in TORRENT_TITLE_TEMPLATE
var AllowedPlaceholders = map[string]struct{}{
	"title_latin_clean": {},
	"title_latin":       {},
	"title_ru":          {},
	"season":            {},
	"episodes":          {},
	"season_episodes":   {},
	"year":              {},
	"type":              {},
	"quality":           {},
	"codec":             {},
}

// Checks if the given template has valid placeholders and matching braces
func ValidateTemplate(tpl string) error {
	inBrace := false
	var currentPlaceholder strings.Builder

	for i, char := range tpl {
		if char == '{' {
			if inBrace {
				return fmt.Errorf("nested braces are not allowed at position %d", i)
			}
			inBrace = true
			currentPlaceholder.Reset()
		} else if char == '}' {
			if !inBrace {
				return fmt.Errorf("unmatched closing brace at position %d", i)
			}
			inBrace = false
			name := currentPlaceholder.String()
			if name == "" {
				return fmt.Errorf("empty placeholder at position %d", i-1)
			}
			if _, ok := AllowedPlaceholders[name]; !ok {
				return fmt.Errorf("unknown placeholder: {%s}", name)
			}
		} else {
			if inBrace {
				currentPlaceholder.WriteRune(char)
			}
		}
	}

	if inBrace {
		return fmt.Errorf("unmatched opening brace at the end of the template")
	}

	return nil
}

// HTTP client for AniLiberty upstream requests
var HTTPClient *http.Client

// Reads env vars and sets up the HTTP client
func init() {
	if h := os.Getenv("HOST"); h != "" {
		HOST = h
	}
	if p := os.Getenv("PORT"); p != "" {
		PORT = p
	}
	if a := os.Getenv("ANILIBERTY_API"); a != "" {
		ANILIBERTY_API = a
	}
	if s := os.Getenv("ANILIBERTY_SITE"); s != "" {
		ANILIBERTY_SITE = s
	}
	if k := os.Getenv("API_KEY"); k != "" {
		API_KEY = k
	}
	if pr := os.Getenv("UPSTREAM_PROXY"); pr != "" {
		UPSTREAM_PROXY = pr
	}
	if t := os.Getenv("TORRENT_TITLE_TEMPLATE"); t != "" {
		TORRENT_TITLE_TEMPLATE = t
	}
	if t := os.Getenv("TORRENT_MOVIE_TITLE_TEMPLATE"); t != "" {
		TORRENT_MOVIE_TITLE_TEMPLATE = t
	}

	if err := ValidateTemplate(TORRENT_TITLE_TEMPLATE); err != nil {
		log.Fatalf("Invalid TORRENT_TITLE_TEMPLATE: %v\n", err)
	}
	if err := ValidateTemplate(TORRENT_MOVIE_TITLE_TEMPLATE); err != nil {
		log.Fatalf("Invalid TORRENT_MOVIE_TITLE_TEMPLATE: %v\n", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if UPSTREAM_PROXY != "" {
		proxyURL, err := url.Parse(UPSTREAM_PROXY)
		if err != nil {
			log.Fatalf("Invalid UPSTREAM_PROXY: %v\n", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	HTTPClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}
}

// Pings AniLiberty site to check proxy status
func PingProxy() string {
	if UPSTREAM_PROXY == "" {
		return "disabled"
	}

	start := time.Now()
	req, err := http.NewRequest("HEAD", ANILIBERTY_SITE, nil)
	if err != nil {
		return fmt.Sprintf("error (%v)", err)
	}
	req.Header.Set("User-Agent", "AniLiberty-Indexer/1.0")

	res, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Sprintf("%s - unreachable (%v)", UPSTREAM_PROXY, err)
	}
	res.Body.Close()

	return fmt.Sprintf("%s - ok (%dms)", UPSTREAM_PROXY, time.Since(start).Milliseconds())
}
