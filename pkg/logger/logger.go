package logger

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	Logger zerolog.Logger
	uniqID string
)

//nolint:gochecknoinits // init() is allowed here to avoid needless calls in each package. Import should be eonugh.
func init() {
	var logLevel zerolog.Level
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logLevel = zerolog.DebugLevel
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logLevel = zerolog.InfoLevel
	}

	uniqID = randID()

	startDir, err := os.Getwd()
	if err != nil {
		Logger.Fatal().Err(err).Msg("Failed to get working directory")
	}

	gitRoot, err := findGitRootDir(startDir)
	if err == nil {
		// Found a git dir
		zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
			// Make caller's filename path relative to the git root
			relPath, err := filepath.Rel(gitRoot, file)
			if err == nil {
				return fmt.Sprintf("%s:%d", relPath, line)
			}
			// Fallback to full path if relative conversion fails
			return fmt.Sprintf("%s:%d", file, line)
		}
	}

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.DurationFieldUnit = time.Second
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack //nolint:reassign // From their own example snippet

	Logger = zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Str("uniq_id", uniqID).
		Caller().
		Logger()

	if isInteractive() {
		Logger = Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	Logger.Debug().Msgf("Starting logger on level %s", Logger.GetLevel())
}

func randID() string {
	data := make([]byte, 8)
	if _, err := rand.Read(data); err != nil {
		// In practice, this shouldn't ever happen with the modern Linux kernel
		// now that entoropy pool is endless.
		data = []byte{'f', 'o', 'o', 'b', 'a', 'r', 'b', 'z'}
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[0:8]
}

func UniqID() string {
	return uniqID
}

func isInteractive() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// findGitRootDir finds the directory where the .git directory is located
func findGitRootDir(startDir string) (string, error) {
	currentDir := startDir
	for {
		// Check if the ".git" directory exists in the current directory
		gitPath := filepath.Join(currentDir, ".git")
		if stat, err := os.Stat(gitPath); err == nil && stat.IsDir() {
			return currentDir, nil
		}

		// Move one level up
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root of the filesystem
			return "", fmt.Errorf("could not find .git directory")
		}
		currentDir = parentDir
	}
}
