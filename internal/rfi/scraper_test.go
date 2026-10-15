package rfi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseIndex_Fixture(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "index_page1.html"))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	episodes, err := ParseIndex(f)
	if err != nil {
		t.Fatalf("ParseIndex returned unexpected error: %v", err)
	}

	if len(episodes) != 3 {
		t.Fatalf("got %d episodes, want 3", len(episodes))
	}

	want := []Episode{
		{
			ID:            "17283aea-c7ea-11f1-ad3b-6356017de54c",
			MediaID:       "BE_6c65788c-5b9b-4a1b-8d98-0930f4ff2b31",
			DiffusionID:   "WBMZ86522-FLE-FR-20261014",
			Title:         "Turquie: Halk TV censurée / Les Houthis attaquent l'Arabie Saoudite / Niger: marée noire / Nobel de la Paix: Navi Pillay sur RFI...",
			URL:           "https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20261014-turquie-halk-tv-censur%C3%A9e-les-houthis-attaquent-l-arabie-saoudite-niger-mar%C3%A9e-noire-nobel-de-la-paix-navi-pillay-sur-rfi",
			MP3URL:        "https://aod-fle.akamaized.net/rfi/francais/audio/jff/202610/journal_francais_facile_16h00_-_16h10_gmt_20261014.mp3",
			DurationSecs:  600,
			PublishedDate: "2026-10-14",
		},
		{
			ID:            "c4887c0c-c657-11f1-bde4-d71353921aeb",
			MediaID:       "BE_1b33f6d5-4c4e-4e63-a0d5-aac49b262ce6",
			DiffusionID:   "WBMZ86446-FLE-FR-20261012",
			Title:         "L'Otan bombe le torse / Inde: la colère gronde / France: stupeur au collège d'Aubagne / Chine: des forêts souterraines...",
			URL:           "https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20261012-l-otan-bombe-le-torse-inde-la-col%C3%A8re-gronde-france-stupeur-au-coll%C3%A8ge-d-aubagne-chine-des-for%C3%AAts-souterraines",
			MP3URL:        "https://aod-fle.akamaized.net/rfi/francais/audio/jff/202610/journal_francais_facile_16h00_-_16h10_gmt_20261012.mp3",
			DurationSecs:  600,
			PublishedDate: "2026-10-12",
		},
		{
			ID:            "465a5e5c-c3fc-11f1-a949-99146b54bbe7",
			MediaID:       "BE_692ab43a-9281-498b-bd9c-32fa8574cba6",
			DiffusionID:   "WBMZ86404-FLE-FR-20261009",
			Title:         "Prix Nobel de la Paix 2026: Navi Pillay / Yémen: après 12 ans de guerre / France: Victor Wembanyama apporte son soutien aux lycéens...",
			URL:           "https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20261009-prix-nobel-de-la-paix-2026-navi-pillay-y%C3%A9men-apr%C3%A8s-12-ans-de-guerre-france-victor-wembanyama-apporte-son-soutien-aux-lyc%C3%A9ens",
			MP3URL:        "https://aod-fle.akamaized.net/rfi/francais/audio/jff/202610/journal_francais_facile_16h00_-_16h10_gmt_20261009.mp3",
			DurationSecs:  600,
			PublishedDate: "2026-10-09",
		},
	}

	for i, w := range want {
		got := episodes[i]
		if got.ID != w.ID {
			t.Errorf("episode[%d].ID = %q, want %q", i, got.ID, w.ID)
		}
		if got.MediaID != w.MediaID {
			t.Errorf("episode[%d].MediaID = %q, want %q", i, got.MediaID, w.MediaID)
		}
		if got.DiffusionID != w.DiffusionID {
			t.Errorf("episode[%d].DiffusionID = %q, want %q", i, got.DiffusionID, w.DiffusionID)
		}
		if got.Title != w.Title {
			t.Errorf("episode[%d].Title = %q, want %q", i, got.Title, w.Title)
		}
		if got.URL != w.URL {
			t.Errorf("episode[%d].URL = %q, want %q", i, got.URL, w.URL)
		}
		if got.MP3URL != w.MP3URL {
			t.Errorf("episode[%d].MP3URL = %q, want %q", i, got.MP3URL, w.MP3URL)
		}
		if got.DurationSecs != w.DurationSecs {
			t.Errorf("episode[%d].DurationSecs = %d, want %d", i, got.DurationSecs, w.DurationSecs)
		}
		if got.PublishedDate != w.PublishedDate {
			t.Errorf("episode[%d].PublishedDate = %q, want %q", i, got.PublishedDate, w.PublishedDate)
		}
		if got.PublishedAt.IsZero() {
			t.Errorf("episode[%d].PublishedAt is zero", i)
		}
	}
}

