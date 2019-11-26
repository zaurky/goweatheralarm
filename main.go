// Example of retrieving stored messages.
package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/barnybug/gogsmmodem"
	"github.com/tarm/serial"

	"github.com/zaurky/go-yrapi/yrapi"
)

const Default_duration int = 5  // in days
const Default_threshold int = 3 // in °C
const Default_tty = "/dev/ttyUSB0"

type Config struct {
	duration  int
	threshold int
	lat       float64
	long      float64
	phone     string
	tty       string
}

func (c *Config) ParseConfig() error {
	flag.IntVar(&c.duration, "duration", Default_duration, "How many days in the future should we watch (default:5)")
	flag.IntVar(&c.threshold, "threshold", Default_threshold, "Under which value (in °C) should we raise (default:3)")
	flag.Float64Var(&c.lat, "latitude", 0, "The latitude where to watch")
	flag.Float64Var(&c.long, "longitude", 0, "The longitude where to watch")
	flag.StringVar(&c.phone, "phone", "", "The phone number to which we send message.")
	flag.StringVar(&c.tty, "tty", Default_tty, "The serial interface for the GSM adapter.")

	flag.Parse()
	if len(c.phone) == 0 {
		fmt.Println("Missing arguement phone")
		return errors.New("Missing arguement phone")
	}
	if c.lat == 0 {
		fmt.Println("Missing arguement lat")
		return errors.New("Missing arguement lat")
	}
	if c.long == 0 {
		fmt.Println("Missing arguement long")
		return errors.New("Missing arguement long")
	}
	return nil
}

func SendSMS(conf Config, message string) {
	serial_conf := serial.Config{Name: conf.tty, Baud: 115200}
	modem, err := gogsmmodem.Open(&serial_conf, true)
	if err != nil {
		panic(err)
	}

	err = modem.SendMessage(conf.phone, message)
	if err != nil {
		panic(err)
	}
}

func main() {
	var conf Config
	err := conf.ParseConfig()
	if err != nil {
		return
	}

	weatherData, err := yrapi.LocationforecastLTS(conf.lat, conf.long)
	if err != nil {
		panic(err)
	}

	full_message := ""
	for index, time := range weatherData.Product.WeatherTimes {
		if index > 24*conf.duration {
			break
		}
		if time.Location.Temperature != nil && time.Location.Temperature.Value <= float32(conf.threshold) {
			message := fmt.Sprintf("Temperature go down %2.1f on %s\n", time.Location.Temperature.Value, time.From)
			fmt.Print(message)
			full_message += message
		}
	}
	if len(full_message) != 0 {
		SendSMS(conf, full_message[0:159])
	}
}
