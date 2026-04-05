package pkg

import (
	"testing"
)

func TestRateRSRP(t *testing.T) {
	tests := []struct {
		name     string
		rsrp     int
		expected SignalQuality
	}{
		{"Excellent signal", -80, Excellent},
		{"Excellent boundary", -89, Excellent},
		{"Good signal upper", -90, Good},
		{"Good signal middle", -95, Good},
		{"Good signal lower", -104, Good},
		{"Fair signal upper", -105, Fair},
		{"Fair signal middle", -110, Fair},
		{"Fair signal lower", -114, Fair},
		{"Poor signal upper", -115, Poor},
		{"Poor signal middle", -120, Poor},
		{"Poor signal lower", -124, Poor},
		{"No signal", -130, None},
		{"Very poor signal", -140, None},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RateRSRP(tt.rsrp)
			if result != tt.expected {
				t.Errorf("RateRSRP(%d) = %s, expected %s", tt.rsrp, result, tt.expected)
			}
		})
	}
}

func TestRateRSRQ(t *testing.T) {
	tests := []struct {
		name     string
		rsrq     int
		expected SignalQuality
	}{
		{"Excellent signal", -5, Excellent},
		{"Excellent boundary", -9, Excellent},
		{"Good signal upper", -10, Good},
		{"Good signal middle", -12, Good},
		{"Good signal lower", -14, Good},
		{"Fair signal upper", -15, Fair},
		{"Fair signal middle", -17, Fair},
		{"Fair signal lower", -19, Fair},
		{"Poor signal", -20, Poor},
		{"Very poor signal", -30, Poor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RateRSRQ(tt.rsrq)
			if result != tt.expected {
				t.Errorf("RateRSRQ(%d) = %s, expected %s", tt.rsrq, result, tt.expected)
			}
		})
	}
}

func TestRateRSSI(t *testing.T) {
	tests := []struct {
		name     string
		rssi     int
		expected SignalQuality
	}{
		{"Excellent signal", -50, Excellent},
		{"Excellent boundary", -65, Excellent},
		{"Good signal upper", -70, Good},
		{"Good signal boundary", -66, Good},
		{"Good signal lower", -75, Good},
		{"Fair signal upper", -80, Fair},
		{"Fair signal lower", -85, Fair},
		{"Poor signal", -90, Poor},
		{"Very poor signal", -100, Poor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RateRSSI(tt.rssi)
			if result != tt.expected {
				t.Errorf("RateRSSI(%d) = %s, expected %s", tt.rssi, result, tt.expected)
			}
		})
	}
}

func TestRateSINR(t *testing.T) {
	tests := []struct {
		name     string
		sinr     int
		expected SignalQuality
	}{
		{"Excellent signal", 20, Excellent},
		{"Excellent boundary", 13, Excellent},
		{"Good signal upper", 10, Good},
		{"Good signal middle", 8, Good},
		{"Good signal lower", 6, Good},
		{"Fair signal upper", 5, Fair},
		{"Fair signal middle", 3, Fair},
		{"Fair signal lower", 0, Fair},
		{"Poor signal", -5, Poor},
		{"Very poor signal", -20, Poor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RateSINR(tt.sinr)
			if result != tt.expected {
				t.Errorf("RateSINR(%d) = %s, expected %s", tt.sinr, result, tt.expected)
			}
		})
	}
}

func TestGetQualityEmoji(t *testing.T) {
	tests := []struct {
		quality  SignalQuality
		expected string
	}{
		{Excellent, "★★★★★"},
		{Good, "★★★★☆"},
		{Fair, "★★★☆☆"},
		{Poor, "★★☆☆☆"},
		{None, "☆☆☆☆☆"},
		{SignalQuality("unknown"), "???"},
	}

	for _, tt := range tests {
		t.Run(string(tt.quality), func(t *testing.T) {
			result := GetQualityEmoji(tt.quality)
			if result != tt.expected {
				t.Errorf("GetQualityEmoji(%s) = %s, expected %s", tt.quality, result, tt.expected)
			}
		})
	}
}

func TestFormatSignalMetric(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		value    int
		unit     string
		quality  SignalQuality
		expected string
	}{
		{
			name:     "RSRP Good",
			metric:   "RSRP",
			value:    -92,
			unit:     "dBm",
			quality:  Good,
			expected: "RSRP: -92 dBm (Good ★★★★☆)",
		},
		{
			name:     "SINR Excellent",
			metric:   "SINR",
			value:    15,
			unit:     "dB",
			quality:  Excellent,
			expected: "SINR: 15 dB (Excellent ★★★★★)",
		},
		{
			name:     "RSRQ Poor",
			metric:   "RSRQ",
			value:    -22,
			unit:     "dB",
			quality:  Poor,
			expected: "RSRQ: -22 dB (Poor ★★☆☆☆)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSignalMetric(tt.metric, tt.value, tt.unit, tt.quality)
			if result != tt.expected {
				t.Errorf("FormatSignalMetric() = %s, expected %s", result, tt.expected)
			}
		})
	}
}
