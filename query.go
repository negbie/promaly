package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
	"github.com/ymotongpoo/datemaki"
)

// queryRange fetches data from Prometheus.
func queryRange(url, user, password, query string, start, end string, step time.Duration) (model.Matrix, error) {
	config := api.Config{
		Address: url,
	}

	if user != "" || password != "" {
		config.RoundTripper = promhttp.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Add("Authorization", "Basic "+
				base64.StdEncoding.EncodeToString([]byte(user+":"+password)))
			return http.DefaultTransport.RoundTrip(req)
		})
	}

	c, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %v", err)
	}

	api := v1.NewAPI(c)
	var stime, etime time.Time

	if end == "" {
		etime = time.Now()
	} else {
		etime, err = dateparse.ParseAny(end)
		if err != nil {
			etime, err = datemaki.Parse(end)
			if err != nil {
				return nil, fmt.Errorf("error parsing end time: %v", err)
			}
		}
	}

	if start == "" {
		stime = etime.Add(-10 * time.Minute)
	} else {
		stime, err = dateparse.ParseAny(start)
		if err != nil {
			stime, err = datemaki.Parse(start)
			if err != nil {
				return nil, fmt.Errorf("error parsing start time: %v", err)
			}
		}
	}

	if !stime.Before(etime) {
		fmt.Fprintln(os.Stderr, "start time is not before end time")
	}

	if step == 0 {
		resolution := math.Max(math.Floor(etime.Sub(stime).Seconds()/250), 1)
		// Convert seconds to nanoseconds such that time.Duration parses correctly.
		step = time.Duration(resolution) * time.Second
	}

	r := v1.Range{Start: stime, End: etime, Step: step}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	val, _, err := api.QueryRange(ctx, query, r) // Ignoring warnings for now.
	cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %v", err)
	}

	metrics, ok := val.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unsupported result format: %s", val.Type().String())
	}

	return metrics, nil
}
