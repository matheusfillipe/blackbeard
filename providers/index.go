// Providers registration

package providers

import (
	"github.com/matheusfillipe/blackbeard/blb"
)

func GetProviders() map[string]blackbeard.VideoProvider {
	return map[string]blackbeard.VideoProvider{
		"9anime": NineAnime{},
		"wcofun": Wcofun{},
		"1337x": Leetx{},
	}
}
