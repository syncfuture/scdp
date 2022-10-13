package scdp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	cd "github.com/chromedp/chromedp"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
	"golang.org/x/net/context"
)

func isPortOpen(port int) bool {
	_, err := net.DialTimeout("tcp", "localhost:"+strconv.Itoa(port), time.Millisecond*500)
	if err != nil {
		return false
	}
	return true
}

func LaunchDebugingBrowser(exePath string, debugPort int, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	portOpen := isPortOpen(debugPort)

	if !portOpen {
		slog.Info("Starting chrome...")

		args = append([]string{fmt.Sprintf("--remote-debugging-port=%d", debugPort)}, args...)
		cmd := exec.Command(exePath, args...)
		err := cmd.Start()
		if err != nil {
			return "", serr.WithStack(err)
		}

		portOpen = isPortOpen(debugPort)
		for !portOpen {
			portOpen = isPortOpen(debugPort)

			// 超时退出
			select {
			case <-ctx.Done():
				return "", serr.New("start browser timeout")
			default:
				break
			}
		}
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/json/version", debugPort))
	if err != nil {
		return "", serr.WithStack(err)
	}
	defer resp.Body.Close()
	configJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", serr.WithStack(err)
	}

	config := make(map[string]string)
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		return "", serr.WithStack(err)
	}

	webSocketDebuggerURL := config["webSocketDebuggerUrl"]
	if webSocketDebuggerURL == "" {
		slog.Fatal("get webSocketDebuggerUrl failed")
	}
	slog.Debug("debug url: ", webSocketDebuggerURL)
	return webSocketDebuggerURL, nil
}

// Create a tab trough opened debugger url chrome instance
func createDebugingTab(debuggerURL string) *Tab {
	cancels := make([]context.CancelFunc, 0, 3)
	ctx := context.Background()
	timeoutCtx, cancel1 := context.WithTimeout(ctx, time.Second*60)
	cancels = append([]context.CancelFunc{cancel1}, cancels...)

	allocCtx, cancel2 := cd.NewRemoteAllocator(timeoutCtx, debuggerURL)
	cancels = append([]context.CancelFunc{cancel2}, cancels...)

	taskCtx, cancel3 := cd.NewContext(allocCtx)
	cancels = append([]context.CancelFunc{cancel3}, cancels...)

	return &Tab{
		Context: taskCtx,
		Cancels: cancels,
	}
}

func DisableImage() cd.ExecAllocatorOption {
	return cd.Flag("blink-settings", "imagesEnabled=false")
}

func InPrivate() cd.ExecAllocatorOption {
	return cd.Flag("inprivate", true)
}

func Incognito() cd.ExecAllocatorOption {
	return cd.Flag("incognito", true)
}

func ProxyAuth(ctx context.Context, username, password string) {
	cd.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			switch ev := ev.(type) {
			case *fetch.EventAuthRequired:
				c := cd.FromContext(ctx)
				execCtx := cdp.WithExecutor(ctx, c.Target)

				resp := &fetch.AuthChallengeResponse{
					Response: fetch.AuthChallengeResponseResponseProvideCredentials,
					Username: username,
					Password: password,
				}

				err := fetch.ContinueWithAuth(ev.RequestID, resp).Do(execCtx)
				u.LogError(err)

			case *fetch.EventRequestPaused:
				c := cd.FromContext(ctx)
				execCtx := cdp.WithExecutor(ctx, c.Target)
				err := fetch.ContinueRequest(ev.RequestID).Do(execCtx)
				if err != nil {
					slog.Debug(err)
				}
			}
		}()
	})
}

func ViewPort(width int64, height int64, deviceScaleFactor float64, mobile bool) cd.Action {
	return emulation.SetDeviceMetricsOverride(width, height, deviceScaleFactor, mobile)
}

func CreateDebugingContext(browserExePath string, port int) (context.Context, error) {
	debuggerURL, err := LaunchDebugingBrowser(browserExePath, port)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	tab := createDebugingTab(debuggerURL)

	// get the list of the targets
	infos, err := chromedp.Targets(tab.Context)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	tabCtx, _ := chromedp.NewContext(tab.Context, chromedp.WithTargetID(infos[0].TargetID))

	return tabCtx, nil
}
