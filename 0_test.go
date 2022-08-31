package scdp

import (
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"golang.org/x/net/context"
)

func Test1(t *testing.T) {
	options := append(DefaultFlags,
		chromedp.ExecPath(`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`),
		InPrivate(),
		DisableImage(),
	)
	baseContext, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	ctx1, cancel := chromedp.NewContext(baseContext)
	defer cancel()

	var err error
	err = chromedp.Run(ctx1)
	assert.NoError(t, err)
}

func Test2(t *testing.T) {
	const path string = `C:\Users\Lukiya\AppData\Local\Google\Chrome\Application\chrome.exe`
	debuggerURL, err := LaunchDebugingBrowser(path, 9222)
	assert.NoError(t, err)
	tab := CreateDebugingTab(debuggerURL)

	// get the list of the targets
	infos, err := chromedp.Targets(tab.Context)
	u.LogFatal(err)

	tabCtx, cancel := chromedp.NewContext(tab.Context, chromedp.WithTargetID(infos[0].TargetID))
	defer cancel()

	var nodes []*cdp.Node
	err = chromedp.Run(tabCtx,
		chromedp.Nodes(".btOptions .slide .b_cards", &nodes, chromedp.ByQueryAll),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, node := range nodes {
				slog.Info(node.NodeID)

				err = chromedp.Run(tabCtx,
					chromedp.WaitReady(".btOptions"),
					chromedp.MouseClickNode(node),
				)
				u.LogFatal(err)
			}

			return nil
		}),
	)
	u.LogFatal(err)

	slog.Info("Done")
}
