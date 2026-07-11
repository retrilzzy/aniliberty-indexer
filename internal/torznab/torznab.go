package torznab

import (
	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"

	"fmt"
	"strconv"
	"strings"
	"time"
)

// Writes a single <torznab:attr> element
func writeTorznabAttr(sb *strings.Builder, name string, value interface{}) {
	sb.WriteString("<torznab:attr name=\"")
	sb.WriteString(name)
	sb.WriteString("\" value=\"")
	switch v := value.(type) {
	case string:
		sb.WriteString(v)
	case int:
		sb.WriteString(strconv.Itoa(v))
	case int64:
		sb.WriteString(strconv.FormatInt(v, 10))
	default:
		sb.WriteString(fmt.Sprintf("%v", v))
	}
	sb.WriteString("\" />\n")
}

// Builds the Torznab RSS feed XML from a list of items
func BuildTorznabXML(items []api.CombinedItem) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:torznab="http://torznab.com/schemas/2015/feed">
<channel>
<title>AniLiberty</title>
<link>%s</link>
`, escapeXML(config.ANILIBERTY_SITE)))

	for _, item := range items {
		release := item.Release
		torrent := item.Torrent

		isMovie := release.Type.Value == "MOVIE"

		var title string
		var cats []int

		res := qualityToResolution(torrent.Quality.Value)

		if isMovie {
			title = BuildMovieTitle(release, torrent)
			cats = resolutionToMovieCategories(res)
		} else {
			title = BuildTVTitle(release, torrent)
			cats = resolutionToTVCategories(res)
		}

		hash := torrent.Hash

		releaseAlias := release.Alias
		if releaseAlias == "" && release.ID != 0 {
			releaseAlias = strconv.Itoa(release.ID)
		}

		detailsUrl := config.ANILIBERTY_SITE + "/anime/releases/release/" + releaseAlias
		downloadUrl := config.ANILIBERTY_API + "/anime/torrents/" + hash + "/file"
		magnetUrl := torrent.Magnet

		pubDate := time.Now().UTC().Format(time.RFC1123)
		if torrent.UpdatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, torrent.UpdatedAt); err == nil {
				pubDate = parsed.UTC().Format(time.RFC1123)
			}
		}

		var descParts []string
		if release.Name.Main != "" {
			descParts = append(descParts, release.Name.Main)
		}
		if release.Name.Latin != "" {
			descParts = append(descParts, release.Name.Latin)
		}
		if release.Description != "" {
			descParts = append(descParts, release.Description)
		}
		description := strings.Join(descParts, " / ")

		var catXml strings.Builder
		for _, cat := range cats {
			catXml.WriteString(fmt.Sprintf("<category>%d</category>\n", cat))
		}

		sb.WriteString(fmt.Sprintf(`<item>
<title>%s</title>
<guid isPermaLink="false">%s</guid>
<link>%s</link>
<comments>%s</comments>
<pubDate>%s</pubDate>
<size>%d</size>
<description>%s</description>
%s<enclosure url="%s" length="%d" type="application/x-bittorrent" />
`,
			escapeXML(title),
			escapeXML(hash),
			escapeXML(detailsUrl),
			escapeXML(detailsUrl),
			escapeXML(pubDate),
			torrent.Size,
			escapeXML(description),
			catXml.String(),
			escapeXML(downloadUrl),
			torrent.Size,
		))

		if magnetUrl != "" {
			writeTorznabAttr(&sb, "magneturl", escapeXML(magnetUrl))
		}
		if hash != "" {
			writeTorznabAttr(&sb, "infohash", escapeXML(hash))
		}

		for _, cat := range cats {
			writeTorznabAttr(&sb, "category", cat)
		}
		writeTorznabAttr(&sb, "seeders", torrent.Seeders)
		writeTorznabAttr(&sb, "peers", torrent.Seeders+torrent.Leechers)
		writeTorznabAttr(&sb, "grabs", torrent.CompletedTimes)
		writeTorznabAttr(&sb, "minimumratio", 1)
		writeTorznabAttr(&sb, "minimumseedtime", 0)
		writeTorznabAttr(&sb, "downloadvolumefactor", 0)
		writeTorznabAttr(&sb, "uploadvolumefactor", 1)

		if res != "" {
			writeTorznabAttr(&sb, "resolution", res)
		}

		posterUrl := release.Poster.Optimized.Src
		if posterUrl == "" {
			posterUrl = release.Poster.Src
		}
		if posterUrl != "" {
			writeTorznabAttr(&sb, "poster", escapeXML(posterUrl))
		}

		sb.WriteString("</item>\n")
	}

	sb.WriteString("</channel>\n</rss>")
	return sb.String()
}

// Returns the indexer capabilities XML
func BuildCapsXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<caps>
<server version="1.0" title="AniLiberty Indexer" />
<limits max="50" default="25" />
<searching>
<search available="yes" supportedParams="q" />
<tv-search available="yes" supportedParams="q" />
<movie-search available="yes" supportedParams="q" />
</searching>
<categories>
<category id="2000" name="Movies">
<subcat id="2030" name="Movies/SD" />
<subcat id="2040" name="Movies/HD" />
<subcat id="2045" name="Movies/UHD" />
<subcat id="2050" name="Movies/BluRay" />
</category>
<category id="5000" name="TV">
<subcat id="5030" name="TV/SD" />
<subcat id="5040" name="TV/HD" />
<subcat id="5045" name="TV/UHD" />
<subcat id="5070" name="TV/Anime" />
</category>
</categories>
</caps>`
}
