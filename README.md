# ev3 is an idiomatic Go interface to an ev3dev device

The goal is to implement a simple Go style ev3dev API and helpers for common tasks.

github.com/kortschak/ev3/ev3dev depends on an ev3dev kernel v3.16.7-ckt26-10-ev3dev-ev3 or better (See http://www.ev3dev.org/news/2016/04/11/Kernel-Release-Cycle-10/)

Currently supported:

## Low level API

- [x] Automatic identification of attached devices
- [ ] Buttons
- [x] Power supply `/sys/class/power_supply`
- [x] LED `/sys/class/leds`
- [x] LCD `/dev/fb0`
- [x] Lego Port `/sys/class/lego-port`
- [x] Sensor `/sys/class/lego-sensor`
- [x] DC motor `/sys/class/dc-motor`
- [x] Linear actuator `/sys/class/tacho-motor`
- [x] Servo motor `/sys/class/servo-motor`
- [x] Tacho motor `/sys/class/tacho-motor`

## Common tasks

None yet.

LEGO® is a trademark of the LEGO Group of companies which does not sponsor, authorize or endorse this software.