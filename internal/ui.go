package internal

import (
	"fmt"
	"strconv"

	signal "github.com/hugoh/cellular-signal"
	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
)

func displayLoginResult(result *tmhi.LoginResult) {
	if result.Success {
		pterm.Success.Println("Successfully logged in")
	}
}

func displayStatusResult(result *tmhi.StatusResult) {
	switch {
	case result.WebInterfaceUp:
		pterm.Success.Println("Web interface up")
	case result.Error != nil:
		pterm.Error.Println("Web interface down: " + result.Error.Error())
	default:
		pterm.Error.Println(fmt.Sprintf("Web interface down: status %d", result.StatusCode))
	}

	if result.Registration != "" {
		pterm.Info.Println("Registration status: " + result.Registration)
	}
}

func displaySignalResult(result *tmhi.SignalResult) {
	if result.FourG != nil {
		displaySignalMetrics("4G LTE Signal", &result.FourG.SignalData,
			[]string{"eNBID", strconv.Itoa(result.FourG.ENBID)})
	}

	if result.FiveG != nil {
		extras := [][]string{}
		if result.FiveG.AntennaUsed != "" {
			extras = append(extras, []string{"Antenna", result.FiveG.AntennaUsed})
		}

		extras = append(extras, []string{"gNBID", strconv.Itoa(result.FiveG.GNBID)})
		displaySignalMetrics("5G Signal", &result.FiveG.SignalData, extras...)
	}

	displayGenericSignalInfo(result)
}

const signalMetricsCount = 6

func displaySignalMetrics(header string, metrics *tmhi.SignalData, extras ...[]string) {
	rater := signal.NewRater()

	pterm.DefaultHeader.Println(header)

	tableData := make(pterm.TableData, 0, 2+len(extras)+signalMetricsCount)
	tableData = append(tableData,
		[]string{"Metric", "Value", "Rating"},
		[]string{"Signal bars", fmt.Sprintf("%.0f", metrics.Bars), ""},
	)

	for _, extra := range extras {
		tableData = append(tableData, append(extra, ""))
	}

	tableData = append(
		tableData,
		[]string{"Bands", fmt.Sprintf("%v", metrics.Bands), ""},
		[]string{
			"RSRP",
			formatMetricValue(rater, metrics.RSRP, rater.RateRSRP),
			formatMetricQuality(rater.RateRSRP(metrics.RSRP)),
		},
		[]string{
			"RSRQ",
			formatMetricValue(rater, metrics.RSRQ, rater.RateRSRQ),
			formatMetricQuality(rater.RateRSRQ(metrics.RSRQ)),
		},
		[]string{
			"RSSI",
			formatMetricValue(rater, metrics.RSSI, rater.RateRSSI),
			formatMetricQuality(rater.RateRSSI(metrics.RSSI)),
		},
		[]string{
			"SINR",
			formatMetricValue(rater, metrics.SINR, rater.RateSINR),
			formatMetricQuality(rater.RateSINR(metrics.SINR)),
		},
		[]string{"CID", strconv.Itoa(metrics.CID), ""},
	)

	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		pterm.Error.Println("Failed to render table:", err)
	}
}

func displayGenericSignalInfo(result *tmhi.SignalResult) {
	pterm.DefaultHeader.Println("Generic Info")

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"APN", result.Generic.APN},
		{"IPv6", strconv.FormatBool(result.Generic.HasIPv6)},
		{"Registration", result.Generic.Registration},
		{"Roaming", strconv.FormatBool(result.Generic.Roaming)},
	}

	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		pterm.Error.Println("Failed to render table:", err)
	}
}

func formatMetricValue(rater *signal.Rater, value int, rateFunc func(int) signal.Rating) string {
	return rater.FormatWith("%v %u", rateFunc(value))
}

func formatMetricQuality(rating signal.Rating) string {
	rater := signal.NewRater()

	return rater.FormatWith("%q %s", rating)
}

func displayInfoResult(result *tmhi.InfoResult) {
	pterm.DefaultBasicText.Println(result.String())
}
