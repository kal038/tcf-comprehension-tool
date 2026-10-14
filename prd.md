Here is the architectural plan for your TCF Canada Comprehension Generator & Daily Prep TUI built in Go (BubbleTea).

1. Discovery: How RFI Journal en français facile is Structured
   I inspected the live HTML on francaisfacile.rfi.fr/fr/podcasts/journal-en-français-facile/ and individual episode pages. The site embeds clean, structured data that makes scraping and chapter-segmented audio playback very reliable:

- Episode List & Audio Metadata: Each episode embeds a <script data-media-id="..." type="application/json"> block containing:
- contentId & diffusion.id
- title & pubdate
- Direct MP3 URL (sources[0].url, e.g., https://aod-fle.akamaized.net/.../journal_francais_facile_...mp3) and duration (600s).
- Canonical episode page URL.
- Chapter Timestamps (div.m-chapters): Every episode page lists chapters with exact start times in seconds (v-bind:init-time="73", "174", "253", etc.) and chapter titles!
- Why this is huge for your Strict TCF Listening Mode: Instead of playing a 10-minute audio file all at once, we can pair each question with its corresponding news chapter segment (e.g., 01:13 to 02:54) and play only that segment once, just like an actual TCF Compréhension Orale audio extract.
- Transcript (div.m-transcription__content): Contains the full verbatim transcript in <p> tags.

2. Tech Stack & Libraries (Go)

- TUI Framework: Charmbracelet stack:
- github.com/charmbracelet/bubbletea (state machine & event loop)
- github.com/charmbracelet/lipgloss (layout, split panes, styling, NCLC badges)
- github.com/charmbracelet/bubbles (viewport for scrolling transcripts/explanations, list for episode selection, progress for audio timer, spinner for loading/LLM generation)
- HTML Scraping: github.com/PuerkitoBio/goquery + standard net/http (with appropriate User-Agent headers).
- LLM Generation: google.golang.org/genai (Official Go Gemini SDK, using gemini-2.5-flash with Structured JSON Outputs via ResponseSchema so parsing never fails).
- Audio Playback: mpv (already installed on your system at /usr/bin/mpv).
- Controlled via background subprocess (mpv --no-video --start=<sec> --end=<sec> <mp3_url_or_cached_path>) or IPC socket so the TUI can show a live progress bar while enforcing strict no-rewind TCF exam rules.
- Local Storage & Caching: modernc.org/sqlite (pure Go SQLite, no CGO needed) or local JSON store in ~/.local/share/tcf-prep/ to cache transcripts, MP3s, generated questions, and your historical scores.

3. TCF Canada Question Generation Strategy (NCLC 7+ Focus)
   To help you hit NCLC 7+ (B2–C1), the Gemini prompt and JSON schema will generate 8 questions per session (separate question sets for Reading vs. Listening so you can even do both on the same episode without spoiled questions, or pick one per day):

- Difficulty Progression:
- 2 × B1 (NCLC 5–6): Specific factual details, main idea of a short segment.
- 4 × B2 (NCLC 7–8): Implicit opinions, cause-and-effect, speaker/author intent, nuanced reformulations (paraphrasing without using the exact words from the transcript).
- 2 × C1 (NCLC 9–10): Subtle attitude/irony, synthesis across arguments, idiomatic/register comprehension (e.g., "bomber le torse", "ne pas faire dans la demi-mesure").
- TCF Distractor Rules:
- 4 choices (A, B, C, D) per question.
- Distractors must include classic TCF traps: lexical traps (uses words from the audio/text but flips the meaning), plausible extrapolations (not mentioned in the text), and partial truths.
- Chapter Mapping:
- Each generated question is linked to a chapter_index (start & end timestamp in seconds) and a transcript_excerpt (for the post-quiz explanation).
- Bilingual/Detailed Explanations:
- Each question stores:
- Why the correct option is right (with exact quote from the transcript).
- Why each of the other 3 distractors is wrong (identifying the specific trap).

4. TUI User Flow & Screens
1. Dashboard / Episode Picker Screen

- Displays recent episodes from RFI (fetched live + cached locally) with completion badges (e.g., 📖 Lu: 7/8 (NCLC 8) | 🎧 Écouté: Non fait) and your overall rolling NCLC estimate.
- Select an episode $\rightarrow$ Choose mode:
- [1] Compréhension Écrite (Reading)
- [2] Compréhension Orale (Listening - Strict TCF)

2. Loading / Generation Screen

- Scrapes transcript & chapters (if not cached).
- Calls Gemini with a spinner to generate the 8 TCF questions (cached immediately so re-opening is instant).

3. Quiz Screen — Mode A: Reading (Compréhension Écrite)

- Left Pane (60% width): Scrollable viewport with the formatted French transcript (j/k or ↑/↓ to scroll).
- Right Pane (40% width): Question X/8, CEFR/NCLC level badge, question stem, and options A / B / C / D (1-4 or a-d to select, Enter to confirm, Tab to switch focus between transcript and questions).

4. Quiz Screen — Mode B: Listening (Compréhension Orale - Strict TCF)

- Top Pane: Chapter title, audio status (🔊 Playing extract 01:13 – 02:54), and a non-seekable progress bar (strict TCF exam feel).
- Includes an optional emergency [r] Replay extract once keybind (with a warning badge that TCF only plays once) in case system audio glitches.
- Bottom Pane: Question stem and options A / B / C / D. Transcript is hidden during the quiz.

5. Results & Review Screen

- Summary Header:
- Score (e.g., 6 / 8 — 75%), breakdown by level (B1: 2/2, B2: 3/4, C1: 1/2), and estimated NCLC Band for this session.
- Interactive Question Review:
- Step through questions 1..8.
- Shows your answer vs. the correct answer (✓ green / ✗ red).
- Shows the relevant transcript excerpt (so in Listening mode, you can finally read what was said!).
- Shows Gemini's detailed explanation of why the correct answer is right and why your chosen distractor was a trap.
- Press [p] on any question in Review mode to replay that chapter's audio segment while reading the transcript!

5. Proposed Project Structure
   tcf-comprehension-generator/
   ├── cmd/
   │ └── tcf-prep/
   │ └── main.go # Entrypoint, CLI flags, env/config check (GEMINI_API_KEY)
   ├── internal/
   │ ├── config/
   │ │ └── config.go # Paths (~/.local/share/tcf-prep), API key loading
   │ ├── rfi/
   │ │ ├── scraper.go # Fetches episode list, parses JSON metadata, chapters & transcript
   │ │ └── models.go # Episode, Chapter, Transcript structs
   │ ├── gemini/
   │ │ ├── generator.go # Gemini API client, structured schema, TCF prompt engineering
   │ │ └── prompts.go # System prompts for CE (Reading) and CO (Listening)
   │ ├── audio/
   │ │ └── player.go # mpv wrapper for playing start..end timestamps & tracking progress
   │ ├── store/
   │ │ └── sqlite.go # Caches episodes, generated quizzes, and user attempt history
   │ └── tui/
   │ ├── app.go # Root BubbleTea model & screen router
   │ ├── home.go # Episode selector + Mode picker + NCLC stats summary
   │ ├── quiz_reading.go # Split-pane Reading UI
   │ ├── quiz_listening.go # Strict Audio segment player + Listening UI
   │ ├── review.go # Score breakdown, NCLC estimate, explanations + audio replay
   │ └── styles.go # Lipgloss color palette and layout helpers
   ├── go.mod
   └── README.md
