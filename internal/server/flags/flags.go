package flags

import (
	"flag"
	"fmt"
	"os"
)

var FlagRunAddr string

func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if flag.NArg() > 0 {
		fmt.Println("Error: unknown flag(s)")
		flag.Usage()
		return
	}

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		fmt.Println("ADDRESS: ", envRunAddr)
		FlagRunAddr = envRunAddr
	}
}
