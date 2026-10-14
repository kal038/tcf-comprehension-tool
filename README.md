# tcf-comprehension-tool

A daily prep tool for the **TCF Canada** comprehension modules
(*Compréhension Écrite* & *Compréhension Orale*), targeting **NCLC 7+**.

Each day it will scrape the latest RFI *Journal en français facile* episode,
generate TCF-style multiple-choice questions with Gemini, and run them in a
terminal UI — showing the transcript on reading days, playing chapter-segmented
audio on listening days, then scoring the session and explaining every mistake.

## Status

Early scaffold (project skeleton + configuration + CLI entrypoint). Milestones
and issues are tracked in the Linear project **TCF Canada Comprehension Prep
TUI**.

## Requirements

- Go 1.27+
- [`mpv`](https://mpv.io/) for listening mode (added in a later milestone)

## Build & run

```sh
go build ./...
go run ./cmd/tcf-prep --help
go test ./...
```

## Configuration

| Env var | Purpose | Default |
| --- | --- | --- |
| `GEMINI_API_KEY` | Gemini API key (required by `quiz`) | — |
| `GEMINI_MODEL` | Model used for question generation | `gemini-2.5-flash` |
| `TCF_PREP_DATA_DIR` | Data directory override | `~/.local/share/tcf-prep` |

Local state (scraped episodes, generated quizzes, downloaded audio, and attempt
history) is cached under the data directory. `XDG_DATA_HOME` is respected when
`TCF_PREP_DATA_DIR` is unset.

## Commands

| Command | Description |
| --- | --- |
| `scrape` | Fetch and cache an RFI *Journal en français facile* episode |
| `quiz` | Generate and print a TCF-style quiz for a cached episode |
| `run` | Launch the interactive practice TUI |
| `version` | Print the version |
| `help` | Print usage |

## Layout

```
cmd/tcf-prep        CLI entrypoint
internal/config     Paths & configuration
internal/rfi        RFI scraping (episodes, transcripts, chapters)
internal/gemini     Gemini question generation
internal/audio      mpv audio playback
internal/store      SQLite cache & history
internal/tui        BubbleTea terminal UI
```
