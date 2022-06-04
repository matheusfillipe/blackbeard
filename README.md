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

`LD_PRELOAD=/usr/lib/libcurl-impersonate-chrome.so blackbeard`


# Usage

You can run or connect to an api using either `-api` or `-connect`. You can also adjust `-port` and `-host`. If you just run it without arguments you will have an interactive fuzzy interface similar to fzf that will let you download whatever episode. You can mark multiple episodes by hitting or holding `TAB`.

# Disclaimer

The app is purely for educational and personal use. It merely scrapes 3rd-party websites that are publicly accessible via any regular web browser. It is the responsibility of user to avoid any actions that might violate the laws governing his/her locality. Use this at your own risk.


# TODO

- [x] Cli arg parsing
- [ ] http api
- [ ] Profiles that store watch history
- [x] wcofun provider
- [ ] 9anime provider
- [ ] More providers
