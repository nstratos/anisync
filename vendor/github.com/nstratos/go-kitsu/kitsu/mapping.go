package kitsu

const (
	ExternalSiteAniDB      = "anidb"
	ExternalSiteMALAnime   = "myanimelist/anime"
	ExternalSiteMALManga   = "myanimelist/manga"
	ExternalSiteTVDBSeason = "thetvdb/season"
	ExternalSiteTVDBSeries = "thetvdb/series"
)

type Mapping struct {
	ID string `jsonapi:"primary,mappings"`

	// --- Attributes ---

	// ISO 8601 date and time, e.g. 2017-07-27T22:21:26.824Z
	CreatedAt string `jsonapi:"attr,createdAt,omitempty"`

	// ISO 8601 of last modification, e.g. 2017-07-27T22:47:45.129Z
	UpdatedAt string `jsonapi:"attr,updatedAt,omitempty"`

	ExternalSite string `jsonapi:"attr,externalSite,omitempty"`

	ExternalID string `jsonapi:"attr,externalId,omitempty"`
}
