// Copyright ©2016 The ev3go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ev3dev

var Prefix string

func init() {
	isTesting = true

	prefix = "testmount"
	Prefix = prefix

	// We cannot use poll(2) for waiting on motor state attribute in testing.
	canPoll = false
}

var StateIsOK = stateIsOK

type mockDevice struct{}

func (d mockDevice) Path() string   { return "path" }
func (d mockDevice) Type() string   { return "mock" }
func (d mockDevice) Err() error     { return nil }
func (d mockDevice) String() string { return "mock" }

const (
	AddressName                   = address
	BinDataName                   = binData
	BatteryTechnologyName         = batteryTechnology
	BatteryTypeName               = batteryType
	BinDataFormatName             = binDataFormat
	BrightnessName                = brightness
	CommandName                   = command
	CommandsName                  = commands
	CountPerMeterName             = countPerMeter
	CountPerRotName               = countPerRot
	CurrentNowName                = currentNow
	DecimalsName                  = decimals
	DelayOffName                  = delayOff
	DelayOnName                   = delayOn
	DirectName                    = direct
	DriverNameName                = driverName
	DutyCycleName                 = dutyCycle
	DutyCycleSetpointName         = dutyCycleSetpoint
	FirmwareVersion               = firmwareVersion
	FullTravelCountName           = fullTravelCount
	HoldPIDName                   = holdPID
	HoldPIDkdName                 = holdPIDkd
	HoldPIDkiName                 = holdPIDki
	HoldPIDkpName                 = holdPIDkp
	KdName                        = kd
	KiName                        = ki
	KpName                        = kp
	MaxBrightnessName             = maxBrightness
	MaxPulseSetpointName          = maxPulseSetpoint
	MidPulseSetpointName          = midPulseSetpoint
	MinPulseSetpointName          = minPulseSetpoint
	MaxSpeedName                  = maxSpeed
	ModeName                      = mode
	ModesName                     = modes
	NumValuesName                 = numValues
	PolarityName                  = polarity
	PollRateName                  = pollRate
	PositionName                  = position
	PositionSetpointName          = positionSetpoint
	PowerName                     = power
	PowerAutosuspendDelayName     = powerAutosuspendDelay
	PowerControlName              = powerControl
	PowerRuntimeActiveTimeName    = powerRuntimeActiveTime
	PowerRuntimeStatusName        = powerRuntimeStatus
	PowerRuntimeSuspendedTimeName = powerRuntimeSuspendedTime
	RampDownSetpointName          = rampDownSetpoint
	RampUpSetpointName            = rampUpSetpoint
	RateSetpointName              = rateSetpoint
	SetDeviceName                 = setDevice
	SpeedName                     = speed
	SpeedPIDName                  = speedPID
	SpeedPIDkdName                = speedPIDkd
	SpeedPIDkiName                = speedPIDki
	SpeedPIDkpName                = speedPIDkp
	SpeedSetpointName             = speedSetpoint
	StateName                     = state
	StatusName                    = status
	StopActionName                = stopAction
	StopActionsName               = stopActions
	SubsystemName                 = subsystem
	TextValuesName                = textValues
	TimeSetpointName              = timeSetpoint
	TriggerName                   = trigger
	UeventName                    = uevent
	UnitsName                     = units
	ValueName                     = value
	VoltageMaxDesignName          = voltageMaxDesign
	VoltageMinDesignName          = voltageMinDesign
	VoltageNowName                = voltageNow
)
