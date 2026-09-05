package models

import "time"

// DashboardPreferences controls how the installed applications dashboard is presented.
type DashboardPreferences struct {
	DefaultView    string          `json:"defaultView"`
	TileDensity    string          `json:"tileDensity"`
	TileSize       string          `json:"tileSize"`
	GroupBy        string          `json:"groupBy"`
	SortBy         string          `json:"sortBy"`
	TileOrder      []string        `json:"tileOrder"`
	FavoriteAppIDs []string        `json:"favoriteAppIds"`
	HiddenFields   map[string]bool `json:"hiddenFields"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

func DefaultDashboardPreferences() *DashboardPreferences {
	return &DashboardPreferences{
		DefaultView: "grid", TileDensity: "normal", TileSize: "medium",
		GroupBy: "none", SortBy: "manual", TileOrder: []string{},
		FavoriteAppIDs: []string{}, HiddenFields: map[string]bool{},
	}
}
