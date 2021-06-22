package bufiox

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScannerShortString(t *testing.T) {
	require := require.New(t)

	scanner := NewScanner(strings.NewReader(`a short string
followed by another short string`))

	// Does not start with an error.
	require.NoError(scanner.Err())

	// Reads the first line.
	require.True(scanner.Scan())
	require.Equal([]byte("a short string"), scanner.Bytes())
	require.Equal("a short string", scanner.Text())
	require.NoError(scanner.Err())

	// Reads the second line.
	require.True(scanner.Scan())
	require.Equal([]byte("followed by another short string"), scanner.Bytes())
	require.Equal("followed by another short string", scanner.Text())
	require.NoError(scanner.Err())

	// Finishes without error.
	require.False(scanner.Scan())
	require.NoError(scanner.Err())
}

func TestScannerLongString(t *testing.T) {
	require := require.New(t)

	// Generate some long strings for testing that
	// require multiple buffers to read.
	var str1 string
	var str2 string
	for len(str1) < bufferSizeBytes*2+bufferSizeBytes/2 {
		str1 += "aaaaaaaaaa"
		str2 += "bbbbbbbbbb"
	}

	scanner := NewScanner(strings.NewReader(str1 + "\n" + str2 + "\n"))

	// Does not start with an error.
	require.NoError(scanner.Err())

	// Reads the first line.
	require.True(scanner.Scan())
	require.Equal([]byte(str1), scanner.Bytes())
	require.Equal(str1, scanner.Text())
	require.NoError(scanner.Err())

	// Reads the second line.
	require.True(scanner.Scan())
	require.Equal([]byte(str2), scanner.Bytes())
	require.Equal(str2, scanner.Text())
	require.NoError(scanner.Err())

	// Finishes without error.
	require.False(scanner.Scan())
	require.NoError(scanner.Err())
}
