// Main interface for blackbeard providers

package blackbeard

type Metadata struct {
	Description  string
	ThumbnailUrl string
}

type Episode struct {
	Title    string
	Number   int
	Url      string
	Video    Video
	Metadata Metadata
}

type Show struct {
	Title    string
	Url      string
	Episodes []Episode
	Metadata Metadata
}

type Video struct {
	Name     string
	Format   string
	Request  Request
	Metadata Metadata
}

type VideoProvider interface {
	SearchShows(string) []Show
	GetEpisodes(*Show, string) []Episode
	GetVideo(*Episode) Video
}

type Request struct {
	Url     string
	Method  string
	Headers map[string]string
	Body    map[string]string
	Curl    bool
}
