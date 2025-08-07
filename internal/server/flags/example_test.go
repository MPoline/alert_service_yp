package flags_test

import (
	"os"
	"testing"

	"github.com/MPoline/alert_service_yp/internal/server/flags"
)

func ExampleParseFlags() {
	os.Setenv("ADDRESS", "localhost:8081")
	os.Setenv("STORE_INTERVAL", "60")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
	}()

	flags.ParseFlags()

	println("Server address:", flags.FlagRunAddr)
	println("Store interval:", flags.FlagStoreInterval)
}

func TestParseFlags(t *testing.T) {
	t.Run("With env vars", func(t *testing.T) {
		os.Setenv("ADDRESS", "localhost:9090")
		defer os.Unsetenv("ADDRESS")

		flags.ParseFlags()
		if flags.FlagRunAddr != "localhost:9090" {
			t.Errorf("Expected localhost:9090, got %s", flags.FlagRunAddr)
		}
	})

	t.Run("With command line flags", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"cmd", "-a=:9091"}
		flags.ParseFlags()
		if flags.FlagRunAddr != ":9091" {
			t.Errorf("Expected :9091, got %s", flags.FlagRunAddr)
		}
	})
}
