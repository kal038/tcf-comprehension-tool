package rfi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	// DefaultBaseURL is the canonical index page for "Journal en français facile".
	DefaultBaseURL = "https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/"

	// DefaultOrigin is used to resolve relative episode hrefs on francaisfacile.rfi.fr.
	DefaultOrigin = "https://francaisfacile.rfi.fr"

	// DefaultUserAgent is sent on all outgoing HTTP requests so RFI serves the
	// standard desktop HTML markup.
	DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"

	defaultTimeout = 30 * time.Second
	showSuffix     = " - Journal en français facile"
	genericTitle   = "Journal en français facile"
)

var slugDatePrefixRe = regexp.MustCompile(`/(\d{8})-`)

// Client fetches and parses "Journal en français facile" pages from RFI.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// Option customizes a Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithBaseURL overrides the podcast index base URL (useful for tests).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = baseURL
		}
	}
}

// WithUserAgent overrides the User-Agent header sent on requests.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// NewClient constructs a scraper Client with sensible defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    DefaultBaseURL,
		userAgent:  DefaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// PageURL returns the index URL for the given 1-based page number.
// Page 1 returns the base index URL; page >= 2 appends "<page>/".
func (c *Client) PageURL(page int) string {
	base := strings.TrimRight(c.baseURL, "/") + "/"
	if page <= 1 {
		return base
	}
	return fmt.Sprintf("%s%d/", base, page)
}

// FetchIndexPage fetches the given 1-based podcast index page and parses its
// episode list and embedded audio metadata.
func (c *Client) FetchIndexPage(ctx context.Context, page int) ([]Episode, error) {
	targetURL := c.PageURL(page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build index request for page %d: %w", page, err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en;q=0.8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch index page %d (%s): %w", page, targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch index page %d (%s): unexpected HTTP status %d %s", page, targetURL, resp.StatusCode, resp.Status)
	}

	return ParseIndex(resp.Body)
}

// mediaJSON models the embedded <script data-media-id type="application/json">
// payload attached to each episode player button on RFI.
type mediaJSON struct {
	ContentID string        `json:"contentId"`
	MediaID   string        `json:"mediaId"`
	URL       string        `json:"url"`
	Title     string        `json:"title"`
	Sources   []mediaSource `json:"sources"`
	Diffusion struct {
		ID      string          `json:"id"`
		Title   string          `json:"title"`
		Pubdate string          `json:"pubdate"`
		RawDur  json.RawMessage `json:"duration"`
	} `json:"diffusion"`
}

type mediaSource struct {
	Name     string  `json:"name"`
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
}

type domItemMeta struct {
	url     string
	title   string
	pubdate string
}

// ParseIndex extracts ordered, deduplicated episodes from an RFI podcast index
// HTML document.
func ParseIndex(r io.Reader) ([]Episode, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse index HTML: %w", err)
	}

	// Collect DOM metadata keyed by data-media-id (from li.m-podcast-item and
	// the featured article card) so we can fall back if any JSON field is sparse.
	domByMediaID := collectDOMMetadata(doc)

	var (
		episodes []Episode
		seen     = make(map[string]struct{})
	)

	doc.Find("script[data-media-id]").Each(func(_ int, s *goquery.Selection) {
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}

		var payload mediaJSON
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return
		}

		mediaAttr, _ := s.Attr("data-media-id")
		domMeta := domByMediaID[firstNonEmpty(payload.MediaID, mediaAttr)]

		ep, ok := buildEpisode(payload, mediaAttr, domMeta)
		if !ok {
			return
		}

		dedupKey := ep.ID
		if dedupKey == "" {
			dedupKey = ep.URL
		}
		if _, exists := seen[dedupKey]; exists {
			return
		}
		seen[dedupKey] = struct{}{}
		episodes = append(episodes, ep)
	})

	return episodes, nil
}

// collectDOMMetadata indexes episode list items (and the featured top article)
// by their media-id so DOM title/URL/date attributes can supplement the JSON.
func collectDOMMetadata(doc *goquery.Document) map[string]domItemMeta {
	out := make(map[string]domItemMeta)

	// Featured top article on page 1: header CTA holds the script[data-media-id]
	// while the featured card holds .article__title a.
	var featuredURL, featuredTitle string
	if link := doc.Find(".article__title a").First(); link.Length() > 0 {
		if href, ok := link.Attr("href"); ok {
			featuredURL = resolveURL(href)
		}
		featuredTitle = normalizeSpace(link.Text())
	}
	if headerScript := doc.Find(".o-podcast-header__infos__cta script[data-media-id]").First(); headerScript.Length() > 0 {
		if mediaID, ok := headerScript.Attr("data-media-id"); ok && mediaID != "" {
			out[mediaID] = domItemMeta{
				url:   featuredURL,
				title: featuredTitle,
			}
		}
	}

	// Standard episode list items in ul.o-podcast-list__list > li.m-podcast-item.
	doc.Find("li.m-podcast-item").Each(func(_ int, item *goquery.Selection) {
		mediaID, _ := item.Find("script[data-media-id]").First().Attr("data-media-id")
		if mediaID == "" {
			mediaID, _ = item.Find("add-to-playlist[media-id]").First().Attr("media-id")
		}
		if mediaID == "" {
			return
		}

		edition := item.Find("a.m-podcast-item__infos__edition").First()
		href, _ := edition.Attr("href")
		if href == "" {
			href, _ = item.Find("a[data-article-item-link]").First().Attr("href")
		}
		title := normalizeSpace(edition.Text())
		if title == "" {
			title, _ = edition.Attr("title")
			title = normalizeSpace(title)
		}
		dateStr := normalizeSpace(item.Find(".m-podcast-item__infos__date").First().Text())

		out[mediaID] = domItemMeta{
			url:     resolveURL(href),
			title:   title,
			pubdate: dateStr,
		}
	})

	return out
}

