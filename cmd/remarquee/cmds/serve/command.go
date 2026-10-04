package serve

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/remarquee/pkg/rmcloud"
	"github.com/go-go-golems/remarquee/pkg/rmfiles"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Settings holds the `remarquee serve` flags.
type Settings struct {
	Addr             string
	Dev              bool
	DefaultRemoteDir string
	IncludeTemplates bool
	Workers          int
}

// NewServeCommand builds the `serve` subcommand.
func NewServeCommand() *cobra.Command {
	s := &Settings{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve a local web UI to upload, manage, and search reMarkable files",
		Long: strings.TrimSpace(`
Serve a local, loopback-only web UI for your reMarkable cloud files.

The UI lets you upload Markdown/PDF/EPUB files, browse and search the cloud
tree, create folders, rename/move entries, and delete entries. It runs in the
existing remarquee binary and never exposes your credentials to the browser.

Examples:
  remarquee serve
  remarquee serve --addr 127.0.0.1:9090
  remarquee serve --remote-dir /ai
  remarquee serve --include-templates --workers 4
`),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context(), s)
		},
	}

	cmd.Flags().StringVar(&s.Addr, "addr", "127.0.0.1:8080", "HTTP listen address (loopback by default)")
	cmd.Flags().BoolVar(&s.Dev, "dev", false, "Disable static asset caching while editing the frontend")
	cmd.Flags().StringVar(&s.DefaultRemoteDir, "remote-dir", "/", "Default upload destination directory")
	cmd.Flags().BoolVar(&s.IncludeTemplates, "include-templates", false, "Show template documents (hidden by default)")
	cmd.Flags().IntVar(&s.Workers, "workers", 2, "Concurrent conversions per upload job")

	return cmd
}

func runServe(ctx context.Context, s *Settings) error {
	log.Info().Str("remote_dir", s.DefaultRemoteDir).Msg("remarquee serve: initializing cloud context (first run may take a while)")

	svc, err := rmfiles.NewServiceFromCloud(ctx, rmfiles.Config{
		Auth:             rmcloud.AuthSettings{NonInteractive: true, Progress: os.Stderr},
		DefaultRemoteDir: s.DefaultRemoteDir,
		IncludeTemplates: s.IncludeTemplates,
		Workers:          s.Workers,
	})
	if err != nil {
		return err
	}

	srv := NewServer(svc, Options{Dev: s.Dev})
	httpSrv := &http.Server{
		Addr:              s.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       120 * time.Second,
		// Downloads and (synchronous) request bodies can be large, so no write
		// timeout is enforced; uploads run as background jobs.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	log.Info().Str("url", "http://"+s.Addr).Msg("remarquee serve listening")

	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
