package internal

import (
	"errors"
	"testing"

	signal "github.com/hugoh/cellular-signal"
	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

func TestDisplayLoginResult(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

	t.Run("success", func(t *testing.T) {
		result := &tmhi.LoginResult{Success: true}

		assert.NotPanics(t, func() { displayLoginResult(result) })
	})

	t.Run("failure", func(t *testing.T) {
		result := &tmhi.LoginResult{Success: false}

		assert.NotPanics(t, func() { displayLoginResult(result) })
	})
}

func TestDisplayStatusResult(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

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
			Registration:   "registered",
		}

		assert.NotPanics(t, func() { displayStatusResult(result) })
	})
}

func TestDisplaySignalResult(t *testing.T) {
	pterm.DisableStyling()

	defer pterm.EnableStyling()

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
				APN:          "test.apn",
				HasIPv6:      false,
				Registration: "registered",
				Roaming:      true,
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

	defer pterm.EnableStyling()

	result := &tmhi.InfoResult{}

	assert.NotPanics(t, func() { displayInfoResult(result) })
}

func TestFormatMetricValue(t *testing.T) {
	rater := signal.NewRater()
	result := formatMetricValue(rater, -100, rater.RateRSRP)
	assert.Contains(t, result, "-100")
}

func TestFormatMetricQuality(t *testing.T) {
	rating := signal.NewRater().RateRSRP(-100)
	result := formatMetricQuality(rating)
	assert.NotEmpty(t, result)
}
