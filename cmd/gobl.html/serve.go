package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/invopop/gobl"
	goblhtml "github.com/invopop/gobl.html"
	"github.com/invopop/gobl.html/assets"
	"github.com/invopop/gobl.html/layout"
	"github.com/invopop/gobl.html/pkg/pdf"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/ziflex/lecho/v3"
)

type serveOpts struct {
	*rootOpts
	port   string
	pdf    string
	pdfURL string

	convertor pdf.Convertor
}

func serve(o *rootOpts) *serveOpts {
	return &serveOpts{rootOpts: o}
}

func (s *serveOpts) cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve serves GOBL files from the examples directory as HTML",
		RunE:  s.runE,
	}

	f := cmd.Flags()
	f.StringVarP(&s.port, "port", "p", "3000", "port to listen on")
	f.StringVarP(&s.pdf, "pdf", "", "", "PDF convertor to use")
	f.StringVarP(&s.pdfURL, "pdf-url", "", "", "URL of the PDF convertor to use (if needed)")

	return cmd
}

func (s *serveOpts) runE(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	var err error
	opts := make([]pdf.Config, 0)
	if s.pdfURL != "" {
		opts = append(opts, pdf.WithURL(s.pdfURL))
	}
	if s.convertor, err = pdf.New(s.pdf, opts...); err != nil {
		return fmt.Errorf("preparing PDF convertor: %w", err)
	}

	e := prepareEcho()
	e.StaticFS("/styles", echo.MustSubFS(assets.Content, "styles"))
	// Changed from GET /:filename to POST /
	e.POST("/", s.handlePost)

	var startErr error
	go func() {
		err := e.Start(":" + s.port)
		if !errors.Is(err, http.ErrServerClosed) {
			startErr = err
		}
		cancel()
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return startErr
}

// handlePost handles POST requests with a GOBL JSON body
func (s *serveOpts) handlePost(c echo.Context) error {
	// Read the request body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("reading request body: %v", err))
	}
	defer c.Request().Body.Close()

	// Unmarshal the JSON into a GOBL envelope (includes validation)
	env := new(gobl.Envelope)
	if err := json.Unmarshal(bodyBytes, env); err != nil {
		// Consider providing more specific validation error feedback if possible
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("unmarshalling GOBL envelope: %v", err))
	}

	// Prepare options for rendering
	// Ensure stylesheets are embedded for PDF generation
	opts := []goblhtml.Option{
		goblhtml.WithEmbeddedStylesheets(),
		goblhtml.WithLayout(layout.A4), // Default to A4 layout
		// Add default locale or other options if needed, potentially from headers?
		// goblhtml.WithLocale(i18n.CodeEN),
	}

	// Render HTML
	// Note: We pass nil for 'req *options' as it's no longer used
	htmlData, err := s.render(c, env, opts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("rendering HTML: %v", err))
	}

	// Render PDF, attaching the original GOBL JSON
	return s.renderPDF(c, bodyBytes, htmlData)
}

func prepareEcho() *echo.Echo {
	e := echo.New()
	zl := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	})
	b := zl.With().Timestamp()
	log.Logger = b.Logger()
	lvl, _ := lecho.MatchZeroLevel(zerolog.DebugLevel)
	logger := lecho.From(
		log.Logger,
		lecho.WithLevel(lvl),
		lecho.WithTimestamp(),
		// lecho.WithCaller(), // useful for debugging
	)
	e.Logger = logger
	e.Use(lecho.Middleware(lecho.Config{Logger: logger}))
	e.Use(middleware.Recover())
	return e
}

// render generates the HTML content for a given envelope and options.
// Removed the 'req *options' parameter as it's no longer needed for the POST endpoint.
// Removed logic related to query parameters as they are not used in the POST flow.
// Corrected function signature: removed reference to undefined 'options' struct.
func (s *serveOpts) render(c echo.Context, env *gobl.Envelope, opts []goblhtml.Option) ([]byte, error) {
	ctx := c.Request().Context()

	// Apply default options or options derived from the envelope/context if needed
	// For now, we rely on options passed in from handlePost

	out, err := goblhtml.Render(ctx, env, opts...)
	if err != nil {
		return nil, fmt.Errorf("generating html: %w", err)
	}

	return out, nil
}

// Removed the 'options' struct as it was tied to the GET /:filename endpoint

// Removed the 'generate' function as it's replaced by 'handlePost'

// renderPDF converts HTML data to PDF, attaching the original GOBL JSON.
// Changed signature to accept goblJSON and htmlData.
func (s *serveOpts) renderPDF(c echo.Context, goblJSON, htmlData []byte) error {
	if s.convertor == nil {
		return errors.New("no PDF convertor available")
	}

	// prepare the GOBL attachment
	opts := []pdf.Option{
		pdf.WithAttachment(&pdf.Attachment{
			Data:     goblJSON, // Use the original JSON data passed in
			Filename: "gobl.json",
		}),
	}

	out, err := s.convertor.HTML(c.Request().Context(), htmlData, opts...)
	if err != nil {
		return fmt.Errorf("converting to PDF: %w", err)
	}

	return c.Blob(http.StatusOK, "application/pdf", out)
}
