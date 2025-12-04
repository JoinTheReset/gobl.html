package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	gotenbergHTMLPath = "/forms/chromium/convert/html"
)

type gotenbergConvertor struct {
	client *resty.Client
}

func newGotenbergConvertor(opts ...Config) (*gotenbergConvertor, error) {
	conf := new(config)
	for _, opt := range opts {
		opt(conf)
	}
	if conf.url == "" {
		return nil, errors.New("gotenberg requires url parameter")
	}

	gc := new(gotenbergConvertor)
	gc.client = resty.New().SetBaseURL(conf.url)

	fmt.Printf("prepared gotenberg convertor with url: %s\n", conf.url)

	return gc, nil
}

func (gc *gotenbergConvertor) HTML(_ context.Context, data []byte, opts ...Option) ([]byte, error) {
	o := prepareOptions(opts)

	buf := bytes.NewBuffer(data)
	req := gc.client.R().
		SetFileReader("files", "index.html", buf)

	// Add metadata if provided
	// Gotenberg uses form fields for PDF metadata
	if o.metadata != nil {
		if o.metadata.Title != "" {
			req.SetFormData(map[string]string{"pdfMetadata": "true"})
			req.SetFormData(map[string]string{"metadata": buildGotenbergMetadata(o.metadata)})
		}
	}

	resp, err := req.Post(gotenbergHTMLPath)
	if err != nil {
		fmt.Printf("error sending to gotenberg: %s\n", err.Error())
		return nil, err
	}

	return resp.Body(), nil
}

// buildGotenbergMetadata creates a JSON string for Gotenberg's metadata field.
// Gotenberg expects metadata as a JSON object with specific keys.
func buildGotenbergMetadata(md *Metadata) string {
	parts := make([]string, 0)

	if md.Title != "" {
		parts = append(parts, fmt.Sprintf(`"Title":"%s"`, escapeJSON(md.Title)))
	}
	if md.Author != "" {
		parts = append(parts, fmt.Sprintf(`"Author":"%s"`, escapeJSON(md.Author)))
	}
	if md.Subject != "" {
		parts = append(parts, fmt.Sprintf(`"Subject":"%s"`, escapeJSON(md.Subject)))
	}
	if md.Keywords != "" {
		parts = append(parts, fmt.Sprintf(`"Keywords":"%s"`, escapeJSON(md.Keywords)))
	}
	if md.Creator != "" {
		parts = append(parts, fmt.Sprintf(`"Creator":"%s"`, escapeJSON(md.Creator)))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// escapeJSON escapes special characters for JSON strings.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
