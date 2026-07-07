// Command watcher polls Waves mainnet for burns and writes the artifacts.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hearthchain/burning-page/internal/chain/waves"
	"github.com/hearthchain/burning-page/internal/config"
	"github.com/hearthchain/burning-page/internal/watcher"
)

const pollInterval = 60 * time.Second // ~Waves block time; polling faster buys nothing

func main() { os.Exit(run()) }

func run() int {
	configPath := flag.String("config", "config.json", "path to the shared config")
	once := flag.Bool("once", false, "run a single poll and exit")
	fixture := flag.String("fixture", "", "fixture directory replacing both nodes (offline end-to-end mode)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("config", "err", err)
		return 1
	}
	var primary, secondary watcher.Node
	if *fixture != "" {
		node := watcher.NewFileNode(*fixture)
		primary, secondary = node, node
	} else {
		primary = waves.NewClient(cfg.Nodes.Primary)
		secondary = waves.NewClient(cfg.Nodes.Secondary)
	}
	w := &watcher.Watcher{Primary: primary, Secondary: secondary, Cfg: cfg}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		if pollErr := w.Poll(ctx); pollErr != nil {
			slog.Error("poll", "err", pollErr)
			if *once {
				return 1
			}
		}
		if *once {
			return 0
		}
		select {
		case <-ctx.Done():
			return 0
		case <-time.After(pollInterval):
		}
	}
}
