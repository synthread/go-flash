package flash

import (
	"github.com/piotrjaromin/gpio"
	"github.com/pkg/errors"
	"go.bug.st/serial"
)

var DefaultBaud = 115200
var DefaultTTY = "/dev/ttyS1"

// Config defines configuration for communicating and flashing the
// microcontroller
type Config struct {
	Boot0GPIO int
	Boot1GPIO int
	PowerGPIO int

	BootloaderBaud int
	TTY            string
}

// Microcontroller represents an embedded microntroller chip that can be
// communicated with over UART
type Microcontroller struct {
	config *Config

	pinPower gpio.Pin
	pinBoot0 gpio.Pin
	pinBoot1 gpio.Pin

	stmCmdCodes          commandCodeMap
	stmBootloaderVersion byte

	ttyPort serial.Port
	ttyRx   chan byte

	ttyActive bool

	identity string
}

// NewMicrocontroller will create a new reference to a particular chip
func NewMicrocontroller(c *Config) (*Microcontroller, error) {
	if c == nil {
		c = &Config{}
	}

	if c.Boot0GPIO <= 0 {
		c.Boot0GPIO = 39
	}
	if c.Boot1GPIO <= 0 {
		c.Boot1GPIO = 41
	}
	if c.PowerGPIO <= 0 {
		c.PowerGPIO = 19
	}

	mc := &Microcontroller{
		config:      c,
		stmCmdCodes: commandCodeMap{},
	}

	if err := mc.setupPins(); err != nil {
		return nil, errors.Wrap(err, "could not setup pins")
	}

	return mc, nil
}

func (mc *Microcontroller) setupPins() (err error) {
	mc.pinPower, err = gpio.NewOutput(uint(mc.config.PowerGPIO), true)
	if err != nil {
		return
	}
	mc.pinBoot0, err = gpio.NewOutput(uint(mc.config.Boot0GPIO), false)
	if err != nil {
		return
	}
	mc.pinBoot1, err = gpio.NewOutput(uint(mc.config.Boot1GPIO), false)
	if err != nil {
		return
	}

	return
}

// Identify will report back a unique string with the ID of the chip
func (mc *Microcontroller) Identify() (string, error) {
	if mc.identity != "" {
		return mc.identity, nil
	}

	if !mc.IsOpen() {
		if err := mc.Open(); err != nil {
			return "", err
		}
		defer mc.Close()
	}

	// TODO: check if this is a STM
	pid, err := mc.stmCmdGetId()
	if err != nil {
		return "", err
	}
	mc.identity = "STM_" + pid

	return mc.identity, nil
}

// TTY will return the TTY that will be used
func (mc *Microcontroller) TTY() string {
	if mc.config.TTY != "" {
		return mc.config.TTY
	}
	return DefaultTTY
}

// BaudRate will return the baud rate used to connect to the TTY
func (mc *Microcontroller) BaudRate() int {
	if mc.config.BootloaderBaud > 0 {
		return mc.config.BootloaderBaud
	}
	return DefaultBaud
}

// Reset will force a power cycle on the microcontroller
func (mc *Microcontroller) Reset() {
	// TODO: check if this is a STM
	mc.exitSTBL()
}
