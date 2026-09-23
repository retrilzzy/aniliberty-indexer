package server

import (
	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"
	"aniliberty-indexer/internal/torznab"

	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Main HTTP handler: /, /health, /download/, /api, /torznab
func HandleRequest(w http.ResponseWriter, req *http.Request) {
	parsedUrl := req.URL
	pathname := parsedUrl.Path
	query := parsedUrl.Query()

	// Health check
	if pathname == "/" || pathname == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","name":"AniLiberty Indexer"}`))
		return
	}

	// Validate the API key (unrelated to AniLiberty)
	if config.API_KEY != "" {
		if !checkAPIKey(req, query) {
			respondUnauthorized(w, pathname)
			return
		}
	}

	// Relay a .torrent file from AniLiberty API to the client
	if strings.HasPrefix(pathname, "/download/") {
		hashOrId := strings.TrimPrefix(pathname, "/download/")
		pk := query.Get("pk")
		downloadUrl := fmt.Sprintf("%s/anime/torrents/%s/file", config.ANILIBERTY_API, hashOrId)
		if pk != "" {
			downloadUrl += "?pk=" + url.QueryEscape(pk)
		}
		api.RelayTorrentFile(downloadUrl, w)
		return
	}

	// Torznab API
	if pathname == "/api" || pathname == "/torznab" || pathname == "/torznab/api" {
		handleTorznab(w, req, query)
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
}

// Checks the service API key from query, X-Api-Key header, or Authorization Bearer
func checkAPIKey(req *http.Request, query url.Values) bool {
	reqKey := query.Get("apikey")
	if reqKey == "" {
		reqKey = req.Header.Get("X-Api-Key")
	}
	if reqKey == "" {
		authHeader := req.Header.Get("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				reqKey = authHeader[7:]
			} else {
				reqKey = authHeader
			}
		}
	}
	return reqKey == config.API_KEY
}

// Replies 401 (XML for Torznab paths, JSON for everything else)
func respondUnauthorized(w http.ResponseWriter, pathname string) {
	if strings.HasPrefix(pathname, "/api") || strings.HasPrefix(pathname, "/torznab") {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<error code="100" description="Invalid API Key" />`))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized: Invalid or missing API key"}`))
	}
}

// Routes Torznab requests by the t= parameter: caps, search, tvsearch, movie, or browse
func handleTorznab(w http.ResponseWriter, req *http.Request, query url.Values) {
	t := strings.ToLower(query.Get("t"))

	if t == "caps" {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(torznab.BuildCapsXML()))
		return
	}

	if t == "" || t == "browse" || t == "search" || t == "tvsearch" || t == "tv" || t == "movie" || t == "movie-search" || t == "moviesearch" {
		handleSearch(w, req, query)
		return
	}

	errMsg := "Unknown Torznab action"
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<error code="201" description="` + errMsg + `" />`))
}

// Parses comma-separated category IDs from cat query parameters
func parseCategories(query url.Values) []int {
	var cats []int
	for _, val := range query["cat"] {
		for _, part := range strings.Split(val, ",") {
			part = strings.TrimSpace(part)
			if c, err := strconv.Atoi(part); err == nil {
				cats = append(cats, c)
			}
		}
	}
	return cats
}

var yearSuffixRegex = regexp.MustCompile(`\s+\d{4}$`)

// Handles search (q= set) or browse (q= empty) and returns a Torznab XML feed
func handleSearch(w http.ResponseWriter, req *http.Request, query url.Values) {
	q := query.Get("q")
	t := strings.ToLower(query.Get("t"))

	if q != "" {
		q = yearSuffixRegex.ReplaceAllString(q, "")
	}

	limit := config.DEFAULT_LIMIT
	if l, err := strconv.Atoi(query.Get("limit")); err == nil {
		limit = l
	}
	if limit > config.MAX_LIMIT {
		limit = config.MAX_LIMIT
	}
	offset := 0
	if o, err := strconv.Atoi(query.Get("offset")); err == nil {
		offset = o
	}

	cats := parseCategories(query)
	if len(cats) == 0 && (t == "movie" || t == "movie-search" || t == "moviesearch") {
		cats = []int{2000, 2030, 2040, 2045, 2050, 2070}
	}

	var season int
	if s, err := strconv.Atoi(query.Get("season")); err == nil {
		season = s
	}
	var ep int
	if e, err := strconv.Atoi(query.Get("ep")); err == nil {
		ep = e
	}

	hasFilter := len(cats) > 0 || season > 0 || ep > 0

	var items []api.CombinedItem
	var err error

	if q != "" {
		searchLimit := limit
		if hasFilter && searchLimit < config.MAX_LIMIT*2 {
			searchLimit = config.MAX_LIMIT * 2
		}
		log.Printf("[search] query=\"%s\" limit=%d\n", q, searchLimit)
		items, err = api.SearchTorrents(q, searchLimit)
	} else {
		page := 1
		if limit > 0 {
			page = (offset / limit) + 1
		}
		if page < 1 {
			page = 1
		}
		log.Printf("[browse] limit=%d offset=%d page=%d\n", limit, offset, page)
		items, err = api.FetchLatestTorrents(limit, page)
	}

	if err != nil {
		log.Printf("[error] %v\n", err)
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	if hasFilter {
		items = torznab.FilterItems(items, torznab.FilterOptions{
			Categories: cats,
			Season:     season,
			Episode:    ep,
		})
	}

	if len(items) > limit {
		items = items[:limit]
	}

	baseURL := config.BASE_URL
	if baseURL == "" && req != nil {
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := req.Host
		if host == "" {
			host = config.HOST + ":" + config.PORT
		}
		baseURL = scheme + "://" + host
	}

	xmlResponse := torznab.BuildTorznabXML(items, baseURL)
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xmlResponse))
}
