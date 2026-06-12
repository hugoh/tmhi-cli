package internal

import (
	"fmt"
	"strconv"

	signal "github.com/hugoh/cellular-signal/v2"
	tmhi "github.com/hugoh/tmhi-gateway"
	"github.com/pterm/pterm"
)

func displayStatusResult(result *tmhi.StatusResult) {
	switch {
	case result.WebInterfaceUp:
		pterm.Success.Println("Web interface up")
	case result.Error != nil:
		pterm.Error.Println("Web interface down: " + result.Error.Error())
	default:
		pterm.Error.Printfln("Web interface down: status %d", result.StatusCode)
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
		var extras [][]string
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

	tableData = append(tableData, []string{"Bands", fmt.Sprintf("%v", metrics.Bands), ""})

	ratedMetrics := []struct {
		name  string
		value int
		rate  func(float64) signal.Rating
	}{
		{"RSRP", metrics.RSRP, rater.RateRSRP},
		{"RSRQ", metrics.RSRQ, rater.RateRSRQ},
		{"RSSI", metrics.RSSI, rater.RateRSSI},
		{"SINR", metrics.SINR, rater.RateSINR},
	}
	for _, metric := range ratedMetrics {
		rating := metric.rate(float64(metric.value))
		tableData = append(tableData, []string{
			metric.name,
			rating.Format("%v %u"),
			rating.Format("%q %s"),
		})
	}

	tableData = append(tableData, []string{"CID", strconv.Itoa(metrics.CID), ""})

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

func displayInfoResult(result *tmhi.InfoResult) {
	pterm.DefaultBasicText.Println(result.String())
}
