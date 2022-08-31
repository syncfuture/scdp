package scdp

import (
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	cd "github.com/chromedp/chromedp"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/slog"
	"golang.org/x/net/context"
)

func BuildActions(ctx context.Context, actions ...cd.Action) []cd.Action {
	proxyNeedAuth, _ := ctx.Value(CTX_ProxyNeedAuth).(bool)
	if proxyNeedAuth {
		actions = append([]cd.Action{fetch.Enable().WithHandleAuthRequests(true)}, actions...)
	}

	return actions
}

func Run(ctx context.Context, actions ...cd.Action) error {
	actions = BuildActions(ctx, actions...)

	err := cd.Run(ctx, actions...)
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}

func GetText(ctx context.Context, selector string, opts ...cd.QueryOption) (r string) {
	if err := Run(ctx, cd.Text(selector, &r, opts...)); err != nil {
		slog.Debug(err.Error())
	}

	return
}

func GetInnerHTML(ctx context.Context, selector string, opts ...cd.QueryOption) (r string) {
	if err := Run(ctx, cd.InnerHTML(selector, &r, opts...)); err != nil {
		slog.Debug(err.Error())
	}

	return
}

func GetOuterHTML(ctx context.Context, selector string, opts ...cd.QueryOption) (r string) {
	if err := Run(ctx, cd.OuterHTML(selector, &r, opts...)); err != nil {
		slog.Debug(err.Error())
	}

	return
}

func GetNodes(ctx context.Context, selector string, opts ...cd.QueryOption) (r []*cdp.Node) {
	if err := Run(ctx, cd.Nodes(selector, &r, opts...)); err != nil {
		slog.Debug(err.Error())
	}

	return
}

func Click(ctx context.Context, selector string, opts ...cd.QueryOption) {
	if err := Run(ctx, cd.Click(selector, opts...)); err != nil {
		slog.Debug(err.Error())
	}

	return
}
