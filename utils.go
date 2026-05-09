package awsping

import (
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Version describes application version
const Version = "3.0.0"

var (
	github    = "https://github.com/ekalinin/awsping"
	useragent = fmt.Sprintf("AwsPing/%s (+%s)", Version, github)
)

const (
	// ShowOnlyRegions describes a type of output when only region's name and code printed out
	ShowOnlyRegions = -1
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Duration2ms converts time.Duration to ms (float64)
func Duration2ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1000 / 1000
}

// mkRandomString returns random string
func mkRandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}

// LatencyOutput prints data into console
type LatencyOutput struct {
	Level   int
	Repeats int
	w       io.Writer
}

// NewOutput creates a new LatencyOutput instance
func NewOutput(level, repeats int) *LatencyOutput {
	return &LatencyOutput{
		Level:   level,
		Repeats: repeats,
		w:       os.Stdout,
	}
}

func (lo *LatencyOutput) show(regions *AWSRegions) {
	for _, r := range *regions {
		fmt.Fprintf(lo.w, "%-15s %-s\n", r.Code, r.Name)
	}
}

func (lo *LatencyOutput) show0(regions *AWSRegions) {
	for _, r := range *regions {
		fmt.Fprintf(lo.w, "%-30s %20s\n", r.ShortName(), r.GetLatencyStr())
	}
}

func (lo *LatencyOutput) show1(regions *AWSRegions) {
	outFmt := "%5v %-15s %-30s %20s\n"
	fmt.Fprintf(lo.w, outFmt, "", "Code", "Region", "Latency")
	for i, r := range *regions {
		fmt.Fprintf(lo.w, outFmt, i, r.Code, r.Name, r.GetLatencyStr())
	}
}

func (lo *LatencyOutput) show2(regions *AWSRegions) {
	// format
	outFmt := "%5v %-15s %-30s"
	outFmt += strings.Repeat(" %15s", lo.Repeats) + " %15s\n"
	// header
	outStr := []any{"", "Code", "Region"}
	for i := range lo.Repeats {
		outStr = append(outStr, "Try #"+strconv.Itoa(i+1))
	}
	outStr = append(outStr, "Avg Latency")

	// show header
	fmt.Fprintf(lo.w, outFmt, outStr...)

	// each region stats
	for i, r := range *regions {
		outData := []any{strconv.Itoa(i), r.Code, r.Name}
		for n := range lo.Repeats {
			if n < len(r.Latencies) {
				outData = append(outData, fmt.Sprintf("%.2f ms",
					Duration2ms(r.Latencies[n])))
			} else {
				outData = append(outData, "-")
			}
		}
		outData = append(outData, r.GetLatencyStr())
		fmt.Fprintf(lo.w, outFmt, outData...)
	}
}

// Show print data
func (lo *LatencyOutput) Show(regions *AWSRegions) {
	switch lo.Level {
	case ShowOnlyRegions:
		lo.show(regions)
	case 0:
		lo.show0(regions)
	case 1:
		lo.show1(regions)
	case 2:
		lo.show2(regions)
	}
}

// GetRegions returns a list of regions
func GetRegions() AWSRegions {
	return AWSRegions{
		NewRegion("Africa (Cape Town)", "af-south-1", "Cape Town"),
		NewRegion("Asia Pacific (Hong Kong)", "ap-east-1", "Hong Kong"),
		NewRegion("Asia Pacific (Taipei)", "ap-east-2", "Taipei"),
		NewRegion("Asia Pacific (Tokyo)", "ap-northeast-1", "Tokyo"),
		NewRegion("Asia Pacific (Seoul)", "ap-northeast-2", "Seoul"),
		NewRegion("Asia Pacific (Osaka)", "ap-northeast-3", "Osaka"),
		NewRegion("Asia Pacific (Mumbai)", "ap-south-1", "Mumbai"),
		NewRegion("Asia Pacific (Hyderabad)", "ap-south-2", "Hyderabad"),
		NewRegion("Asia Pacific (Singapore)", "ap-southeast-1", "Singapore"),
		NewRegion("Asia Pacific (Sydney)", "ap-southeast-2", "Sydney"),
		NewRegion("Asia Pacific (Jakarta)", "ap-southeast-3", "Jakarta"),
		NewRegion("Asia Pacific (Melbourne)", "ap-southeast-4", "Melbourne"),
		NewRegion("Asia Pacific (Malaysia)", "ap-southeast-5", "Malaysia"),
		NewRegion("Asia Pacific (New Zealand)", "ap-southeast-6", "New Zealand"),
		NewRegion("Asia Pacific (Thailand)", "ap-southeast-7", "Thailand"),
		NewRegion("Canada (Central)", "ca-central-1", "Central"),
		NewRegion("Canada West (Calgary)", "ca-west-1", "Calgary"),
		NewRegion("Europe (Frankfurt)", "eu-central-1", "Frankfurt"),
		NewRegion("Europe (Zurich)", "eu-central-2", "Zurich"),
		NewRegion("Europe (Stockholm)", "eu-north-1", "Stockholm"),
		NewRegion("Europe (Milan)", "eu-south-1", "Milan"),
		NewRegion("Europe (Spain)", "eu-south-2", "Spain"),
		NewRegion("Europe (Ireland)", "eu-west-1", "Ireland"),
		NewRegion("Europe (London)", "eu-west-2", "London"),
		NewRegion("Europe (Paris)", "eu-west-3", "Paris"),
		NewRegion("Israel (Tel Aviv)", "il-central-1", "Tel Aviv"),
		NewRegion("Middle East (UAE)", "me-central-1", "UAE"),
		NewRegion("Middle East (Bahrain)", "me-south-1", "Bahrain"),
		NewRegion("Mexico (Central)", "mx-central-1", "Central"),
		NewRegion("South America (São Paulo)", "sa-east-1", "São Paulo"),
		NewRegion("US East (N. Virginia)", "us-east-1", "N. Virginia"),
		NewRegion("US East (Ohio)", "us-east-2", "Ohio"),
		NewRegion("US West (N. California)", "us-west-1", "N. California"),
		NewRegion("US West (Oregon)", "us-west-2", "Oregon"),
	}
}

// CalcLatency returns list of aws regions sorted by Latency
func CalcLatency(regions AWSRegions, repeats int, useHTTP bool, useHTTPS bool, service string) {
	regions.SetService(service)
	switch {
	case useHTTP:
		regions.SetCheckType(CheckTypeHTTP)
	case useHTTPS:
		regions.SetCheckType(CheckTypeHTTPS)
	default:
		regions.SetCheckType(CheckTypeTCP)
	}
	regions.SetDefaultTarget()

	var wg sync.WaitGroup
	for n := 1; n <= repeats; n++ {
		wg.Add(len(regions))
		for i := range regions {
			go regions[i].CheckLatency(&wg)
		}
		wg.Wait()
	}

	sort.Sort(regions)
}
