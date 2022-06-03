// Main interface for blackbeard providers

package blackbeard

type Episode struct {
	Title  string
	Number int
	Url    string
	Video  Video
}

type Show struct {
	Title    string
	Url      string
	Episodes []Episode
}

type Video struct {
	Url     string
	Format  string
	Headers map[string]string
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
