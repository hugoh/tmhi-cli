package pkg

import "fmt"

// SignalQuality represents the quality rating for a signal metric.
type SignalQuality string

// Signal quality rating constants.
const (
	// Excellent signal quality.
	Excellent SignalQuality = "Excellent"
	// Good signal quality.
	Good SignalQuality = "Good"
	// Fair signal quality.
	Fair SignalQuality = "Fair"
	// Poor signal quality.
	Poor SignalQuality = "Poor"
	// None represents no usable signal.
	None SignalQuality = "No Signal"
)

// SignalQualityRating represents a quality rating with its threshold boundaries.
type SignalQualityRating struct {
	MinValue float64
	MaxValue float64
	Rating   SignalQuality
}

// getRSRPRatings returns quality thresholds for RSRP (Reference Signal Received Power).
// RSRP measures the average power received from a single reference signal element.
// Values are in dBm (decibel-milliwatts).
//
// Sources and references:
//   - Powerful Signal (cellular signal booster manufacturer)
//   - Digi International (industrial cellular router manufacturer)
//   - Telco Antennas (professional antenna installation)
//   - 3GPP TS 36.133 defines measurement ranges (operator-specific thresholds)
//   - FreeRTOS Cellular Interface implementation
//
// Typical ranges: -44 dBm (excellent) to -140 dBm (no signal).
func getRSRPRatings() []SignalQualityRating {
	return []SignalQualityRating{
		{MinValue: -89, MaxValue: 0, Rating: Excellent},
		{MinValue: -104, MaxValue: -89, Rating: Good},
		{MinValue: -114, MaxValue: -104, Rating: Fair},
		{MinValue: -124, MaxValue: -114, Rating: Poor},
		{MinValue: -200, MaxValue: -124, Rating: None},
	}
}

// getRSRQRatings returns quality thresholds for RSRQ (Reference Signal Received Quality).
// RSRQ indicates the quality of the received signal and measures interference.
// Values are in dB (decibels).
//
// Sources and references:
//   - Powerful Signal
//   - 3GPP TS 36.133 defines measurement ranges (-43 dB to 20 dB for 5G)
//   - Industry practice: higher (less negative) values indicate better quality
//
// RSRQ = N * (RSRP / RSSI) where N is the number of Resource Blocks.
// Typical ranges: -3 dB (excellent) to -20 dB (poor).
func getRSRQRatings() []SignalQualityRating {
	return []SignalQualityRating{ //nolint:mnd
		{MinValue: -9, MaxValue: 20, Rating: Excellent},
		{MinValue: -14, MaxValue: -9, Rating: Good},
		{MinValue: -19, MaxValue: -14, Rating: Fair},
		{MinValue: -50, MaxValue: -19, Rating: Poor},
	}
}

// getRSSIRatings returns quality thresholds for RSSI (Received Signal Strength Indicator).
// RSSI represents the total received power including signal and noise.
// Values are in dBm (decibel-milliwatts).
//
// Sources and references:
//   - Digi International
//   - ESP-IDF WiFi RSSI thresholds
//   - RSSI is less commonly used in LTE/5G compared to RSRP
//
// Typical ranges: -50 dBm (excellent) to -110 dBm (poor).
func getRSSIRatings() []SignalQualityRating {
	return []SignalQualityRating{
		{MinValue: -65, MaxValue: 0, Rating: Excellent},
		{MinValue: -75, MaxValue: -65, Rating: Good},
		{MinValue: -85, MaxValue: -75, Rating: Fair},
		{MinValue: -120, MaxValue: -85, Rating: Poor},
	}
}

// getSINRRatings returns quality thresholds for SINR (Signal to Interference-plus-Noise Ratio).
// SINR measures the ratio of signal strength to interference plus noise.
// Values are in dB (decibels).
//
// Sources and references:
//   - Powerful Signal
//   - Nature Scientific Reports (2026)
//   - Higher values indicate cleaner signal with less noise
//
// SINR is a positive number in most contexts; negative values indicate
// the signal is weaker than the noise floor.
func getSINRRatings() []SignalQualityRating {
	return []SignalQualityRating{
		{MinValue: 13, MaxValue: 100, Rating: Excellent}, //nolint:mnd
		{MinValue: 6, MaxValue: 13, Rating: Good},        //nolint:mnd
		{MinValue: 0, MaxValue: 6, Rating: Fair},         //nolint:mnd
		{MinValue: -100, MaxValue: 0, Rating: Poor},
	}
}

// RateRSRP returns the quality rating for an RSRP value in dBm.
func RateRSRP(rsrp int) SignalQuality {
	return rateSignalValue(float64(rsrp), getRSRPRatings())
}

// RateRSRQ returns the quality rating for an RSRQ value in dB.
func RateRSRQ(rsrq int) SignalQuality {
	return rateSignalValue(float64(rsrq), getRSRQRatings())
}

// RateRSSI returns the quality rating for an RSSI value in dBm.
func RateRSSI(rssi int) SignalQuality {
	return rateSignalValue(float64(rssi), getRSSIRatings())
}

// RateSINR returns the quality rating for a SINR value in dB.
func RateSINR(sinr int) SignalQuality {
	return rateSignalValue(float64(sinr), getSINRRatings())
}

func rateSignalValue(value float64, ratings []SignalQualityRating) SignalQuality {
	for _, r := range ratings {
		if value >= r.MinValue && value < r.MaxValue {
			return r.Rating
		}
	}
	if value >= ratings[0].MaxValue {
		return ratings[0].Rating
	}
	return ratings[len(ratings)-1].Rating
}

// GetQualityEmoji returns an emoji representation of signal quality.
func GetQualityEmoji(quality SignalQuality) string {
	switch quality {
	case Excellent:
		return "★★★★★"
	case Good:
		return "★★★★☆"
	case Fair:
		return "★★★☆☆"
	case Poor:
		return "★★☆☆☆"
	case None:
		return "☆☆☆☆☆"
	default:
		return "???"
	}
}

// FormatSignalMetric formats a signal metric with its unit and quality rating.
// Returns a string like "RSRP: -92 dBm (Good ★★★★☆)".
func FormatSignalMetric(name string, value int, unit string, quality SignalQuality) string {
	emoji := GetQualityEmoji(quality)
	return fmt.Sprintf("%s: %d %s (%s %s)", name, value, unit, quality, emoji)
}
