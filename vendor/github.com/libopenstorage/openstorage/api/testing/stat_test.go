package testing

import (
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

func TestZeroOrSmallStats(t *testing.T) {
	stats := []*api.Stats{
		&api.Stats{},
		&api.Stats{
			IntervalMs: 999,
		},
	}
	zero := uint64(0)
	for _, s := range stats {
		require.Equal(t, zero, s.WriteLatency(), "nil write latency")
		require.Equal(t, zero, s.ReadLatency(), "nil read latency")
		require.Equal(t, zero, s.Latency(), "nil latency")
		require.Equal(t, zero, s.WriteThroughput(), "nil write througput")
		require.Equal(t, zero, s.ReadThroughput(), "nil read througput")
		require.Equal(t, zero, s.Iops(), "nil iops")
	}
}

func TestNonZeroStats(t *testing.T) {
	s := &api.Stats{
		WriteBytes: 20,
		ReadBytes:  10,
		Reads:      10,
		Writes:     20,
		ReadMs:     1,
		WriteMs:    2,
		IoMs:       3,
		IntervalMs: 2000,
	}
	require.Equal(t, uint64(100), s.WriteLatency(), "write latency")
	require.Equal(t, uint64(100), s.ReadLatency(), "read latency")
	require.Equal(t, uint64(100), s.Latency(), "latency")
	require.Equal(t, uint64(10), s.WriteThroughput(), "write througput")
	require.Equal(t, uint64(5), s.ReadThroughput(), "read througput")
	require.Equal(t, uint64(15), s.Iops(), "iops")
}
