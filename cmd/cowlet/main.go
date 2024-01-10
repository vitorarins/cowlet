package main

import (
	"flag"
	"fmt"
	"math"
	"time"

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

	flag.Parse()

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

	attempts := 1
	for {
		realTimeVitals, err := client.GetRealTimeVitals(client.Device.DSN)
		if err != nil {
			attempts = backoff(fmt.Sprintf("Failed to get properties for %s. %+v\n", client.Device.DSN, err), attempts)
			continue
		}
		attempts = 1

		fmt.Printf("%+v\n", *realTimeVitals)

		_, err = client.SetAppActiveStatus(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to set APP_ACTIVE: %+v\n", err)
		}

		time.Sleep(2 * time.Second)
	}
}
