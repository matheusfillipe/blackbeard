// Entry point, argparsing for either CLI or API

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	blb "github.com/matheusfillipe/blackbeard/blb"
	"github.com/matheusfillipe/blackbeard/providers"
	"github.com/matheusfillipe/go-fuzzyfinder"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/term"
)

var Version = "development"
var BuildDate = "development"

const DEFAULT_PORT = 8080

var cliOpts = struct {
	provider     *string
	show         *int
	episode      *int
	all          *bool
	xnum         *int
	list         *bool
	search       *string
	downloadPath *string
	num          *bool
}{}

func completer(d prompt.Document, provider string) []prompt.Suggest {
	// TODO create show cache
	previousSearches, ok := getSearchCache(provider)
	if !ok {
		return []prompt.Suggest{}
	}
	s := []prompt.Suggest{}
	for _, search := range previousSearches {
		s = append(s, prompt.Suggest{Text: search.Query, Description: search.Description})
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// This is the flow that both the api and local client go through for downloading a video.
// Flow as "a sequence of basic steps"
type TuiFlowTemplate interface {
	getProviders() map[string]blb.VideoProvider
	setProvider(provider blb.VideoProvider, name string)
	searchShows(t string) []blb.Show
	getEpisodes(show blb.Show) []blb.Episode
	getVideo(episode blb.Episode) blb.Video
}

type localFlow struct {
	provider blb.VideoProvider
}

func (flow localFlow) getProviders() map[string]blb.VideoProvider {
	return providers.GetProviders()
}

func (flow *localFlow) setProvider(provider blb.VideoProvider, name string) {
	flow.provider = provider
}

func (flow localFlow) searchShows(t string) []blb.Show {
	return flow.provider.SearchShows(t)
}

func (flow localFlow) getEpisodes(show blb.Show) []blb.Episode {
	return flow.provider.GetEpisodes(&show)
}

func (flow localFlow) getVideo(episode blb.Episode) blb.Video {
	return flow.provider.GetVideo(&episode)
}

type apiFlow struct {
	provider    blb.VideoProvider
	baseRequest blb.Request
}

type apiProvider struct {
	Name        string
	BaseRequest blb.Request
	info        blb.ProviderInfo
}

func (a apiProvider) Info() blb.ProviderInfo {
	return a.info
}

func (a apiProvider) SearchShows(query string) []blb.Show {
	path := fmt.Sprintf("search?provider=%s&q=%s", a.Name, url.QueryEscape(query))
	data := struct {
		Shows []blb.Show `json:"shows"`
	}{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data.Shows
}

func (a apiProvider) GetEpisodes(show *blb.Show) []blb.Episode {
	showurl := show.Url
	path := fmt.Sprintf("episodes?provider=%s&showurl=%s", a.Name, url.QueryEscape(showurl))
	data := struct {
		Episodes []blb.Episode `json:"episodes"`
	}{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data.Episodes
}

func (a apiProvider) GetVideo(episode *blb.Episode) blb.Video {
	epurl := episode.Url
	path := fmt.Sprintf("video?provider=%s&epurl=%s", a.Name, url.QueryEscape(epurl))
	data := blb.Video{}
	blb.GetJson(a.BaseRequest.New(path), &data)
	return data
}

func (flow apiFlow) getProviders() map[string]blb.VideoProvider {
	type Res struct {
		Name string
		Info blb.ProviderInfo
	}
	providers := struct {
		Providers []Res `json:"providers"`
	}{}
	request := flow.baseRequest.New("providers")
	blb.GetJson(request, &providers)

	resp := make(map[string]blb.VideoProvider)
	for _, res := range providers.Providers {
		resp[res.Name] = apiProvider{info: res.Info}
	}
	return resp
}

func (flow *apiFlow) setProvider(provider blb.VideoProvider, name string) {
	prov := provider.(apiProvider)
	prov.BaseRequest = flow.baseRequest
	prov.Name = name
	prov.info = provider.Info()
	flow.provider = prov
}

func (flow apiFlow) searchShows(t string) []blb.Show {
	return flow.provider.SearchShows(t)
}

func (flow apiFlow) getEpisodes(show blb.Show) []blb.Episode {
	return flow.provider.GetEpisodes(&show)
}

func (flow apiFlow) getVideo(episode blb.Episode) blb.Video {
	return flow.provider.GetVideo(&episode)
}

// Workaround for https://github.com/c-bata/go-prompt/issues/233
// TODO fork the lib with the patch or remove this when https://github.com/c-bata/go-prompt/pull/239
// is merged
var termState *term.State

func saveTermState() {
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		return
	}
	termState = oldState
}

func restoreTermState() {
	if termState != nil {
		term.Restore(int(os.Stdin.Fd()), termState)
	}
}

////////////////////////////////////////////////////////////////////////////////

// Choose provider --> Search --> Choose show --> choose episode if any
func downloadTuiFlow(flow TuiFlowTemplate) {
	saveTermState()
	defer restoreTermState()

	providers := flow.getProviders()
	providerNames := blb.Keys(providers)

	if len(providers) == 0 {
		log.Fatal("No providers")
	}

	// If not provider is specified and -list is passed
	if !blb.IsDefault(*cliOpts.list) && blb.IsDefault(*cliOpts.provider) {
		// List providers
		for _, name := range providerNames {
			fmt.Println(name)
		}
		return
	}

	var idx int
	var err error
	if !blb.IsDefault(*cliOpts.provider) {
		if !blb.Contains(providerNames, *cliOpts.provider) {
			fmt.Printf("Provider %s not found, available providers are:\n", *cliOpts.provider)
			for _, name := range providerNames {
				fmt.Printf("%s\n", name)
			}
			return
		}
		idx = blb.IndexOf(providerNames, *cliOpts.provider)
	} else {
		idx, err = fuzzyfinder.Find(
			providerNames,
			func(i int) string {
				return providerNames[i]
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				// Give some safety margin
				w = w / 4

				provider := providers[providerNames[i]]
				name := provider.Info().Name
				if name == "" {
					name = providerNames[i]
				}

				return fmt.Sprintf("%s\n%s\n\n%s",
					strings.ToUpper(name),
					blb.WrapString(provider.Info().Description, uint(w)),
					blb.WrapStringReguardlessly(provider.Info().Url, w),
				)
			}))
	}

	if err != nil {
		log.Fatal(err)
	}

	providerName := providerNames[idx]
	flow.setProvider(providers[providerName], providerName)

	var search string
	if !blb.IsDefault(*cliOpts.search) {
		search = *cliOpts.search
	} else {
		fmt.Println("Search show/anime/movie (hit tab for history autocomplete, C-D to cancel): ")
		search = prompt.Input("> ", func(d prompt.Document) []prompt.Suggest { return completer(d, providerName) })
		restoreTermState()
		if search == "" {
			log.Fatal("No search query")
		}
	}

	shows := flow.searchShows(search)

	if len(shows) == 0 {
		if providers[providerName].Info().Cloudflared {
			fmt.Println("You might want to install curl impersonate or use the api: https://github.com/matheusfillipe/blackbeard#usage")
		}
		log.Fatal("No shows/movies found")
	}
	// If list is passed display shows, don't cache
	if !blb.IsDefault(*cliOpts.list) && blb.IsDefault(*cliOpts.show) {
		for i, show := range shows {
			fmt.Printf("%d > %s > %s\n", i+1, show.Title, show.Url)
		}
		return
	}
	writeSearchCache(providerName, search)

	// Choose show
	if !blb.IsDefault(*cliOpts.show) && *cliOpts.show > 0 {
		idx = *cliOpts.show - 1
	} else {
		idx, err = fuzzyfinder.Find(
			shows,
			func(i int) string {
				return shows[i].Title
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}
				// Give some safety margin
				w = w/2 - 12
				return fmt.Sprintf("Provider: %s\nShow: %s\n\nDescription: %s\n\n\n%s",
					strings.ToUpper(providerName),
					blb.WrapString(shows[i].Title, uint(w)),
					blb.WrapString(shows[i].Metadata.Description, uint(w)),
					blb.WrapStringReguardlessly(shows[i].Metadata.ThumbnailUrl, w),
				)
			}))
	}

	if err != nil {
		log.Fatal(err)
	}

	// Check if idx out of range
	if idx >= len(shows) {
		fmt.Println("Show index out of range. Found Shows were:")
		for i, show := range shows {
			fmt.Printf("%d > %s > %s\n", i+1, show.Title, show.Url)
		}
		return
	}
	show := shows[idx]
	episodes := flow.getEpisodes(show)

	if len(episodes) == 0 {
		log.Fatal("No episodes found")
	}

	// Passed list but didn't specify episode
	if !blb.IsDefault(*cliOpts.list) && blb.IsDefault(*cliOpts.episode) {
		for i, episode := range episodes {
			fmt.Printf("%d > %s > %s\n", i+1, episode.Title, episode.Url)
		}
		return
	}

	var indexes []int

	// if show is movie we can skip the episode list
	if show.IsMovie {
		indexes = []int{0}
	} else {
		// Show has episodes
		if !blb.IsDefault(*cliOpts.episode) {
			indexes = []int{*cliOpts.episode - 1}
		} else if !blb.IsDefault(*cliOpts.all) {
			indexes = []int{}
			for i := 0; i < len(episodes); i++ {
				indexes = append(indexes, i)
			}
		} else {
			idxs, err2 := fuzzyfinder.FindMulti(
				episodes,
				func(i int) string {
					return fmt.Sprintf("%v > %v", episodes[i].Number+1, episodes[i].Title)
				},
				fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
					if i == -1 {
						return ""
					}
					// Give some safety margin
					w = w/2 - 12
					return fmt.Sprintf("Provider: %s\nShow: %s\nEpisode n. %d\n\nDescription: %s",
						strings.ToUpper(providerName),
						blb.WrapString(show.Title, uint(w)),
						episodes[i].Number+1,
						blb.WrapString(episodes[i].Metadata.Description, uint(w)),
					)
				}))

			if err2 != nil {
				log.Fatal(err)
			}
			indexes = idxs
		}
	}

	sort.Slice(indexes, func(i, j int) bool { return indexes[i] < indexes[j] })

	// Create dir if multiple
	dir := "./"
	if len(indexes) > 1 && blb.IsDefault(*cliOpts.downloadPath) {
		err = os.MkdirAll(blb.SanitizeFilename(show.Title), 0755)
		if err != nil {
			log.Fatal(err)
		}
		dir = blb.SanitizeFilename(show.Title)
	}
	if !blb.IsDefault(*cliOpts.downloadPath) {
		dir = blb.SanitizeFilename(*cliOpts.downloadPath)
	}

	// Download all episodes in parallel
	maxConcurrency := 1
	if !blb.IsDefault(*cliOpts.xnum) {
		maxConcurrency = *cliOpts.xnum
	}
	var throttle = make(chan int, maxConcurrency)
	var wg sync.WaitGroup

	// Formatting options
	var n_episodes = strconv.Itoa(len(strconv.Itoa(len(episodes))))
	var formatstr = "%0" + n_episodes + "d-"
	if !blb.IsDefault(*cliOpts.num) {
		formatstr = ""
	}

	// clear screen
	fmt.Print("\033[H\033[2J")

	// Create space for download lines
	fmt.Print(blb.Repeat("\n", len(indexes)+1))

	// Go back up
	fmt.Print(blb.Repeat("\033[1A", len(indexes)+1))

	for _, idx := range indexes {
		fmt.Println("")
		throttle <- 1
		wg.Add(1)
		go func(idx int, wg *sync.WaitGroup, throttle chan int) {
			defer wg.Done()
			defer func() { <-throttle; fmt.Println("") }()
			episode := episodes[idx]
			video := flow.getVideo(episode)
			switch video.Format {
			case "mp4":
				if !video.Download(dir, idx, formatstr) {
					fmt.Printf("Failed to download %s", video.Name)
					fmt.Printf(blb.Repeat("\n", maxConcurrency+1))
				}
				break
			default:
				fmt.Printf("Download not implemenetd for format: %s\nURL: %s", video.Format, video.Request.Url)
			}
		}(idx, &wg, throttle)
	}
	wg.Wait()
	fmt.Println("\n\n                        All Done!!!")
}

