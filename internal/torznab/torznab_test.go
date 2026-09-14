package torznab

import (
	"testing"

	"aniliberty-indexer/internal/api"
	"aniliberty-indexer/internal/config"
)

func TestExtractEpisodes(t *testing.T) {
	tests := []struct {
		label    string
		expected string
	}{
		{"Jidou Hanbaiki ni Umarekawatta Ore wa Meikyuu wo Samayou 3rd Season - AniLiberty.TOP [WEBRip 1080p][AVC][1-12]", "E01-E12"},
		{"Frieren [WEBRip 1080p][HEVC][05]", "E05"},
		{"Frieren [WEBRip 1080p][HEVC][01-24 из 24]", "E01-E24"},
		{"Frieren [WEBRip 1080p]", ""},
		{"No brackets at all", ""},
		{"Title [1-5]", "E01-E05"},
		{"Title [7]", "E07"},
	}

	for _, tc := range tests {
		got := extractEpisodes(tc.label)
		if got != tc.expected {
			t.Errorf("extractEpisodes(%q) = %q; expected %q", tc.label, got, tc.expected)
		}
	}
}

func TestExtractSeason(t *testing.T) {
	tests := []struct {
		title    string
		expected int
	}{
		{"Some Title Season 3", 3},
		{"Some Title 2nd Season", 2},
		{"Another Title season 12", 12},
		{"Title 4", 4},
		{"Title with 200 in it", 1},
		{"Title with no season", 1},
		{"Title 3rd season part 2", 3},
		{"Monogatari Series: Second Season", 2},
		{"Monogatari Series: Off & Monster Season", 1},
		{"Devil May Cry Season 2", 2},
		{"One Punch Man 2nd Season", 2},
		{"One Punch Man 2nd Season. Episode Zero", 2},
		{"Lv2 kara Cheat datta Motoyuusha Kouho no Mattari Isekai Life", 1},
		{"Luck & Logic", 1},
		{"Sakamoto Days Part 2", 2},
		{"Hito! Umr 2", 2},
		{"Kimi ni Todoke 2nd Season", 2},
		{"Shakugan no Shana Second", 2},
		{"Around 40 Otoko no Isekai Tsuuhan", 1},
		{"Hanyou no Yashahime: Ni no Shou", 1},
		{"Dolls' Frontline", 1},
		{"Maou Gakuin no Futekigousha II", 2},
		{"Kimetsu no Yaiba Movie 1: Mugenjou-hen - Akaza Sairai", 1},
		{"UQ Holder!: Mahou Sensei Negima! 2", 2},
		{"Danganronpa 3: The End of Kibougamine Gakuen - Mirai-hen", 1},
		{"Детективное агентство Хаматора [TV-2]", 2},
		{"Re: Hamatora", 1},
		{"Diamond no Ace: Act II Second Season", 2},
		{"Diamond no Ace: Act II", 2},
	}

	for _, tc := range tests {
		got := extractSeason(tc.title)
		if got != tc.expected {
			t.Errorf("extractSeason(%q) = %d; expected %d", tc.title, got, tc.expected)
		}
	}
}

