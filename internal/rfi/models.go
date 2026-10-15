package rfi

import "time"

// Episode represents a single "Journal en français facile" broadcast on RFI.
type Episode struct {
	// ID is the RFI contentId (UUID), falling back to MediaID or URL if absent.
	ID string
	// MediaID is the audio media identifier (e.g. "BE_6c65788c-...").
	MediaID string
	// DiffusionID is the broadcast identifier (e.g. "WBMZ86522-FLE-FR-20261014").
	DiffusionID string
	// Title is the cleaned episode headline.
	Title string
	// URL is the canonical episode page URL on francaisfacile.rfi.fr.
	URL string
	// MP3URL is the direct audio stream/download URL from sources[0].url.
	MP3URL string
	// DurationSecs is the audio duration in whole seconds (e.g. 600).
	DurationSecs int
	// PublishedAt is the parsed publication date/time in UTC.
	PublishedAt time.Time
	// PublishedDate is the normalized "YYYY-MM-DD" publication date for display and indexing.
	PublishedDate string

	// Transcript and Chapters are populated when scraping an individual episode page (1.3).
	Transcript string
	Chapters   []Chapter
}

// Chapter represents a timestamped topical segment within an Episode.
type Chapter struct {
	// Index is the 0-based order of the chapter within the episode.
	Index int
	// Title is the chapter label from div.m-chapters.
	Title string
	// StartSec is the chapter start timestamp in seconds.
	StartSec int
	// EndSec is the chapter end timestamp in seconds (next chapter start or episode duration).
	EndSec int
}
