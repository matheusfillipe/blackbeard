// Providers registration

package providers

import (
	"blackbeard/blackbeard"
)

func GetProviders() map[string]blackbeard.VideoProvider {
	return map[string]blackbeard.VideoProvider{
		"9anime": NineAnime{},
	}
}
