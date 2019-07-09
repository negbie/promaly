package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const version = "v0.1"

func main() {
	var (
		title    = flag.String("title", "Prometheus metrics", "Title of graph.")
		format   = flag.String("format", "png", "Image format.")
		file     = flag.String("file", "./graph", "File to save image to. Use '-' to write to stdout.")
		server   = flag.String("server", "http://localhost:9090", "Prometheus server to query.")
		user     = flag.String("user", "", "Basic Auth User.")
		password = flag.String("password", "", "Basic Auth Password.")
		query    = flag.String("query", "up", "PromQL query expression.")
		start    = flag.String("start", "1 hour ago", "Query range start time (RFC3339 or Unix timestamp or human).")
		end      = flag.String("end", "", "Query range end time (RFC3339 or Unix timestamp or human).")
		step     = flag.Duration("step", 1*time.Minute, "Query step size (duration).")
		subseq   = flag.Int("m", 0, "Subsequence length.")
		ver      = flag.Bool("version", false, "Print binary version.")
	)
	flag.Parse()

	if *ver {
		fmt.Printf("promaly %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	metrics, err := queryRange(*server, *user, *password, *query, *start, *end, *step)
	fail(err, "failed to get metrics")

	plot, err := Plot(metrics, *title, *format, *subseq)
	fail(err, "failed to create plot")

	if *file != "" {
		var out *os.File
		if *file == "-" {
			out = os.Stdout
		} else {
			f := *file
			if !strings.HasSuffix(f, "."+*format) {
				f = f + "." + *format
			}
			out, err = os.Create(f)
			fail(err, "failed to create file")
		}
		_, err = plot.WriteTo(out)
		fail(err, "failed to copy to file")
	}
}

func fail(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "msg: %v\n", err)
		os.Exit(1)
	}
}
