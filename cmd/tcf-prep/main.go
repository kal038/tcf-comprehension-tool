// Command tcf-prep is the entrypoint for the TCF Canada comprehension prep
// tool: it scrapes "Journal en français facile" episodes from RFI, generates
// TCF-style questions with Gemini, and runs the interactive practice TUI.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kal038/tcf-comprehension-tool/internal/config"
)

// version is overridable at build time:
//
//	go build -ldflags "-X main.version=1.2.3"
var version = "0.1.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches the CLI and returns the process exit code. It is separated
// from main so tests can exercise it without spawning a process.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printUsage(stdout)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "tcf-prep %s\n", version)
		return 0
	case "--test-scrape":
		// Convenience alias from the milestone plan; equivalent to
		// "tcf-prep scrape --test".
		return runScrape(append([]string{"--test"}, args[1:]...), stdout, stderr)
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "scrape":
		return runScrape(rest, stdout, stderr)
	case "quiz":
		return runQuiz(rest, stdout, stderr)
	case "run":
		return runTUI(rest, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "tcf-prep: unknown command %q\n\n", cmd)
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `tcf-prep — TCF Canada comprehension practice

Usage:
  tcf-prep <command> [flags]

Commands:
  scrape   Fetch and cache an RFI "Journal en français facile" episode
  quiz     Generate and print a TCF-style quiz for a cached episode
  run      Launch the interactive practice TUI
  version  Print the version and exit
  help     Print this help and exit

Global flags:
  -h, --help     Show this help
  -v, --version  Show the version

Environment:
  GEMINI_API_KEY      Required by "quiz" (Gemini API key)
  GEMINI_MODEL        Optional model override (default %q)
  TCF_PREP_DATA_DIR   Optional data directory override (default ~/.local/share/tcf-prep)

Run "tcf-prep <command> --help" for command-specific flags.
`, config.DefaultGeminiModel)
}

// loadConfig resolves configuration, reporting any error to stderr.
func loadConfig(stderr io.Writer) (*config.Config, int) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "tcf-prep: %v\n", err)
		return nil, 1
	}
	return cfg, 0
}

func runScrape(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("scrape", flag.ContinueOnError)
	fs.SetOutput(stderr)
	episode := fs.String("episode", "latest", "episode to fetch: \"latest\" or a full episode URL")
	test := fs.Bool("test", false, "print a transcript preview after scraping")
	if err := fs.Parse(args); err != nil {
		return parseExitCode(err)
	}

	if _, code := loadConfig(stderr); code != 0 {
		return code
	}
	_ = episode
	_ = test
	_ = stdout

	fmt.Fprintln(stderr, "tcf-prep: scrape is not implemented yet")
	return 1
}

func runQuiz(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("quiz", flag.ContinueOnError)
	fs.SetOutput(stderr)
	mode := fs.String("mode", "ce", "quiz mode: \"ce\" (reading) or \"co\" (listening)")
	episode := fs.String("episode", "latest", "episode to use: \"latest\" or a full episode URL")
	if err := fs.Parse(args); err != nil {
		return parseExitCode(err)
	}

	switch *mode {
	case "ce", "co":
	default:
		fmt.Fprintf(stderr, "tcf-prep: invalid --mode %q (want \"ce\" or \"co\")\n", *mode)
		return 2
	}

	cfg, code := loadConfig(stderr)
	if code != 0 {
		return code
	}
	if err := cfg.RequireGeminiAPIKey(); err != nil {
		fmt.Fprintf(stderr, "tcf-prep: %v\n", err)
		fmt.Fprintln(stderr, "Set it in your environment, for example:  export GEMINI_API_KEY=your-key")
		return 1
	}
	_ = episode
	_ = stdout

	fmt.Fprintln(stderr, "tcf-prep: quiz generation is not implemented yet")
	return 1
}

func runTUI(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return parseExitCode(err)
	}

	if _, code := loadConfig(stderr); code != 0 {
		return code
	}
	_ = stdout

	fmt.Fprintln(stderr, "tcf-prep: the interactive TUI is not implemented yet")
	return 1
}

// parseExitCode maps a flag parsing error to an exit code. An explicit help
// request is a success.
func parseExitCode(err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	return 2
}