func buildEpisode(p mediaJSON, attrMediaID string, dom domItemMeta) (Episode, bool) {
	mediaID := firstNonEmpty(p.MediaID, attrMediaID)
	canonURL := resolveURL(firstNonEmpty(p.URL, dom.url))

	var mp3URL string
	var durationSecs int
	if len(p.Sources) > 0 {
		mp3URL = strings.TrimSpace(p.Sources[0].URL)
		if p.Sources[0].Duration > 0 {
			durationSecs = int(math.Round(p.Sources[0].Duration))
		}
	}
	if durationSecs == 0 {
		durationSecs = parseDiffusionDuration(p.Diffusion.RawDur)
	}

	// Require at least an episode URL or MP3 URL so non-episode blocks are ignored.
	if canonURL == "" && mp3URL == "" {
		return Episode{}, false
	}

	title := cleanTitle(p.Diffusion.Title, dom.title, p.Title)
	pubTime, pubDate := parsePubDate(firstNonEmpty(p.Diffusion.Pubdate, dom.pubdate), canonURL)

	id := firstNonEmpty(p.ContentID, mediaID, canonURL)

	return Episode{
		ID:            id,
		MediaID:       mediaID,
		DiffusionID:   strings.TrimSpace(p.Diffusion.ID),
		Title:         title,
		URL:           canonURL,
		MP3URL:        mp3URL,
		DurationSecs:  durationSecs,
		PublishedAt:   pubTime,
		PublishedDate: pubDate,
	}, true
}

// cleanTitle picks the most specific episode headline and strips redundant
// show-title suffixes.
func cleanTitle(diffusionTitle, domTitle, topTitle string) string {
	for _, candidate := range []string{diffusionTitle, domTitle, topTitle} {
		t := normalizeSpace(candidate)
		if t == "" {
			continue
		}
		t = strings.TrimSuffix(t, showSuffix)
		t = normalizeSpace(t)
		if t != "" && !strings.EqualFold(t, genericTitle) {
			return t
		}
	}
	return normalizeSpace(firstNonEmpty(diffusionTitle, domTitle, topTitle))
}

// parsePubDate parses RFI's "DD/MM/YYYY - HH:MM" or "DD/MM/YYYY" date strings
// and falls back to the YYYYMMDD date prefix in the canonical URL slug.
func parsePubDate(raw string, canonURL string) (time.Time, string) {
	s := normalizeSpace(raw)
	for _, layout := range []string{"02/01/2006 - 15:04", "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			utc := t.UTC()
			return utc, utc.Format("2006-01-02")
		}
	}

	if m := slugDatePrefixRe.FindStringSubmatch(canonURL); len(m) == 2 {
		if t, err := time.Parse("20060102", m[1]); err == nil {
			utc := t.UTC()
			return utc, utc.Format("2006-01-02")
		}
	}

	return time.Time{}, ""
}

// parseDiffusionDuration parses diffusion.duration, which RFI encodes either
// as a numeric float (e.g. 600.02) or as a "MM:SS" / "HH:MM:SS" string (e.g. "10:00").
func parseDiffusionDuration(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var secFloat float64
	if err := json.Unmarshal(raw, &secFloat); err == nil && secFloat > 0 {
		return int(math.Round(secFloat))
	}
	var durStr string
	if err := json.Unmarshal(raw, &durStr); err == nil {
		parts := strings.Split(strings.TrimSpace(durStr), ":")
		switch len(parts) {
		case 2:
			m, errM := strconv.Atoi(parts[0])
			s, errS := strconv.Atoi(parts[1])
			if errM == nil && errS == nil && m >= 0 && s >= 0 {
				return m*60 + s
			}
		case 3:
			h, errH := strconv.Atoi(parts[0])
			m, errM := strconv.Atoi(parts[1])
			s, errS := strconv.Atoi(parts[2])
			if errH == nil && errM == nil && errS == nil && h >= 0 && m >= 0 && s >= 0 {
				return h*3600 + m*60 + s
			}
		}
	}
	return 0
}

func resolveURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	if u.IsAbs() {
		return u.String()
	}
	base, err := url.Parse(DefaultOrigin)
	if err != nil {
		return s
	}
	return base.ResolveReference(u).String()
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
