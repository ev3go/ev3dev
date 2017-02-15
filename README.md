![Gopherbrick](gopherbrick.png)
# ev3dev is an idiomatic Go interface to an ev3dev device

[![Build Status](https://travis-ci.org/ev3go/ev3dev.svg?branch=master)](https://travis-ci.org/ev3go/ev3dev) [![Coverage Status](https://coveralls.io/repos/ev3go/ev3dev/badge.svg?branch=master&service=github)](https://coveralls.io/github/ev3go/ev3dev?branch=master) [![GoDoc](https://godoc.org/github.com/ev3go/ev3dev?status.svg)](https://godoc.org/github.com/ev3go/ev3dev)

The goal is to implement a simple Go style ev3dev API and helpers for common tasks.

github.com/ev3go/ev3dev depends on an 18-ev3dev kernel or better (http://www.ev3dev.org/news/2017/01/25/kernel-release-cycle-18/).

For device-specific functions see [EV3](https://github.com/ev3go/ev3) and [BrickPi](https://github.com/ev3go/brickpi).

## Currently supported:

### Low level API

- [x] Automatic identification of attached devices
- [x] Buttons `/dev/input/by-path/platform-gpio-keys.0-event`
- [x] Power supply `/sys/class/power_supply`
- [x] LED `/sys/class/leds`
- [x] LCD `/dev/fb0`
- [x] Lego Port `/sys/class/lego-port`
- [x] Sensor `/sys/class/lego-sensor`
- [x] DC motor `/sys/class/dc-motor`
- [x] Linear actuator `/sys/class/tacho-motor`
- [x] Servo motor `/sys/class/servo-motor`
- [x] Tacho motor `/sys/class/tacho-motor`

### Common tasks

None yet.

LEGO® is a trademark of the LEGO Group of companies which does not sponsor, authorize or endorse this software.
