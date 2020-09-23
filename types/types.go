package types

import "github.com/dgrijalva/jwt-go"

type SingleTrack struct {
	Title       string   `json:"title"`
	Duration    int      `json:"duration"`
	Artistes    []string `json:"artistes"`
	URL         string   `json:"url"`
	Preview     string   `json:"preview"`
	Cover       string   `json:"cover"`
	ReleaseDate string   `json:"release_date"`
	Explicit    bool     `json:"explicit"`
	Platform    string   `json:"platform"`
	ID          string   `json:"id"`
}

type HostSpotifyAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type HostDeezerTrack struct {
	ID                    int      `json:"id"`
	Readable              bool     `json:"readable"`
	Title                 string   `json:"title"`
	TitleShort            string   `json:"title_short"`
	TitleVersion          string   `json:"title_version"`
	Isrc                  string   `json:"isrc"`
	Link                  string   `json:"link"`
	Share                 string   `json:"share"`
	Duration              int      `json:"duration"`
	TrackPosition         int      `json:"track_position"`
	DiskNumber            int      `json:"disk_number"`
	Rank                  int      `json:"rank"`
	ReleaseDate           string   `json:"release_date"`
	ExplicitLyrics        bool     `json:"explicit_lyrics"`
	ExplicitContentLyrics int      `json:"explicit_content_lyrics"`
	ExplicitContentCover  int      `json:"explicit_content_cover"`
	Preview               string   `json:"preview"`
	Bpm                   float64  `json:"bpm"`
	Gain                  float64  `json:"gain"`
	AvailableCountries    []string `json:"available_countries"`
	Contributors          []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Link          string `json:"link"`
		Share         string `json:"share"`
		Picture       string `json:"picture"`
		PictureSmall  string `json:"picture_small"`
		PictureMedium string `json:"picture_medium"`
		PictureBig    string `json:"picture_big"`
		PictureXl     string `json:"picture_xl"`
		Radio         bool   `json:"radio"`
		Tracklist     string `json:"tracklist"`
		Type          string `json:"type"`
		Role          string `json:"role"`
	} `json:"contributors"`
	Md5Image string `json:"md5_image"`
	Artist   struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Link          string `json:"link"`
		Share         string `json:"share"`
		Picture       string `json:"picture"`
		PictureSmall  string `json:"picture_small"`
		PictureMedium string `json:"picture_medium"`
		PictureBig    string `json:"picture_big"`
		PictureXl     string `json:"picture_xl"`
		Radio         bool   `json:"radio"`
		Tracklist     string `json:"tracklist"`
		Type          string `json:"type"`
	} `json:"artist"`
	Album struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Link        string `json:"link"`
		Cover       string `json:"cover"`
		CoverSmall  string `json:"cover_small"`
		CoverMedium string `json:"cover_medium"`
		CoverBig    string `json:"cover_big"`
		CoverXl     string `json:"cover_xl"`
		Md5Image    string `json:"md5_image"`
		ReleaseDate string `json:"release_date"`
		Tracklist   string `json:"tracklist"`
		Type        string `json:"type"`
	} `json:"album"`
	Type string `json:"type"`
}

type HostSpotifyTrack struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`
	DiscNumber       int      `json:"disc_number"`
	DurationMs       int      `json:"duration_ms"`
	Explicit         bool     `json:"explicit"`
	ExternalIds      struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string `json:"href"`
	ID          string `json:"id"`
	IsLocal     bool   `json:"is_local"`
	Name        string `json:"name"`
	Popularity  int    `json:"popularity"`
	PreviewURL  string `json:"preview_url"`
	TrackNumber int    `json:"track_number"`
	Type        string `json:"type"`
	URI         string `json:"uri"`
}

type ExtractedInfo struct {
	Host string
	URL  string
	ID   string
}

type TrackToSearch struct {
	Title   string
	Artiste string
}

type HostSpotifySearchTrack struct {
	Tracks struct {
		Href  string `json:"href"`
		Items []struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height int    `json:"height"`
					URL    string `json:"url"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"release_date_precision"`
				TotalTracks          int    `json:"total_tracks"`
				Type                 string `json:"type"`
				URI                  string `json:"uri"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string `json:"href"`
			ID          string `json:"id"`
			IsLocal     bool   `json:"is_local"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
		} `json:"items"`
		Limit    int         `json:"limit"`
		Next     interface{} `json:"next"`
		Offset   int         `json:"offset"`
		Previous interface{} `json:"previous"`
		Total    int         `json:"total"`
	} `json:"tracks"`
}

type HostDeezerSearchTrack struct {
	Data []struct {
		ID                    int    `json:"id"`
		Readable              bool   `json:"readable"`
		Title                 string `json:"title"`
		TitleShort            string `json:"title_short"`
		TitleVersion          string `json:"title_version"`
		Link                  string `json:"link"`
		Duration              int    `json:"duration"`
		Rank                  int    `json:"rank"`
		ExplicitLyrics        bool   `json:"explicit_lyrics"`
		ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
		ExplicitContentCover  int    `json:"explicit_content_cover"`
		Preview               string `json:"preview"`
		Md5Image              string `json:"md5_image"`
		Artist                struct {
			ID            int    `json:"id"`
			Name          string `json:"name"`
			Link          string `json:"link"`
			Picture       string `json:"picture"`
			PictureSmall  string `json:"picture_small"`
			PictureMedium string `json:"picture_medium"`
			PictureBig    string `json:"picture_big"`
			PictureXl     string `json:"picture_xl"`
			Tracklist     string `json:"tracklist"`
			Type          string `json:"type"`
		} `json:"artist"`
		Album struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Cover       string `json:"cover"`
			CoverSmall  string `json:"cover_small"`
			CoverMedium string `json:"cover_medium"`
			CoverBig    string `json:"cover_big"`
			CoverXl     string `json:"cover_xl"`
			Md5Image    string `json:"md5_image"`
			Tracklist   string `json:"tracklist"`
			Type        string `json:"type"`
		} `json:"album"`
		Type string `json:"type"`
	} `json:"data"`
	Total int `json:"total"`
}

type HostDeezerUserProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	Avatar    string `json:"picture,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Gender    string `json:"gender,omitempty"`
	Country   string `json:"country,omitempty"`
	Language  string `json:"lang,omitempty"`
	Playlist  string `json:"tracklist,omitempty"`
	// Explicit  bool   `json:"lang"`
	// TODO: handle explicit content. Later...
	// Token    string `json:"token"`
}

type UserProfile struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type Token struct {
	jwt.StandardClaims
	Platform      string
	PlatformID    string
	PlatformToken string
	UUID          string
}
