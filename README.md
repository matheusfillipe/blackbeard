[![CircleCI Build Status](https://circleci.com/gh/matheusfillipe/blackbeard.svg?style=shield)](https://circleci.com/gh/matheusfillipe/blackbeard)
[![Heroku](https://heroku-badge.herokuapp.com/?app=blackbeardapi)](https://blackbeardapi.herokuapp.com)


# blackbeard

Api and cli that scrapes content from video providers.

You need [curl-impersonate](https://github.com/lwthiker/curl-impersonate) for being able to run the wcofun provider using: `export LD_PRELOAD=/usr/lib/libcurl-impersonate-chrome.so`

Similar projects:
- https://github.com/anime-dl
- https://github.com/LagradOst/CloudStream-3

# Installation

Make sure `$(go env GOPATH)` is on your `PATH` and run:

``` sh
go install github.com/matheusfillipe/blackbeard@latest
```

Then check the help with `blackbeard -h`

To get cloudflare blocked providers to work you need [curl-impersonate](https://github.com/lwthiker/curl-impersonate). On archlinux you can: `yay -S curl-impersonate-chrome`.

After curl-impersonate is installed, run it like:

`LD_PRELOAD=/usr/local/lib/libcurl-impersonate-chrome.so blackbeard`


# Usage

You can run or connect to an api using either `-api` or `-connect`. You can also adjust `-port` and `-host`. If you just run it without arguments you will have an interactive fuzzy interface similar to fzf that will let you download whatever episode. You can mark multiple episodes by hitting or holding `TAB`.

You can test this without having to compile curl-impersonate using the heroku api:

``` sh
blackbeard -connect https://blackbeard.fly.dev/
```

# Example usage

**List providers**

``` sh
blackbeard -list
```



**Search and output to stdout**

``` sh
blackbeard -provider wcofun -search "attack on titan" -list
```

If you remove the `-list` it will go to the fzf prompt and proceed to download

**List episodes for a show**

``` sh
blackbeard -provider wcofun -search "attack on titan" -show 1 -list
```


**Download an ep directly, no prompt**

``` sh
blackbeard -provider wcofun -search "attack on titan" -show 1 -ep 1
```


**Download all episodes with 4 downloads in parallel using the api to scrape**

``` sh
blackbeard -provider wcofun -search "attack on titan" -show 1 -D -x 4 -connect https://blackbeardapi.herokuapp.com/
```



# Disclaimer

The app is purely for educational and personal use. It merely scrapes 3rd-party websites that are publicly accessible via any regular web browser. It is the responsibility of user to avoid any actions that might violate the laws governing his/her locality. Use this at your own risk.


# TODO

- [x] Cli arg parsing
- [x] http api
- [x] Profiles that store watch history
- [x] wcofun provider
- [ ] m3u8 downloader https://github.com/canhlinh/hlsdl
- [ ] headless browser
    - [ ] 9anime provider
    - [ ] soap2day provider
- [ ] More providers
