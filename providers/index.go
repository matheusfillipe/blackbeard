// Providers registration

package providers

import (
	"blackbeard/blb"
)

func GetProviders() map[string]blackbeard.VideoProvider {
	return map[string]blackbeard.VideoProvider{
		"9anime": NineAnime{},
		"wcofun": Wcofun{},
	}
}
