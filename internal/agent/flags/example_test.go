package flags_test

import (
	"os"
	"testing"

	"github.com/MPoline/alert_service_yp/internal/agent/flags"
)

func ExampleParseFlags() {

	os.Setenv("ADDRESS", "localhost:8081")
	os.Setenv("REPORT_INTERVAL", "5")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
	}()

	flags.ParseFlags()

	println("Server address:", flags.FlagRunAddr)
	println("Report interval:", flags.FlagReportInterval)
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		setEnv         func()
		expectedAddr   string
		expectedReport int64
	}{
		{
			name:           "Default values",
			setEnv:         func() {},
			expectedAddr:   ":8080",
			expectedReport: 10,
		},
		{
			name: "Environment override",
			setEnv: func() {
				os.Setenv("ADDRESS", "localhost:9090")
				os.Setenv("REPORT_INTERVAL", "5")
			},
			expectedAddr:   "localhost:9090",
			expectedReport: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEnv := map[string]string{
				"ADDRESS":         os.Getenv("ADDRESS"),
				"REPORT_INTERVAL": os.Getenv("REPORT_INTERVAL"),
				"POLL_INTERVAL":   os.Getenv("POLL_INTERVAL"),
				"KEY":             os.Getenv("KEY"),
				"RATE_LIMIT":      os.Getenv("RATE_LIMIT"),
			}
			defer func() {
				for k, v := range originalEnv {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}
			}()

			tt.setEnv()

			flags.ParseFlags()

			if flags.FlagRunAddr != tt.expectedAddr {
				t.Errorf("Expected address %s, got %s", tt.expectedAddr, flags.FlagRunAddr)
			}
			if flags.FlagReportInterval != tt.expectedReport {
				t.Errorf("Expected report interval %d, got %d", tt.expectedReport, flags.FlagReportInterval)
			}
		})
	}
}
