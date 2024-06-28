package vdd

import (
	"device/nrf"
	"runtime/volatile"
	"unsafe"
)

func init() {
	nrf.SAADC.ENABLE.Set(nrf.SAADC_ENABLE_ENABLE_Enabled << nrf.SAADC_ENABLE_ENABLE_Pos)
	nrf.SAADC.RESOLUTION.Set(nrf.SAADC_RESOLUTION_VAL_12bit)
	nrf.SAADC.OVERSAMPLE.Set(nrf.SAADC_OVERSAMPLE_OVERSAMPLE_Bypass)
	nrf.SAADC.CH[0].CONFIG.Set(
		(nrf.SAADC_CH_CONFIG_GAIN_Gain1_6 << nrf.SAADC_CH_CONFIG_GAIN_Pos) |
			(nrf.SAADC_CH_CONFIG_MODE_SE << nrf.SAADC_CH_CONFIG_MODE_Pos) |
			(nrf.SAADC_CH_CONFIG_REFSEL_Internal << nrf.SAADC_CH_CONFIG_REFSEL_Pos) |
			(nrf.SAADC_CH_CONFIG_RESN_Bypass << nrf.SAADC_CH_CONFIG_RESN_Pos) |
			(nrf.SAADC_CH_CONFIG_RESP_Bypass << nrf.SAADC_CH_CONFIG_RESP_Pos) |
			(nrf.SAADC_CH_CONFIG_TACQ_3us << nrf.SAADC_CH_CONFIG_TACQ_Pos),
	)
	nrf.SAADC.CH[0].SetPSELP(nrf.SAADC_CH_PSELP_PSELP_VDD << nrf.SAADC_CH_PSELP_PSELP_Pos)
	nrf.SAADC.CH[0].SetPSELN(nrf.SAADC_CH_PSELN_PSELN_NC << nrf.SAADC_CH_PSELN_PSELN_Pos)
}

func Measure() float32 {
	var rawValue volatile.Register16
	// Destination for sample result.
	nrf.SAADC.RESULT.PTR.Set(uint32(uintptr(unsafe.Pointer(&rawValue))))
	nrf.SAADC.RESULT.MAXCNT.Set(1) // One sample

	// Start tasks.
	nrf.SAADC.TASKS_START.Set(1)
	for nrf.SAADC.EVENTS_STARTED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STARTED.Set(0x00)

	// Start the sample task.
	nrf.SAADC.TASKS_SAMPLE.Set(1)

	// Wait until the sample task is done.
	for nrf.SAADC.EVENTS_END.Get() == 0 {
	}
	nrf.SAADC.EVENTS_END.Set(0x00)

	// Stop the ADC
	nrf.SAADC.TASKS_STOP.Set(1)
	for nrf.SAADC.EVENTS_STOPPED.Get() == 0 {
	}
	nrf.SAADC.EVENTS_STOPPED.Set(0)

	value := int16(rawValue.Get())
	if value < 0 {
		value = 0
	}
	// Result = [V(p) - V(n)] * GAIN/REFERENCE * 2^(RESOLUTION)
	// Result = (VDD - 0) * (1./6) / 0.6 * 2**12
	// VDD = Result / 1137.7777777777778
	return float32(value&0xfff) / 1137.7777777777778
}
