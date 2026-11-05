//go:build windows

package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/systray"

	"netra-uptime/config"
	"netra-uptime/internal/app"
	"netra-uptime/pkg/browser"
	"netra-uptime/pkg/logger"
	"netra-uptime/pkg/singleinstance"
)

//go:embed icon.ico
var iconData []byte

const mutexName = "Local\\NetraUptimeSingleInstanceMutex"

func main() {
	noBrowser := flag.Bool("no-browser", false, "Do not auto-open browser on startup")
	background := flag.Bool("background", false, "Run silently in background without opening browser")
	flag.Parse()

	// Single instance check using Windows named mutex before initializing file logger or server
	singleMutex, alreadyRunning, mutexErr := singleinstance.Acquire(mutexName)
	if mutexErr == nil && alreadyRunning {
		// Another instance is already running: resolve configured port, open the UI in browser, and exit cleanly
		cfg := config.Load()
		appURL := fmt.Sprintf("http://127.0.0.1:%s", cfg.Port)
		_ = browser.Open(appURL)
		os.Exit(0)
	}
	defer singleMutex.Release()

	// Redirect logs to netra.log in executable directory before config.Load() so all startup logs are captured
	logCleanup, err := logger.SetupFileLogger()
	if err != nil {
		log.Printf("[Desktop] Warning: failed to initialize file logger: %v", err)
	} else {
		defer logCleanup()
	}

	if mutexErr != nil {
		log.Printf("[Desktop] Warning: single-instance check encountered error: %v", mutexErr)
	}

	cfg := config.Load()
	appURL := fmt.Sprintf("http://127.0.0.1:%s", cfg.Port)

	log.Printf("[Netra Uptime Desktop] Starting desktop service on %s:%s (DB: %s)\n", cfg.Host, cfg.Port, cfg.DBDriver)

	serverApp, err := app.Setup(cfg)
	if err != nil {
		log.Fatalf("[Desktop] Failed to bootstrap application: %v", err)
	}

	go func() {
		if err := serverApp.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Desktop] Server failed to run on %s: %v", serverApp.Server.Addr, err)
		}
	}()

	shouldOpenBrowser := !*noBrowser && !*background

	onReady := func() {
		if len(iconData) > 0 {
			systray.SetIcon(iconData)
		}
		systray.SetTitle("Netra Uptime")
		systray.SetTooltip("Netra Uptime - Monitoring Server")

		systray.SetOnTapped(func() {
			_ = browser.Open(appURL)
		})

		mOpen := systray.AddMenuItem("Open Netra", "Open Netra Uptime dashboard in browser")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Shut down Netra Uptime cleanly")

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					_ = browser.Open(appURL)
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()

		if shouldOpenBrowser {
			_ = browser.Open(appURL)
		}
	}

	onExit := func() {
		log.Println("[Netra Uptime Desktop] Shutting down application...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := serverApp.Shutdown(ctx); err != nil {
			log.Printf("[Netra Uptime Desktop] Shutdown warning: %v\n", err)
		}
		singleMutex.Release()
		log.Println("[Netra Uptime Desktop] Exit complete")
	}

	systray.Run(onReady, onExit)
}
