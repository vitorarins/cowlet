package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/kouhin/envflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/vitorarins/cowlet/internal/metrics"
	"github.com/vitorarins/cowlet/pkg/client"
)

func backoff(msg string, attempt int) int {
	maxAttempts := 20
	backoff_time := math.Pow(2.0, float64(attempt))
	fmt.Printf("Backing off for %.2f milliseconds. Attempt %d of %d\n", backoff_time, attempt, maxAttempts)
	fmt.Printf("msg: %s\n", msg)

	if attempt >= maxAttempts {
		attempt = maxAttempts - 1
	}

	time.Sleep(time.Duration(backoff_time) * time.Millisecond)
	attempt++
	return attempt
}

func main() {
	email := flag.String("email", "", "User's email")
	password := flag.String("password", "", "User's password")

	metricsAddress := flag.String("metrics-address", ":9190", "Address used for the metrics service")

	if err := envflag.Parse(); err != nil {
		fmt.Printf("Could not parse flags: %v", err)
		os.Exit(1)
	}

	metrics := startMetrics(*metricsAddress)

	client, err := client.New(*email, *password)
	if err != nil {
		fmt.Printf("Failed to create client. %+v\n", err)
		return
	}

	err = client.SetFirstDevice()
	if err != nil {
		fmt.Printf("Failed to get devices. %+v\n", err)
		return
	}

	fmt.Println("Serving metrics...")

	attempts := 1
	for {
		realTimeVitals, err := client.GetRealTimeVitals(client.Device.DSN)
		if err != nil {
			attempts = backoff(fmt.Sprintf("Failed to get properties for %s. %+v\n", client.Device.DSN, err), attempts)
			continue
		}
		attempts = 1

		metrics.OxygenSaturationSet(realTimeVitals.OxygenSaturation)

		_, err = client.SetAppActiveStatus(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to set APP_ACTIVE: %+v\n", err)
		}

		time.Sleep(2 * time.Second)
	}
}

func startMetrics(metricsAddress string) metrics.Metrics {
	reg := prometheus.NewRegistry()

	metrics := metrics.New(reg)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	go func() {
		server := &http.Server{
			Addr: metricsAddress,
			// prevents DOS attacks according to
			// https://deepsource.io/directory/analyzers/go/issues/GO-S2114
			ReadHeaderTimeout: 3 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Prometheus stopped serving metrics: %v\n", err)

			return
		}
	}()

	return metrics
}
