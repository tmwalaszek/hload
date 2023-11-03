package time_formats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeToEpoch(t *testing.T) {
	times := []string{"01.02.2022",
		"01.02.22",
		"02/01/2022",
		"02/01/22",
		"02012022",
		"020122",
		"00:00_20220201",
		"20220201",
		"02/01/22",
	}

	for _, atTime := range times {
		epoch, err := TimeToEpoch(atTime)
		require.Nil(t, err)
		require.Equal(t, int64(1643670000), epoch)
	}

	brokenTimes := []string{"01.02.2022.01",
		"333.01.2022"}
	for _, brokenTime := range brokenTimes {
		_, err := TimeToEpoch(brokenTime)
		require.NotNil(t, err)
	}
}

func TestTimeFromDuration(t *testing.T) {
	faceNow := time.Date(2022, 2, 1, 10, 0, 0, 0, time.UTC)

	var tt = []struct {
		Duration string
		Expected int64
	}{
		{
			Duration: "1m",
			Expected: 1643709540,
		},
		{
			Duration: "1s",
			Expected: 1643709599,
		},
		{
			Duration: "1h",
			Expected: 1643706000,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Duration, func(t *testing.T) {
			actual, err := timeFromDuration(faceNow, tc.Duration)
			require.Nil(t, err)
			require.Equal(t, tc.Expected, actual)
		})
	}
}

func TestTimeFromBrokenDurationString(t *testing.T) {
	now := time.Now()

	broken := []string{
		"-1d",
		"1dd",
		"s1",
		"",
	}

	for _, b := range broken {
		t.Run(b, func(t *testing.T) {
			_, err := timeFromDuration(now, b)
			require.NotNil(t, err)
		})
	}
}