func apiConnect(url string) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	fmt.Printf("Attempting connection to blackbeard api at %q\n", url)

	// Check if there is a valid reply
	res, ok := blb.Timeout(10, func() *http.Response {
		request := blb.Request{Url: url + "version"}
		res, ok := blb.Perform(request)
		if !ok {
			log.Fatal("Connection failed")
		}
		return res
	})
	if !ok {
		log.Fatal("Connection timed out")
	}

	body := res.Body
	defer body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)

	if strings.Contains(buf.String(), "version") {
		fmt.Println("Connection successful")
	} else {
		log.Fatal("Connection failed")
	}

	flow := apiFlow{}
	flow.baseRequest = blb.Request{Url: url}
	downloadTuiFlow(&flow)
}

func main() {
	defaultPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		defaultPort = DEFAULT_PORT
	}

	username := "default"
	user, err := user.Current()
	if err == nil {
		username = blb.SanitizeFilename(user.Username)
	}

	const default_host = "0.0.0.0:8080"

	// API opts
	apiMode := flag.Bool("api", false, "Start a blackbeard api.")
	apiPort := flag.Int("port", defaultPort, "Port to bind to if api. Will also read 'PORT' from env.")
	apiHost := flag.String("host", "0.0.0.0", "Host to bind to if api.")

	// Client opts
	connectAddr := flag.String("connect", "", "Start a client that connects to a blackbeard api with the given address.")
	profileName := flag.String("profile", username, "Use a different profile folder.")
	cliOpts.provider = flag.String("provider", "", "Use a provider.")
	cliOpts.show = flag.Int("show", 0, "Choose a show/movie directly by Number.")
	cliOpts.episode = flag.Int("ep", 0, "Choose an episode number to download. Both -show and -ep start from 1.")
	cliOpts.list = flag.Bool("list", false, "List line separated and stop execution. Can be used alone to list providers, with search to list show results or with show to list episodes. This is the only option that will prevent from downloading anything and just output data.")
	cliOpts.search = flag.String("search", "", "Searches for show/movie.")
	cliOpts.all = flag.Bool("D", false, "Download all episodes.")
	cliOpts.xnum = flag.Int("x", 0, "Number of parallel download workers")
	cliOpts.downloadPath = flag.String("path", "", "Directory to save to")
	cliOpts.num = flag.Bool("n", false, "Supresss prepending the episode number to the filename")

	version := flag.Bool("version", false, "Prints the version then exits")

	flag.Parse()

	if *version {
		fmt.Println("Blackbeard")
		fmt.Println("Version: ", Version)
		fmt.Println("Date: ", BuildDate)
		return
	}

	if *apiMode && *connectAddr != "" {
		log.Fatal("Cannot start api and client at the same time")
		return
	}

	if *apiMode {
		startApiServer(*apiHost, *apiPort)
		return
	}

	createCacheDir(*profileName)
	if *connectAddr != "" {
		apiConnect(*connectAddr)
		return
	}

	// Interactive cli
	downloadTuiFlow(&localFlow{})
}
