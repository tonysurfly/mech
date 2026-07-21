package cmd

import (
	"io"
	"log/slog"
	"os"
)

// LevelTrace is below slog.LevelDebug. It carries wire-level detail (request
// payloads, response bodies) that is only useful for offline diagnosis via
// --trace, never for interactive --debug output.
const LevelTrace = slog.Level(-8)

// L is the package-wide diagnostic logger. It is configured by initLogger
// from the --debug and --trace root flags. It is never used for the
// program's primary report/discovery output, which goes through logger
// (see cmd_root.go) regardless of these flags.
var L *slog.Logger

// initLogger configures L according to level ("", "debug" or "trace") and
// returns a cleanup func that must run before the process exits.
//
//   - no flags: warnings and errors on stderr.
//   - level "debug": verbose diagnostics on stderr.
//   - level "trace": maximum detail written to tracePath, truncated on every
//     start, for agent self-diagnosis.
func initLogger(tracePath, level string) func() {
	w := io.Writer(os.Stderr)
	cleanup := func() {}
	lvl := slog.LevelWarn
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "trace":
		lvl = LevelTrace
		if tracePath != "" {
			f, err := os.OpenFile(tracePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
			if err == nil {
				w = f
				cleanup = func() { _ = f.Close() }
			}
		}
	}
	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: lvl})
	L = slog.New(h)
	slog.SetDefault(L)
	return cleanup
}

// L defaults to warnings-and-errors-on-stderr so it is always safe to use,
// even outside of Execute() (e.g. in tests or library use). PersistentPreRunE
// reconfigures it based on --debug/--trace.
func init() {
	initLogger("", "")
}
