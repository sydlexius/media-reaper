package emby

// SystemInfo represents Emby server system information.
type SystemInfo struct {
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
	ID         string `json:"Id"`
}

// User represents an Emby user.
type User struct {
	ID     string      `json:"Id"`
	Name   string      `json:"Name"`
	Policy *UserPolicy `json:"Policy,omitempty"`
}

// UserPolicy represents an Emby user's access policy.
type UserPolicy struct {
	IsAdministrator  bool     `json:"IsAdministrator"`
	EnableAllFolders bool     `json:"EnableAllFolders"`
	EnabledFolders   []string `json:"EnabledFolders"`
}

// Library represents an Emby media library (folder).
type Library struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	CollectionType string `json:"CollectionType"`
}

// MediaFoldersResponse wraps the response from GET /Library/MediaFolders.
type MediaFoldersResponse struct {
	Items []Library `json:"Items"`
}

// ItemQuery specifies query parameters for fetching user items.
type ItemQuery struct {
	ParentID     string
	IncludeTypes string
	Recursive    bool
	Limit        int
	StartIndex   int
	SortBy       string
	SortOrder    string
	IsPlayed     *bool
	Fields       string
}

// ItemsResult wraps a paginated list of items.
type ItemsResult struct {
	Items            []*Item `json:"Items"`
	TotalRecordCount int     `json:"TotalRecordCount"`
}

// Item represents an Emby media item.
type Item struct {
	ID                string    `json:"Id"`
	Name              string    `json:"Name"`
	Type              string    `json:"Type"`
	SeriesName        string    `json:"SeriesName,omitempty"`
	SeasonName        string    `json:"SeasonName,omitempty"`
	IndexNumber       *int      `json:"IndexNumber,omitempty"`
	ParentIndexNumber *int      `json:"ParentIndexNumber,omitempty"`
	Path              string    `json:"Path,omitempty"`
	ProviderIDs       Providers `json:"ProviderIds,omitempty"`
	UserData          *UserData `json:"UserData,omitempty"`
}

// Providers holds external IDs for an item.
type Providers struct {
	TMDB string `json:"Tmdb,omitempty"`
	TVDB string `json:"Tvdb,omitempty"`
	IMDB string `json:"Imdb,omitempty"`
}

// UserData holds per-user play state for an item.
type UserData struct {
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	PlayCount             int    `json:"PlayCount"`
	IsFavorite            bool   `json:"IsFavorite"`
	Played                bool   `json:"Played"`
	LastPlayedDate        string `json:"LastPlayedDate,omitempty"`
}
