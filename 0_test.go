package scdp

import (
	"fmt"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/srand"
	"github.com/syncfuture/go/u"
	"golang.org/x/net/context"
)

const CHOMRE_PATH string = `C:\Users\Lukiya\AppData\Local\Google\Chrome\Application\chrome.exe`

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

func Test_runTestYourSmart(t *testing.T) {
	debuggerURL, err := LaunchDebugingBrowser(CHOMRE_PATH, 9222)
	assert.NoError(t, err)
	tab := CreateDebugingTab(debuggerURL)

	// get the list of the targets
	infos, err := chromedp.Targets(tab.Context)
	u.LogFatal(err)

	tabCtx, cancel := chromedp.NewContext(tab.Context, chromedp.WithTargetID(infos[0].TargetID))
	defer cancel()

	runTestYourSmart(tabCtx, 10)

	slog.Info("Done")
}

func Test_runSupersonicQuiz(t *testing.T) {
	debuggerURL, err := LaunchDebugingBrowser(CHOMRE_PATH, 9222)
	assert.NoError(t, err)
	tab := CreateDebugingTab(debuggerURL)

	// get the list of the targets
	infos, err := chromedp.Targets(tab.Context)
	u.LogFatal(err)

	tabCtx, _ := chromedp.NewContext(tab.Context, chromedp.WithTargetID(infos[0].TargetID))

	runSupersonicQuiz(tabCtx)

	slog.Info("Done")
}

func runTestYourSmart(tabCtx context.Context, max int) (err error) {
	for i := 0; i < max; i++ {
		selQ := fmt.Sprintf("#QuestionPane%d", i)
		selA := fmt.Sprintf("#AnswerPane%d", i)
		slog.Infof("selQ: %s, selA: %s", selQ, selA)
		// selQNext := fmt.Sprintf("#QuestionPane%d", i+1)

		var nodes []*cdp.Node
		err = chromedp.Run(tabCtx,
			chromedp.WaitReady("#ListOfQuestionAndAnswerPanes"),
			chromedp.Nodes("#ListOfQuestionAndAnswerPanes .wk_choicesInstLink", &nodes, chromedp.ByQueryAll),
			chromedp.WaitReady(selQ),
			chromedp.ActionFunc(func(ctx context.Context) error {
				node := nodes[srand.IntRange(0, len(nodes)-1)]

				err = chromedp.Run(tabCtx,
					chromedp.MouseClickNode(node),
					chromedp.WaitVisible(selA+" input"),
					chromedp.Click(selA+" input"),
					// chromedp.WaitReady(selQNext),
				)
				u.LogFatal(err)

				return nil
			}),
		)
		u.LogFatal(err)
	}

	return
}

func runSupersonicQuiz(tabCtx context.Context) (err error) {
	err = chromedp.Run(tabCtx,
		chromedp.WaitReady(".btOptions"),
	)
	u.LogFatal(err)

	for i := 0; i < 5; i++ {
		err = chromedp.Run(tabCtx,
			chromedp.Click(".btOptions .btcc:not(.btsel)", chromedp.AtLeast(1)),
			chromedp.WaitReady(".btOptions .btcc:not(.btsel)", chromedp.AtLeast(1)),
		)
		u.LogFatal(err)
	}

	return
}
