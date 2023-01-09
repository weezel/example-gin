package logger

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	Logger zerolog.Logger
	uniqID string
)

func randID() string {
	data := make([]byte, 8)
	if _, err := rand.Read(data); err != nil {
		data = []byte{'f', 'o', 'o', 'b', 'a', 'r', 'b', 'z'}
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))[0:8]
}

func UniqID() string {
	return uniqID
}

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

	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	Logger = zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Str("uniq_id", uniqID).
		Caller().
		Logger()

	Logger.Debug().Msgf("Starting logger on level %s", Logger.GetLevel())
}
