package main

import (
	"flag"
	"fmt"
	"os"
)

type CmdlineArgs struct {
	UpstreamInterface string
}

// Print usage info
func Usage(exitCode int) {
	fmt.Printf("Usage: \n\n")
	fmt.Printf("  wirejumpd --upstream INTERFACE\n\n")
	fmt.Printf("WireJump background server\n")
	os.Exit(exitCode)
}

// Parse cmdline arguments
func ParseArgs() string {
	var upstream string

	flag.StringVar(&upstream, "upstream", "", "upstream WireGuard interface name")
	flag.Parse()

	if len(upstream) == 0 {
		fmt.Println("Error: --upstream option is missing")
		Usage(1)
	}

	return upstream
}
