package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
	"wirejump/internal/version"
)

func tabwriterPrint(lines []string) {
	writer := new(tabwriter.Writer)

	writer.Init(os.Stdout, 30, 4, 1, ' ', 0)

	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}

	writer.Flush()
}

func defaultUsageHandler(f string) error {
	ProgramUsage()

	os.Exit(0)

	return nil
}

func defaultVersionHandler(f string) error {
	fmt.Println(version.VersionString())

	os.Exit(0)

	return nil
}

func prettyTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(time.RFC1123)
}
