Piece 1: RFI Scraper & Local SQLite Cache
Goal: Fetch the latest episodes from RFI, extract the audio URL, chapter timestamps, and full transcript, and save them locally so we never scrape the same page twice.

- What we'll build:
- go.mod setup and project skeleton.
- internal/rfi: Scraper that fetches the podcast index page, extracts the embedded JSON metadata (title, date, MP3 URL, duration), and scrapes individual episode pages for the verbatim transcript (div.m-transcription__content) and chapter timestamps (div.m-chapters).
- internal/store: SQLite database (~/.local/share/tcf-prep/tcf.db) to cache episodes and chapters.
- How we'll test it: Run go run ./cmd/tcf-prep --test-scrape (or a unit test) to print today's episode title, MP3 link, chapter list with start/end times, and transcript preview in the terminal.
  Piece 2: Gemini TCF Question Generator (B1–C1 / NCLC 7+ Focus)
  Goal: Feed a cached episode (transcript + chapters) into Gemini and get back strictly-typed JSON containing 8 TCF-style multiple-choice questions mapped to chapters.
- What we'll build:
- internal/gemini: Integration with google.golang.org/genai (gemini-2.5-flash) using Structured Outputs (ResponseSchema).
- Specialized prompts for Compréhension Écrite (Reading) and Compréhension Orale (Listening) targeting 2× B1, 4× B2, 2× C1 with authentic TCF traps and detailed explanations (why the right answer is right + why each distractor is wrong + exact quote).
- Save generated quizzes in SQLite so they are cached per episode & mode.
- How we'll test it: Run a CLI command to generate a quiz for the latest cached episode and pretty-print the generated questions, levels, chapter time ranges, and explanations.
  Piece 3: Core TUI + Reading Mode (Compréhension Écrite)
  Goal: Build the interactive BubbleTea TUI for selecting an episode, taking a Reading quiz, and reviewing explanations.
- What we'll build:
- Home Screen: Episode list (showing date, title, completion status, and rolling NCLC score) + Mode selector ([1] Reading, [2] Listening).
- Reading Quiz Screen: Split-pane layout — scrollable French transcript on the left (j/k or ↑/↓), question & A/B/C/D choices on the right.
- Review & Score Screen: Shows total score, score by CEFR level, estimated NCLC level, and lets you step through each question to read the transcript quote and explanation for mistakes.
- Persist completed attempts in SQLite.
- How we'll test it: Launch the TUI, pick today's episode in Reading mode, answer the 8 questions, and check the explanations and saved score.
  Piece 4: Audio Engine & Strict TCF Listening Mode (Compréhension Orale)
  Goal: Add mpv-powered chapter-segmented audio playback and the Listening quiz screen.
- What we'll build:
- internal/audio: Controller for mpv that plays a specific [start_sec, end_sec] slice of the episode MP3 (or downloads/caches the MP3 locally first for instant seeking) and tracks playback progress.
- Listening Quiz Screen: Hides the transcript, automatically plays the current question's audio chapter once with a live progress bar, and lets you select A/B/C/D (plus an emergency [r] single-replay keybind).
- Review Screen Audio Hook: Press [p] while reviewing results to replay that question's audio segment while reading the transcript excerpt and explanation.
- How we'll test it: Launch the TUI, pick Listening mode, listen to each chapter extract, answer the questions, and test replaying segments on the Review screen.
  Piece 5: Polish, Stats Dashboard & UX Refinements
  Goal: Make daily usage effortless and informative.
- What we'll build:
- Progress tracker header on the Home Screen (streak, average Reading NCLC, average Listening NCLC).
- Keybind help bar at the bottom of each screen.
- Graceful error states (missing GEMINI_API_KEY, network offline fallback to cached episodes, clean mpv process cleanup on exit)
