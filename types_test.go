package redgiant

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntBoolUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input       int
		expected    bool
		shouldError bool
	}{
		{input: 0, expected: false, shouldError: false},
		{input: 1, expected: true, shouldError: false},
		{input: -1, shouldError: true},
		{input: 42, shouldError: true},
	}

	for _, test := range tests {
		t.Run(
			strconv.Itoa(test.input), func(t *testing.T) {
				i := struct{ Actual int }{test.input}
				j, err := json.Marshal(i)
				require.NoError(t, err)

				o := struct{ Actual intBool }{}
				err = json.Unmarshal(j, &o)

				if test.shouldError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, test.expected, bool(o.Actual))
				}
			},
		)
	}
}
