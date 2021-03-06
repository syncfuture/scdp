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
	cd "github.com/chromedp/chromedp"
	"github.com/syncfuture/go/serr"
	log "github.com/syncfuture/go/slog"
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

func createDebugingBrowser(exePath string, debugPort int, args ...string) (error, string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	portOpen := isPortOpen(debugPort)

	if !portOpen {
		log.Info("Starting chrome...")

		args = append([]string{fmt.Sprintf("--remote-debugging-port=%d", debugPort)}, args...)
		cmd := exec.Command(exePath, args...)
		err := cmd.Start()
		if err != nil {
			return err, ""
		}

		portOpen = isPortOpen(debugPort)
		for !portOpen {
			portOpen = isPortOpen(debugPort)

			// 超时退出
			select {
			case <-ctx.Done():
				return serr.New("start browser timeout"), ""
			default:
				break
			}
		}
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/json/version", debugPort))
	if err != nil {
		return err, ""
	}
	defer resp.Body.Close()
	configJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, ""
	}

	config := make(map[string]string)
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		return err, ""
	}

	webSocketDebuggerURL := config["webSocketDebuggerUrl"]
	if webSocketDebuggerURL == "" {
		log.Fatal("get webSocketDebuggerUrl failed")
	}
	log.Debug("debug url: ", webSocketDebuggerURL)
	return nil, webSocketDebuggerURL
}

func createDebugingContext(debuggerURL string) *TabContext {
	cancels := make([]context.CancelFunc, 0, 3)
	ctx := context.Background()
	timeoutCtx, cancel1 := context.WithTimeout(ctx, time.Second*30)
	cancels = append([]context.CancelFunc{cancel1}, cancels...)

	allocCtx, cancel2 := cd.NewRemoteAllocator(timeoutCtx, debuggerURL)
	cancels = append([]context.CancelFunc{cancel2}, cancels...)

	taskCtx, cancel3 := cd.NewContext(allocCtx)
	cancels = append([]context.CancelFunc{cancel3}, cancels...)

	return &TabContext{
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
					log.Debug(err)
				}
			}
		}()
	})
}

func ViewPort(width int64, height int64, deviceScaleFactor float64, mobile bool) cd.Action {
	return emulation.SetDeviceMetricsOverride(width, height, deviceScaleFactor, mobile)
}
