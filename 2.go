package scdp

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	c "github.com/chromedp/chromedp"
	"github.com/syncfuture/go/serr"
	log "github.com/syncfuture/go/slog"
	"golang.org/x/net/context"
)

func BuildActions(ctx context.Context, actions ...c.Action) []c.Action {
	proxyNeedAuth, _ := ctx.Value(Ctx_ProxyNeedAuth).(bool)
	if proxyNeedAuth {
		actions = append([]c.Action{fetch.Enable().WithHandleAuthRequests(true)}, actions...)
	}

	return actions
}

func Run(ctx context.Context, actions ...c.Action) error {
	actions = BuildActions(ctx, actions...)

	err := c.Run(ctx, actions...)
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}

func GetText(ctx context.Context, selector string, opts ...c.QueryOption) (r string) {
	// opts = append(opts,
	// 	// c.ByQuery,
	// 	c.AtLeast(0),
	// )

	if err := Run(ctx, c.Text(selector, &r, opts...)); err != nil {
		log.Debug(err.Error())
	}

	return
}

func GetNodes(ctx context.Context, selector string, opts ...c.QueryOption) (r []*cdp.Node) {
	// opts = append(opts,
	// 	// c.ByQuery,
	// 	c.AtLeast(0),
	// )

	if err := Run(ctx, c.Nodes(selector, &r, opts...)); err != nil {
		log.Debug(err.Error())
	}

	return
}

func Click(ctx context.Context, selector string, opts ...c.QueryOption) {
	// opts = append(opts,
	// 	// c.ByQuery,
	// 	c.AtLeast(0),
	// )

	if err := Run(ctx, c.Click(selector, opts...)); err != nil {
		log.Debug(err.Error())
	}

	return
}
