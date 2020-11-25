package types

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

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
	PlayedAt    string   `json:"played_at,omitempty"` // this is because this struct is also used for the single listening history object which contains (and needs) a "when was it played" body which is this.
	AddedAt     string   `json:"added_at,omitempty"`  // similar situation above but in this case, its for Playlists. To know when a track was added to a playlist.
	Album       string   `json:"album"`
}
type Playlist struct {
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Duration      int           `json:"duration"`
	Collaborative bool          `json:"public"`
	TracksNumber  int           `json:"tracks_number"`
	Owner         PlaylistOwner `json:"owner"`
	Tracks        []SingleTrack `json:"tracks"`
	URL           string        `json:"playlist_url"`
	Cover         string        `json:"playlist_cover"`
}

type HostSpotifyCreatePlaylist struct {
	Name string `json:"name"`
}

type PlaylistOwner struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Avatar string `json:"avatar"`
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
	Type string
}

type TrackToSearch struct {
	Title   string
	Artiste string
}

type NewPlaylist struct {
	Title           string   `json:"title"`
	IsPublic        bool     `json:"is_public"`
	IsCollaborative bool     `json:"is_collab"`
	Payload         []string `json:"tracks"`
}

type DeezerPlaylistCreationResponse struct {
	ID int64 `json:"id"`
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
type HostDeezerRawUserProfile struct {
	ID                             int      `json:"id"`
	Name                           string   `json:"name"`
	Lastname                       string   `json:"lastname"`
	Firstname                      string   `json:"firstname"`
	Email                          string   `json:"email"`
	Status                         int      `json:"status"`
	Birthday                       string   `json:"birthday"`
	InscriptionDate                string   `json:"inscription_date"`
	Gender                         string   `json:"gender"`
	Link                           string   `json:"link"`
	Picture                        string   `json:"picture"`
	PictureSmall                   string   `json:"picture_small"`
	PictureMedium                  string   `json:"picture_medium"`
	PictureBig                     string   `json:"picture_big"`
	PictureXl                      string   `json:"picture_xl"`
	Country                        string   `json:"country"`
	Lang                           string   `json:"lang"`
	IsKid                          bool     `json:"is_kid"`
	ExplicitContentLevel           string   `json:"explicit_content_level"`
	ExplicitContentLevelsAvailable []string `json:"explicit_content_levels_available"`
	Tracklist                      string   `json:"tracklist"`
	Type                           string   `json:"type"`
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
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Country   string `json:"country"`
	Lang      string `json:"lang"`
	UUID      string `json:"uuid,omitempty"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Platform  string `json:"platform"`
	Token     string `json:"platform_token"`
}

type NewUser struct {
	ID         int    `json:"id,omitempty"`
	Email      string `json:"email"`
	FirstName  string `json:"firstname"`
	LastName   string `json:"lastname"`
	Country    string `json:"country"`
	Lang       string `json:"lang"`
	UUID       string `json:"uuid,omitempty"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	Platform   string `json:"platform"`
	Token      string `json:"platform_token"`
	Plan       string `json:"platform_plan"`
	PlatformID string `json:"platform_id"`
}
type Token struct {
	jwt.StandardClaims
	Platform      string
	PlatformID    string
	PlatformToken string
	UUID          string
}

type UserProfileUpdate struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

type HostDeezerHistory struct {
	Data []struct {
		ID                    int    `json:"id"`
		Readable              bool   `json:"readable"`
		Title                 string `json:"title"`
		TitleShort            string `json:"title_short"`
		TitleVersion          string `json:"title_version,omitempty"`
		Link                  string `json:"link"`
		Duration              int    `json:"duration"`
		Rank                  int    `json:"rank"`
		ExplicitLyrics        bool   `json:"explicit_lyrics"`
		ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
		ExplicitContentCover  int    `json:"explicit_content_cover"`
		Preview               string `json:"preview"`
		Md5Image              string `json:"md5_image"`
		Timestamp             int    `json:"timestamp"`
		Artist                struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Link      string `json:"link"`
			Tracklist string `json:"tracklist"`
			Type      string `json:"type"`
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
			Tracklist   string `json:"tracklist"`
			Type        string `json:"type"`
		} `json:"album"`
		Type string `json:"type"`
	} `json:"data"`
	Total int    `json:"total"`
	Next  string `json:"next"`
}

type HostSpotifyHistory struct {
	Items []struct {
		Track struct {
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
		} `json:"track"`
		PlayedAt time.Time `json:"played_at"`
		Context  struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"context"`
	} `json:"items"`
	Next    string `json:"next"`
	Cursors struct {
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursors"`
	Limit int    `json:"limit"`
	Href  string `json:"href"`
}

type HostSpotifyAccessTokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
}

type HostDeezerPlaylistResponse struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Duration      int    `json:"duration"`
	Public        bool   `json:"public"`
	IsLovedTrack  bool   `json:"is_loved_track"`
	Collaborative bool   `json:"collaborative"`
	NbTracks      int    `json:"nb_tracks"`
	Fans          int    `json:"fans"`
	Link          string `json:"link"`
	Share         string `json:"share"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Checksum      string `json:"checksum"`
	Tracklist     string `json:"tracklist"`
	CreationDate  string `json:"creation_date"`
	Md5Image      string `json:"md5_image"`
	Creator       struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Tracklist string `json:"tracklist"`
		Type      string `json:"type"`
	} `json:"creator"`
	Type   string `json:"type"`
	Tracks struct {
		Data []struct {
			ID                    int    `json:"id"`
			Readable              bool   `json:"readable"`
			Title                 string `json:"title"`
			TitleShort            string `json:"title_short"`
			TitleVersion          string `json:"title_version,omitempty"`
			Link                  string `json:"link"`
			Duration              int    `json:"duration"`
			Rank                  int    `json:"rank"`
			ExplicitLyrics        bool   `json:"explicit_lyrics"`
			ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
			ExplicitContentCover  int    `json:"explicit_content_cover"`
			Preview               string `json:"preview"`
			Md5Image              string `json:"md5_image"`
			TimeAdd               int    `json:"time_add"`
			Artist                struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				Link      string `json:"link"`
				Tracklist string `json:"tracklist"`
				Type      string `json:"type"`
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
		Checksum string `json:"checksum"`
	} `json:"tracks"`
}

type HostSpotifyNewPlaylistCreationResponse struct {
	Collaborative bool        `json:"collaborative"`
	Description   interface{} `json:"description"`
	ExternalUrls  struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Followers struct {
		Href  interface{} `json:"href"`
		Total int         `json:"total"`
	} `json:"followers"`
	Href   string        `json:"href"`
	ID     string        `json:"id"`
	Images []interface{} `json:"images"`
	Name   string        `json:"name"`
	Owner  struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"owner"`
	Public     bool   `json:"public"`
	SnapshotID string `json:"snapshot_id"`
	Tracks     struct {
		Href     string        `json:"href"`
		Items    []interface{} `json:"items"`
		Limit    int           `json:"limit"`
		Next     interface{}   `json:"next"`
		Offset   int           `json:"offset"`
		Previous interface{}   `json:"previous"`
		Total    int           `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type HostSpotifyAddNewPlaylistTracksResponse struct {
	SnapshotID string `json:"snapshot_id"`
}

type NewSpotifyPlaylist struct {
	Name string `json:"name"`
}
