package internal

import (
	"bytes"
	"errors"
	"os"
	"testing"

	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestDisplayStatusResult(t *testing.T) {
	pterm.DisableStyling()
	t.Cleanup(pterm.EnableStyling)

	t.Run("web interface up", func(t *testing.T) {
		result := &tmhi.StatusResult{WebInterfaceUp: true}

		assert.NotPanics(t, func() { displayStatusResult(result) })
	})

	t.Run("web interface down with error", func(t *testing.T) {
		result := &tmhi.StatusResult{
			WebInterfaceUp: false,
			Error:          errors.New("connection refused"),
		}

		assert.NotPanics(t, func() { displayStatusResult(result) })
	})

	t.Run("web interface down with status code", func(t *testing.T) {
		result := &tmhi.StatusResult{
			WebInterfaceUp: false,
			StatusCode:     503,
		}

		assert.NotPanics(t, func() { displayStatusResult(result) })
	})

	t.Run("with registration status", func(t *testing.T) {
		result := &tmhi.StatusResult{
			WebInterfaceUp: true,
			Registration:   testRegState,
		}

		assert.NotPanics(t, func() { displayStatusResult(result) })
	})
}

func TestDisplaySignalResult(t *testing.T) {
	pterm.DisableStyling()
	t.Cleanup(pterm.EnableStyling)

	t.Run("with 4G signal", func(t *testing.T) {
		result := &tmhi.SignalResult{
			FourG: &tmhi.FourGSignal{
				ENBID: 12345,
				SignalData: tmhi.SignalData{
					Bars:  4,
					Bands: []string{"B12"},
					RSRP:  -100,
					RSRQ:  -10,
					RSSI:  -70,
					SINR:  10,
					CID:   1001,
				},
			},
			Generic: tmhi.GenericSignalInfo{
				APN:          "test.apn",
				HasIPv6:      true,
				Registration: "registered",
				Roaming:      false,
			},
		}

		assert.NotPanics(t, func() { displaySignalResult(result) })
	})

	t.Run("with 5G signal and antenna", func(t *testing.T) {
		result := &tmhi.SignalResult{
			FiveG: &tmhi.FiveGSignal{
				GNBID:       67890,
				AntennaUsed: "external",
				SignalData: tmhi.SignalData{
					Bars:  5,
					Bands: []string{"n41"},
					RSRP:  -90,
					RSRQ:  -8,
					RSSI:  -60,
					SINR:  15,
					CID:   2001,
				},
			},
			Generic: tmhi.GenericSignalInfo{
				APN:          testAPN,
				HasIPv6:      true,
				Registration: testRegState,
				Roaming:      false,
			},
		}

		assert.NotPanics(t, func() { displaySignalResult(result) })
	})

	t.Run("with both 4G and 5G", func(t *testing.T) {
		result := &tmhi.SignalResult{
			FourG: &tmhi.FourGSignal{
				ENBID: 12345,
				SignalData: tmhi.SignalData{
					Bars:  3,
					Bands: []string{"B4"},
					RSRP:  -105,
					RSRQ:  -12,
					RSSI:  -75,
					SINR:  8,
					CID:   1001,
				},
			},
			FiveG: &tmhi.FiveGSignal{
				GNBID:       67890,
				AntennaUsed: "",
				SignalData: tmhi.SignalData{
					Bars:  4,
					Bands: []string{"n71"},
					RSRP:  -95,
					RSRQ:  -9,
					RSSI:  -65,
					SINR:  12,
					CID:   2001,
				},
			},
			Generic: tmhi.GenericSignalInfo{
				APN:          "test.apn",
				HasIPv6:      true,
				Registration: "registered",
				Roaming:      false,
			},
		}

		assert.NotPanics(t, func() { displaySignalResult(result) })
	})
}

func TestDisplayInfoResult(t *testing.T) {
	pterm.DisableStyling()
	t.Cleanup(pterm.EnableStyling)

	result := &tmhi.InfoResult{}

	assert.NotPanics(t, func() { displayInfoResult(result) })
}

func TestDisplaySignalMetrics_Output(t *testing.T) {
	pterm.DisableStyling()
	t.Cleanup(pterm.EnableStyling)

	var buf bytes.Buffer

	pterm.SetDefaultOutput(&buf)
	t.Cleanup(func() { pterm.SetDefaultOutput(os.Stdout) })

	displaySignalMetrics("Test Signal", &tmhi.SignalData{
		Bars:  3,
		Bands: []string{"n41"},
		RSRP:  -100,
		RSRQ:  -10,
		RSSI:  -70,
		SINR:  10,
		CID:   42,
	})

	for _, want := range []string{"RSRP", "RSRQ", "RSSI", "SINR", "-100", "42"} {
		assert.Contains(t, buf.String(), want)
	}
}
