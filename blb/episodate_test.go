package blackbeard

import (
	"fmt"
	"testing"
)

func TestEpisodate(t *testing.T) {
	t.Run("Show info", func(t *testing.T) {
		show := Show{Title: "Attack on titan"}
		EpisodatePopulateShowMetadata(&show)
		if show.Metadata.Description == "" {
			t.Error("No show description returned")
		}

		show = Show{Title: "k on"}
		EpisodatePopulateShowMetadata(&show)
		if show.Metadata.Description == "" {
			t.Error("No show description returned")
		}
	})
	t.Run("Episodes info", func(t *testing.T) {
		show := Show{Title: "Attack on titan"}
		for i := 0; i < 80; i++ {
			show.Episodes = append(show.Episodes, Episode{Title: fmt.Sprintf("%d", i)})
		}
		EpisodatePopulateEpisodesMetadata(&show)
		if show.Episodes[0].Metadata.Description == "" {
			t.Error("No show description returned")
		}

		show = Show{Title: "k on"}
		for i := 0; i < 80; i++ {
			show.Episodes = append(show.Episodes, Episode{Title: fmt.Sprintf("%d", i)})
		}
		EpisodatePopulateEpisodesMetadata(&show)
		if show.Episodes[0].Metadata.Description == "" {
			t.Error("No show description returned")
		}
	})
}
