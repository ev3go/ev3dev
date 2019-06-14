![Gopherbrick](gopherbrick.png)
# ev3dev is an idiomatic Go interface to an [ev3dev](http://ev3dev.org) device

[![Build Status](https://travis-ci.org/ev3go/ev3dev.svg?branch=master)](https://travis-ci.org/ev3go/ev3dev) [![Coverage Status](https://coveralls.io/repos/ev3go/ev3dev/badge.svg?branch=master&service=github)](https://coveralls.io/github/ev3go/ev3dev?branch=master) [![GoDoc](https://godoc.org/github.com/ev3go/ev3dev?status.svg)](https://godoc.org/github.com/ev3go/ev3dev)

The goal is to implement a simple Go style ev3dev API and helpers for common tasks.

github.com/ev3go/ev3dev depends on ev3dev stretch and has been tested on kernel 4.14.61-ev3dev-2.2.2-ev3. For jessie support see the [ev3dev-jessie branch](https://github.com/ev3go/ev3dev/tree/ev3dev-jessie).

For device-specific functions see [EV3](https://github.com/ev3go/ev3) and [BrickPi](https://github.com/ev3go/brickpi).

## Currently supported:

### Low level API

- [x] Automatic identification of attached devices
- [x] Buttons `/dev/input/by-path/platform-gpio_keys-event`
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

- [x] Steering helper similar to EV-G steering block

## Quick start compiling for a brick

Compiling for a brick can be done on the platform itself if Go is installed there, but it is generally quicker on your computer. This requires that you prefix the `go build` invocation with `GOOS=linux GOARCH=arm GOARM=5`. For example, to build the [demo program](https://github.com/ev3go/ev3dev/tree/master/examples/demo) you can do this:

```
$ GOOS=linux GOARCH=arm GOARM=5 go build github.com/ev3go/ev3dev/examples/demo
```

This will leave a `demo` executable (from the name of the package path) in your current directory. You can then copy the executable over to your brick using `scp`.

---

LEGO® is a trademark of the LEGO Group of companies which does not sponsor, authorize or endorse this software.

