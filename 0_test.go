package scdp

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
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
