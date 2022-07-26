// Headless browser helpers

package blackbeard

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	_ "github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

func Headless() {
	// Command line args
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			// chromedp.Flag("disable-extensions", false),
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-notifications", true),
			chromedp.Flag("block-new-web-contents", true),
			// chromedp.Flag("disable-popup-blocking", false),
			// chromedp.Flag("load-extension", "/home/matheus/Downloads/uBlock0.chromium"),
		)...)
	defer cancel()

	// create chrome instance
	ctx, cancel = chromedp.NewContext(
		ctx,
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	var buf []byte
	var bufe []byte
	var outerhtml string
	var elm = `#play-now`

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		// listen for the Target.targetCreated event
		if ev, ok := ev.(*target.EventTargetCreated); ok {
			c := chromedp.FromContext(ctx)
			// when the new target is open by the current tab
			if ev.TargetInfo.OpenerID == c.Target.TargetID {
				go func() {
					// close the new target
					// note that the executor should be "c.Browser".
					if err := target.CloseTarget(ev.TargetInfo.TargetID).Do(cdp.WithExecutor(ctx, c.Browser)); err != nil {
						log.Printf("failed to close target: %s, %v", ev.TargetInfo.TargetID, err)
					}
				}()
			}
		}
	})
	// navigate to a page, wait for an element, click
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.2embed.ru/embed/tmdb/tv?id=1429&s=4&e=1`),
		chromedp.WaitVisible(elm),
		chromedp.Sleep(5*time.Second),
		chromedp.Click(elm, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		chromedp.Click(elm, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		// chromedp.Sleep(5*time.Second),
		// chromedp.Click(elm, chromedp.NodeVisible),
		// chromedp.Sleep(20*time.Second),
		// chromedp.Click(elm, chromedp.NodeVisible),
		chromedp.Sleep(10*time.Second),
		// chromedp.Screenshot(elm, &bufe, chromedp.NodeVisible),
		chromedp.OuterHTML("body", &outerhtml, chromedp.ByQuery),
		chromedp.FullScreenshot(&buf, 90),
		// chromedp.Value(`body`, &page),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s/n", outerhtml)

	if err := ioutil.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("selector.png", bufe, 0o644); err != nil {
		log.Fatal(err)
	}

	// log.Printf("wrote fullScreenshot.png")
}
