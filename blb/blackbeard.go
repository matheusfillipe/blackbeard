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

type ProviderInfo struct {
	Name   string
	Movie bool
}

type VideoProvider interface {
	SearchShows(string) []Show
	GetEpisodes(*Show) []Episode
	GetVideo(*Episode) Video
	Info() ProviderInfo
}

type Request struct {
	Url     string
	Method  string
	Headers map[string]string
	Body    map[string]string
	Curl    bool
}

// Create a new request from an existing one, appending to the url
func (r Request) New(path string) Request {
	return Request{
		Url:     r.Url + path,
		Method:  r.Method,
		Headers: r.Headers,
		Body:    r.Body,
		Curl:    r.Curl,
	}
}
