package types

type DocumentProperties struct {
	Title string `json:"title"`

	// papers
	Author string `json:"author,omitempty"`
	Year   string `json:"year,omitempty"`
	Source string `json:"source,omitempty"`

	// content
	Abstract string   `json:"abstract,omitempty"`
	Notes    string   `json:"notes,omitempty"`
	Keywords []string `json:"keywords,omitempty"`

	// web
	URL         string `json:"url,omitempty"`
	HeaderImage string `json:"headerImage,omitempty"`

	Unread    bool  `json:"unread"`
	Marked    bool  `json:"marked"`
	PublishAt int64 `json:"publishAt,omitempty"`
}
