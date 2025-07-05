package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlowdownFilter(t *testing.T) {
	t.Parallel()

	testData := []struct {
		name     string
		config   SlowdownInterceptorConfig
		input    string
		expected bool
	}{
		{
			name: "no config no match",
			config: SlowdownInterceptorConfig{},
			input:    "a",
			expected: false,
		},
		{
			name: "specific inclusions match",
			config: SlowdownInterceptorConfig{
				IncludedMethods: "a,b,c",
			},
			input:    "a",
			expected: true,
		},
		{
			name: "specific inclusions no match",
			config: SlowdownInterceptorConfig{
				IncludedMethods: "a,b,c",
			},
			input:    "d",
			expected: false,
		},
		{
			name: "all methods match",
			config: SlowdownInterceptorConfig{
				IncludedMethods: "*",
			},
			input:    "d",
			expected: true,
		},
		{
			name: "exclusions match",
			config: SlowdownInterceptorConfig{
				ExcludedMethods: "d,e,f",
			},
			input:    "d",
			expected: false,
		},
		{
			name: "exclusions no match",
			config: SlowdownInterceptorConfig{
				ExcludedMethods: "d,e,f",
			},
			input:    "a",
			expected: true,
		},
	}
	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			inter, err := NewSlowdownInterceptor(tc.config)
			require.NoError(t, err)
			require.Equal(t, tc.expected, inter.filter(tc.input))
		})
	}
}
