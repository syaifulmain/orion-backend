package corpus

// Corpus represents a single record in the catalog
type Corpus struct {
	ID        string `bson:"_id" json:"id"`
	Text      string `bson:"text" json:"text"`
	Source    string `bson:"source" json:"source"`
	License   string `bson:"license" json:"license"`
	Split     string `bson:"split" json:"split"`
	CrawlDate string `bson:"crawl_date" json:"crawl_date"`
}

// ExportFilterParams defines query parameters for streaming export.
type ExportFilterParams struct {
	// Case-insensitive text & source search (max 200 chars)
	Q string `query:"q" example:"kesehatan" maxLength:"200"`
	// Dataset split partition
	Split string `query:"split" enums:"train,eval,test" example:"eval"`
	// Source origin (exact match or prefix, e.g. youtube)
	Source string `query:"source" example:"youtube"`
	// License type (case-insensitive exact match, e.g. CC-BY-3.0)
	License string `query:"license" example:"CC-BY-3.0"`
}

// ToFilterParams converts ExportFilterParams to FilterParams.
func (e ExportFilterParams) ToFilterParams() FilterParams {
	return FilterParams{
		ExportFilterParams: e,
	}
}

// FilterParams defines query parameters for listing corpus records.
type FilterParams struct {
	ExportFilterParams

	// Items per page (default: 25, max: 1000)
	Limit int `query:"limit" default:"25"`
	// Items to skip (default: 0)
	Offset int `query:"offset" default:"0"`
}
