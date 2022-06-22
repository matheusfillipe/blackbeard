// Headless browser helpers

package blackbeard

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func Headless() {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var page string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.2embed.ru/embed/tmdb/tv?id=1429&s=4&e=1`),
		// wait for footer element is visible (ie, page is loaded)
		chromedp.WaitVisible(`div.loading-play`),
		// find and click "Example" link
		chromedp.Click(`div.loading-play`, chromedp.NodeVisible),
		// retrieve the text of the textarea
		// chromedp.Value(`body`, &page),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(">>>>>\n%s", page)
}
