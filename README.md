<img alt="Gopherbrick" src="https://raw.githubusercontent.com/ev3go/ev3/logo/gopherbrick.png" width="200">
# ev3 is an idiomatic Go interface to an ev3dev device

[![Build Status](https://travis-ci.org/ev3go/ev3.svg?branch=master)](https://travis-ci.org/ev3go/ev3) [![Coverage Status](https://coveralls.io/repos/ev3go/ev3/badge.svg?branch=master&service=github)](https://coveralls.io/github/ev3go/ev3?branch=master) [![GoDoc](https://godoc.org/github.com/ev3go/ev3?status.svg)](https://godoc.org/github.com/ev3go/ev3)

The goal is to implement a simple Go style ev3dev API and helpers for common tasks.

github.com/ev3go/ev3/ev3dev depends on an ev3dev kernel v3.16.7-ckt26-10-ev3dev-ev3 or better (See http://www.ev3dev.org/news/2016/04/11/Kernel-Release-Cycle-10/)

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
