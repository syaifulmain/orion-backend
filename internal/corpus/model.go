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

// FilterParams defines query parameters for listing corpus records.
type FilterParams struct {
	Q       string `query:"q"`
	Split   string `query:"split"`
	Source  string `query:"source"`
	License string `query:"license"`
	Limit   int    `query:"limit"`
	Offset  int    `query:"offset"`
}
