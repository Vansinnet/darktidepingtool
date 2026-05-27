package main

import (
	"net"
	"sort"
	"time"
)

type PingResult struct {
	Region     Region
	Median     int
	Average    int
	Min        int
	Max        int
	P10        int
	P90        int
	Jitter     int
	PacketLoss float64
	Samples    int
	Failed     int
	Error      string
}

const (
	pingSamples = 10
	pingTimeout = 3 * time.Second
	pingDelay   = 80 * time.Millisecond
)

func pingBeacon(beacon string, timeout time.Duration) (time.Duration, error) {
	addr, err := net.ResolveUDPAddr("udp", beacon)
	if err != nil {
		return 0, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}

	pingData := []byte{0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	start := time.Now()
	if _, err := conn.Write(pingData); err != nil {
		return 0, err
	}

	buf := make([]byte, 256)
	_, err = conn.Read(buf)
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

func measureRegion(region Region, samples int, timeout time.Duration) PingResult {
	var durations []float64
	var failed int

	for i := 0; i < samples; i++ {
		rtt, err := pingBeacon(region.Beacon, timeout)
		if err != nil {
			failed++
			continue
		}
		durations = append(durations, float64(rtt.Milliseconds()))
		time.Sleep(pingDelay)
	}

	if len(durations) == 0 {
		return PingResult{
			Region:     region,
			Median:     9999,
			Average:    9999,
			Min:        9999,
			Max:        9999,
			P10:        9999,
			P90:        9999,
			Jitter:     0,
			PacketLoss: 100,
			Samples:    0,
			Failed:     failed,
			Error:      "all pings failed",
		}
	}

	sort.Float64s(durations)

	n := len(durations)
	median := int(durations[n/2])
	minVal := int(durations[0])
	maxVal := int(durations[n-1])

	var sum float64
	for _, d := range durations {
		sum += d
	}
	average := int(sum / float64(n))

	p10 := int(durations[max(0, n*1/10)])
	p90 := int(durations[min(n-1, n*9/10)])

	jitter := 0
	if n > 1 {
		diffs := make([]float64, 0, n-1)
		for i := 1; i < n; i++ {
			d := durations[i] - durations[i-1]
			if d < 0 {
				d = -d
			}
			diffs = append(diffs, d)
		}
		sumDiffs := 0.0
		for _, d := range diffs {
			sumDiffs += d
		}
		jitter = int(sumDiffs / float64(len(diffs)))
	}

	packetLoss := float64(failed) / float64(samples) * 100

	return PingResult{
		Region:     region,
		Median:     median,
		Average:    average,
		Min:        minVal,
		Max:        maxVal,
		P10:        p10,
		P90:        p90,
		Jitter:     jitter,
		PacketLoss: packetLoss,
		Samples:    n,
		Failed:     failed,
	}
}