func TestCleanTitleSeason(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Some Title Season 3", "Some Title"},
		{"Some Title 2nd Season", "Some Title"},
		{"Title 4", "Title"},
		{"Title with 200 in it", "Title with 200 in it"},
		{"Title 3rd season part 2", "Title part 2"},
		{"Normal Title", "Normal Title"},
		{"Another Title -", "Another Title"},
		{"Ending with colon:", "Ending with colon"},
		{"Monogatari Series: Second Season", "Monogatari Series"},
		{"Monogatari Series: Off & Monster Season", "Monogatari Series: Off & Monster Season"},
		{"Devil May Cry Season 2", "Devil May Cry"},
		{"One Punch Man 2nd Season", "One Punch Man"},
		{"One Punch Man 2nd Season. Episode Zero", "One Punch Man Episode Zero"},
		{"Lv2 kara Cheat datta Motoyuusha Kouho no Mattari Isekai Life", "Lv2 kara Cheat datta Motoyuusha Kouho no Mattari Isekai Life"},
		{"Luck & Logic", "Luck & Logic"},
		{"Sakamoto Days Part 2", "Sakamoto Days"},
		{"Hito! Umr 2", "Hito! Umr"},
		{"Kimi ni Todoke 2nd Season", "Kimi ni Todoke"},
		{"Shakugan no Shana Second", "Shakugan no Shana"},
		{"Around 40 Otoko no Isekai Tsuuhan", "Around 40 Otoko no Isekai Tsuuhan"},
		{"Hanyou no Yashahime: Ni no Shou", "Hanyou no Yashahime: Ni no Shou"},
		{"Dolls' Frontline", "Dolls' Frontline"},
		{"Maou Gakuin no Futekigousha II", "Maou Gakuin no Futekigousha"},
		{"Kimetsu no Yaiba Movie 1: Mugenjou-hen - Akaza Sairai", "Kimetsu no Yaiba Movie 1: Mugenjou-hen - Akaza Sairai"},
		{"UQ Holder!: Mahou Sensei Negima! 2", "UQ Holder!: Mahou Sensei Negima!"},
		{"Danganronpa 3: The End of Kibougamine Gakuen - Mirai-hen", "Danganronpa 3: The End of Kibougamine Gakuen - Mirai-hen"},
		{"Детективное агентство Хаматора [TV-2]", "Детективное агентство Хаматора"},
		{"Re: Hamatora", "Re: Hamatora"},
		{"Diamond no Ace: Act II Second Season", "Diamond no Ace: Act II"},
		{"Diamond no Ace: Act II", "Diamond no Ace: Act II"},
	}

	for _, tc := range tests {
		got := cleanTitleSeason(tc.title)
		if got != tc.expected {
			t.Errorf("cleanTitleSeason(%q) = %q; expected %q", tc.title, got, tc.expected)
		}
	}
}

func TestCleanFinalTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[AniLiberty] Title - S01 [RUS] [WEB-DL 1080p] [x265] (RuTitle, 2023)", "[AniLiberty] Title - S01 [RUS] [WEB-DL 1080p] [x265] (RuTitle, 2023)"},
		{"[AniLiberty] Title - S01 [RUS] [] [x265] (, 2023)", "[AniLiberty] Title - S01 [RUS] [x265] (2023)"},
		{"[AniLiberty] Title - S01 [RUS] [ ] [x265] (RuTitle, )", "[AniLiberty] Title - S01 [RUS] [x265] (RuTitle)"},
		{"[AniLiberty] Title - S01 [RUS] [] [] (, )", "[AniLiberty] Title - S01 [RUS]"},
		{"- Title - S01 ,", "Title - S01"},
		{"() [] {} (, ) [, ] {, }", ""},
		{"[ ] ( ) { }", ""},
		{"[AniLiberty] Title -  [RUS]", "[AniLiberty] Title - [RUS]"},
	}

	for _, tc := range tests {
		got := cleanFinalTitle(tc.input)
		if got != tc.expected {
			t.Errorf("cleanFinalTitle(%q) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestBuildTVTitleSeasonEpisodes(t *testing.T) {
	origTpl := config.TORRENT_TITLE_TEMPLATE
	defer func() { config.TORRENT_TITLE_TEMPLATE = origTpl }()

	config.TORRENT_TITLE_TEMPLATE = "[AniLiberty] {title_latin_clean} - {season_episodes} [RUS][{type} {quality} {codec}] ({year})"

	release := api.Release{
		Name: struct {
			Main  string `json:"main"`
			Latin string `json:"english"`
		}{
			Main:  "Цикл Историй: Второй Сезон",
			Latin: "Monogatari Series: Second Season",
		},
		Year: 2013,
	}

	tests := []struct {
		label    string
		expected string
	}{
		{"[E01-E23]", "[AniLiberty] Monogatari Series - S02E01-S02E23 [RUS][WEBRip 1080p x264] (2013)"},
		{"[05]", "[AniLiberty] Monogatari Series - S02E05 [RUS][WEBRip 1080p x264] (2013)"},
		{"", "[AniLiberty] Monogatari Series - S02 [RUS][WEBRip 1080p x264] (2013)"},
	}

	for _, tc := range tests {
		torrent := api.Torrent{
			Type: struct {
				Value string `json:"value"`
			}{Value: "WEBRip"},
			Quality: struct {
				Value string `json:"value"`
			}{Value: "1080p"},
			Codec: struct {
				Value string `json:"value"`
			}{Value: "x264"},
			Label: tc.label,
		}

		got := BuildTVTitle(release, torrent)
		if got != tc.expected {
			t.Errorf("BuildTVTitle with label %q = %q; expected %q", tc.label, got, tc.expected)
		}
	}
}
