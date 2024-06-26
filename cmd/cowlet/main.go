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

func entrypointHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User request received!")
}

func main() {
	email := flag.String("email", "", "User's email")
	password := flag.String("password", "", "User's password")

	httpAddress := flag.String("http-address", ":8000", "Address used for the service")
	metricsAddress := flag.String("metrics-address", ":8080", "Address used for the metrics service")

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

	entrypointMux := http.NewServeMux()
	entrypointMux.HandleFunc("/", entrypointHandler)
	entrypointMux.HandleFunc("/startup", entrypointHandler)
	entrypointMux.HandleFunc("/liveness", entrypointHandler)

	go func() {
		if err := http.ListenAndServe(*httpAddress, entrypointMux); err != nil {
			fmt.Printf("Prometheus stopped serving metrics: %v\n", err)

			return
		}
	}()

	maxAttempts := 5
	attempt := 1
	for {
		realTimeVitals, err := client.GetRealTimeVitals(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to get properties for %s. %+v\n", client.Device.DSN, err)

			if attempt > maxAttempts {
				os.Exit(1)
			}

			backoff_time := math.Pow(2.0, float64(attempt))

			fmt.Printf("Backing off for %.2f milliseconds. Attempt %d of %d\n", backoff_time, attempt, maxAttempts)

			time.Sleep(time.Duration(backoff_time) * time.Millisecond)
			attempt++

			continue
		}
		attempt = 1

		metrics.OxygenSaturationSet(realTimeVitals.OxygenSaturation)
		metrics.HeartRateSet(realTimeVitals.HeartRate)
		metrics.BatteryPercentageSet(realTimeVitals.BatteryPercentage)
		metrics.BatteryMinutesSet(realTimeVitals.BatteryMinutes)
		metrics.SignalStrengthSet(realTimeVitals.SignalStrength)
		metrics.OxygenTenAVSet(realTimeVitals.OxygenTenAV)
		metrics.SockConnectionSet(realTimeVitals.SockConnection)
		metrics.SleepStateSet(realTimeVitals.SleepState)
		metrics.SkinTemperatureSet(realTimeVitals.SkinTemperature)
		metrics.MovementSet(realTimeVitals.Movement)
		metrics.AlertPausedStatusSet(realTimeVitals.AlertPausedStatus)
		metrics.ChargingSet(realTimeVitals.Charging)
		metrics.MovementBucketSet(realTimeVitals.MovementBucket)
		metrics.WellnessAlertSet(realTimeVitals.WellnessAlert)
		metrics.MonitoringStartTimeSet(realTimeVitals.MonitoringStartTime)
		metrics.BaseBatteryStatusSet(realTimeVitals.BaseBatteryStatus)
		metrics.BaseStationOnSet(realTimeVitals.BaseStationOn)

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
