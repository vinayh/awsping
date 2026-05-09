package awsping

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// defaultRequest is shared by every region constructed via NewRegion;
// http.Client and net.Dialer are safe for concurrent use.
var defaultRequest = NewAWSRequest()

// CheckType describes a type for a check
type CheckType int

const (
	// CheckTypeTCP is TCP type of check
	CheckTypeTCP CheckType = iota
	// CheckTypeHTTP is HTTP type of check
	CheckTypeHTTP
	// CheckTypeHTTPS is HTTPS type of check
	CheckTypeHTTPS
)

// --------------------------------------------

// AWSRegion description of the AWS EC2 region
type AWSRegion struct {
	Name      string
	Code      string
	Location  string
	Service   string
	Latencies []time.Duration
	Error     error
	CheckType CheckType

	Target  Targetter
	Request Requester
}

// NewRegion creates a new region. Location is the city/state/country shown
// next to the region code in the default output (e.g. "Calgary", "Bahrain",
// "N. Virginia").
func NewRegion(name, code, location string) AWSRegion {
	return AWSRegion{
		Name:      name,
		Code:      code,
		Location:  location,
		CheckType: CheckTypeTCP,
		Request:   defaultRequest,
	}
}

// CheckLatency does a latency check for a region
func (r *AWSRegion) CheckLatency(wg *sync.WaitGroup) {
	defer wg.Done()

	if r.CheckType == CheckTypeHTTP || r.CheckType == CheckTypeHTTPS {
		r.checkLatencyHTTP()
	} else {
		r.checkLatencyTCP()
	}
}

// checkLatencyHTTP Test Latency via HTTP
func (r *AWSRegion) checkLatencyHTTP() {
	url := r.Target.GetURL()
	l, err := r.Request.Do(useragent, url, RequestTypeHTTP)
	if err != nil {
		r.Error = err
		return
	}
	r.Latencies = append(r.Latencies, l)
}

// checkLatencyTCP Test Latency via TCP
func (r *AWSRegion) checkLatencyTCP() {
	tcpAddr, err := r.Target.GetIP()
	if err != nil {
		r.Error = err
		return
	}

	l, err := r.Request.Do(useragent, tcpAddr.String(), RequestTypeTCP)
	if err != nil {
		r.Error = err
		return
	}
	r.Latencies = append(r.Latencies, l)
}

// ShortName returns the region as "<code> (<location>)"
// (e.g. "ca-west-1 (Calgary)", "me-south-1 (Bahrain)").
func (r *AWSRegion) ShortName() string {
	return fmt.Sprintf("%s (%s)", r.Code, r.Location)
}

// GetLatency returns the average latency across successful tries (ms).
// Returns 0 if no try succeeded.
func (r *AWSRegion) GetLatency() float64 {
	if len(r.Latencies) == 0 {
		return 0
	}
	sum := float64(0)
	for _, l := range r.Latencies {
		sum += Duration2ms(l)
	}
	return sum / float64(len(r.Latencies))
}

// GetLatencyStr returns the latency string. If at least one try succeeded,
// it returns the average; otherwise it returns the most recent error.
func (r *AWSRegion) GetLatencyStr() string {
	if len(r.Latencies) == 0 && r.Error != nil {
		return shortErr(r.Error)
	}
	return fmt.Sprintf("%.2f ms", r.GetLatency())
}

// shortErr returns a concise, column-friendly form of a network error,
// stripping the noisy "dial tcp <addr>:" prefix from net.OpError.
func shortErr(err error) string {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "timeout"
	}
	var oe *net.OpError
	if errors.As(err, &oe) && oe.Err != nil {
		return oe.Err.Error()
	}
	return err.Error()
}

// --------------------------------------------

// AWSRegions slice of the AWSRegion
type AWSRegions []AWSRegion

// Len returns a count of regions
func (rs AWSRegions) Len() int {
	return len(rs)
}

// Less compares two regions by latency. Regions with no successful tries
// sort last so errors appear after live regions.
func (rs AWSRegions) Less(i, j int) bool {
	iEmpty := len(rs[i].Latencies) == 0
	jEmpty := len(rs[j].Latencies) == 0
	if iEmpty != jEmpty {
		return !iEmpty
	}
	return rs[i].GetLatency() < rs[j].GetLatency()
}

// Swap two regions by index
func (rs AWSRegions) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// SetService sets service for all regions
func (rs AWSRegions) SetService(service string) {
	for i := range rs {
		rs[i].Service = service
	}
}

// SetCheckType sets Check Type for all regions
func (rs AWSRegions) SetCheckType(checkType CheckType) {
	for i := range rs {
		rs[i].CheckType = checkType
	}
}

// SetDefaultTarget sets default target instance
func (rs AWSRegions) SetDefaultTarget() {
	rs.SetTarget(func(r *AWSRegion) {
		r.Target = &AWSTarget{
			HTTPS:   r.CheckType == CheckTypeHTTPS,
			Code:    r.Code,
			Service: r.Service,
			Rnd:     mkRandomString(13),
		}
	})
}

// SetTarget applies fn to every region, allowing callers to install a custom Target.
func (rs AWSRegions) SetTarget(fn func(r *AWSRegion)) {
	for i := range rs {
		fn(&rs[i])
	}
}
