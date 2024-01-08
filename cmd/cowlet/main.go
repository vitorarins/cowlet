package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
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

	// Log out to file instead of new one each day.
	file, err := os.OpenFile("owlet_data.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to create file. %s\n", err)
		return
	}
	defer file.Close()

	attempts := 1
	propAttempts := 1
	for {
		properties, err := client.GetProperties(client.Device.DSN)
		if err != nil || properties == nil {
			propAttempts = backoff(fmt.Sprintf("Failed to get properties for %s. %+v\n", client.Device.DSN, err), propAttempts)
			continue
		}
		propAttempts = 1

		props_json, err := json.Marshal(properties)
		if err != nil {
			fmt.Printf("Failed to unmarshal properties for %s. %+v\n", client.Device.DSN, err)
			continue
		}

		// Append to our file.
		fmt.Fprintf(file, "%s\n", string(props_json))

		_, ok := properties["REAL_TIME_VITALS"]
		if !ok {
			attempts = backoff(fmt.Sprintf("Backing off due to REAL_TIME_VITALS"), attempts)
			continue

		}

		attempts = 1

		// STDOUT logging of specific stats.
		fmt.Printf("%+v\n", properties["REAL_TIME_VITALS"])
		fmt.Printf("%+v\n", properties["BASE_STATION_ON"])
		fmt.Printf("%+v\n", properties["BATT_LEVEL"])
		fmt.Printf("%+v\n", properties["OXYGEN_LEVEL"])
		fmt.Printf("%+v\n", properties["HEART_RATE"])

		_, err = client.SetAppActiveStatus(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to set APP_ACTIVE: %+v\n", err)
		}

		time.Sleep(2 * time.Second)
	}
}