func TestParseIndex_DeduplicatesPaginatedHeaderAndList(t *testing.T) {
	// On paginated pages (/2/, /3/), RFI repeats the first list item in the
	// .o-podcast-header__infos__cta block. Ensure it is returned only once.
	const html = `<!DOCTYPE html><html><body>
		<div class="o-podcast-header__infos__cta">
			<script data-media-id="BE_dup" type="application/json">
			{"contentId":"cid-1","mediaId":"BE_dup","sources":[{"url":"https://example.com/1.mp3","duration":600.02}],"diffusion":{"id":"D1","title":"Épisode 1","pubdate":"22/09/2026 - 20:49"},"url":"https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20260922-ep1","title":"Journal en français facile - Journal en français facile"}
			</script>
		</div>
		<ul class="o-podcast-list__list">
			<li class="m-podcast-item">
				<script data-media-id="BE_dup" type="application/json">
				{"contentId":"cid-1","mediaId":"BE_dup","sources":[{"url":"https://example.com/1.mp3","duration":600.02}],"diffusion":{"id":"D1","title":"Épisode 1","pubdate":"22/09/2026"},"url":"https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20260922-ep1","title":"Épisode 1 - Journal en français facile"}
				</script>
			</li>
			<li class="m-podcast-item">
				<script data-media-id="BE_2" type="application/json">
				{"contentId":"cid-2","mediaId":"BE_2","sources":[{"url":"https://example.com/2.mp3","duration":595.4}],"diffusion":{"id":"D2","title":"Épisode 2","pubdate":"21/09/2026"},"url":"https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20260921-ep2","title":"Épisode 2 - Journal en français facile"}
				</script>
			</li>
		</ul>
	</body></html>`

	episodes, err := ParseIndex(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseIndex error: %v", err)
	}
	if len(episodes) != 2 {
		t.Fatalf("got %d episodes, want 2", len(episodes))
	}
	if episodes[0].ID != "cid-1" || episodes[1].ID != "cid-2" {
		t.Errorf("unexpected episode IDs: %q, %q", episodes[0].ID, episodes[1].ID)
	}
	if episodes[1].DurationSecs != 595 {
		t.Errorf("episodes[1].DurationSecs = %d, want 595", episodes[1].DurationSecs)
	}
}

func TestParseIndex_FallbacksAndMalformedScripts(t *testing.T) {
	// Verifies:
	// - malformed JSON script is skipped without failing the whole page
	// - missing JSON url/title/pubdate falls back to DOM metadata and URL slug date
	const html = `<!DOCTYPE html><html><body>
		<script data-media-id="BE_bad" type="application/json">{not valid json</script>
		<ul class="o-podcast-list__list">
			<li class="m-podcast-item">
				<div class="m-podcast-item__infos">
					<span class="m-podcast-item__infos__date"></span>
					<a class="m-podcast-item__infos__edition" href="/fr/podcasts/journal-en-fran%C3%A7ais-facile/20261005-titre-depuis-dom">
						<h2>Titre depuis DOM</h2>
					</a>
				</div>
				<script data-media-id="BE_fallback" type="application/json">
				{"contentId":"cid-fallback","mediaId":"BE_fallback","sources":[{"url":"https://example.com/fb.mp3","duration":0}],"diffusion":{"id":"DFB","title":"","duration":"10:00"},"url":"","title":"Journal en français facile - Journal en français facile"}
				</script>
			</li>
		</ul>
	</body></html>`

	episodes, err := ParseIndex(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseIndex error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("got %d episodes, want 1", len(episodes))
	}
	ep := episodes[0]
	if ep.Title != "Titre depuis DOM" {
		t.Errorf("Title = %q, want %q", ep.Title, "Titre depuis DOM")
	}
	wantURL := "https://francaisfacile.rfi.fr/fr/podcasts/journal-en-fran%C3%A7ais-facile/20261005-titre-depuis-dom"
	if ep.URL != wantURL {
		t.Errorf("URL = %q, want %q", ep.URL, wantURL)
	}
	if ep.PublishedDate != "2026-10-05" {
		t.Errorf("PublishedDate = %q, want %q", ep.PublishedDate, "2026-10-05")
	}
	if ep.DurationSecs != 600 {
		t.Errorf("DurationSecs = %d, want 600", ep.DurationSecs)
	}
}

func TestClient_PageURL(t *testing.T) {
	c := NewClient()
	if got := c.PageURL(1); got != DefaultBaseURL {
		t.Errorf("PageURL(1) = %q, want %q", got, DefaultBaseURL)
	}
	wantPage2 := DefaultBaseURL + "2/"
	if got := c.PageURL(2); got != wantPage2 {
		t.Errorf("PageURL(2) = %q, want %q", got, wantPage2)
	}
}

func TestClient_FetchIndexPage(t *testing.T) {
	fixtureBytes, err := os.ReadFile(filepath.Join("testdata", "index_page1.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotPath, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotUA = r.Header.Get("User-Agent")
		if r.URL.Path == "/podcasts/3/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(fixtureBytes)
	}))
	defer srv.Close()

	client := NewClient(
		WithHTTPClient(srv.Client()),
		WithBaseURL(srv.URL+"/podcasts/"),
		WithUserAgent("tcf-prep-test/1.0"),
	)

	episodes, err := client.FetchIndexPage(context.Background(), 2)
	if err != nil {
		t.Fatalf("FetchIndexPage(2) unexpected error: %v", err)
	}
	if gotPath != "/podcasts/2/" {
		t.Errorf("requested path = %q, want /podcasts/2/", gotPath)
	}
	if gotUA != "tcf-prep-test/1.0" {
		t.Errorf("User-Agent = %q, want tcf-prep-test/1.0", gotUA)
	}
	if len(episodes) != 3 {
		t.Errorf("got %d episodes, want 3", len(episodes))
	}

	// Non-200 status should return an error.
	if _, err := client.FetchIndexPage(context.Background(), 3); err == nil {
		t.Errorf("expected error for HTTP 404 on page 3, got nil")
	}
}
