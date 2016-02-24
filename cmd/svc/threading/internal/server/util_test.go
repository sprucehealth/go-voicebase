package server

import (
	"testing"

	"github.com/sprucehealth/backend/test"
)

func TestTruncateUTF8(t *testing.T) {
	cases := []struct {
		i string
		o string
		n int
	}{
		{"", "", 10},
		{"test", "", -1},
		{"test", "", 0},
		{"test", "t", 1},
		{"test", "te", 2},
		{"test", "tes", 3},
		{"test", "test", 4},
		{"test", "test", 5},
		{"😶", "😶", 1},
		{"😶", "😶", 2},
		{`\😶/`, `\`, 1},
		{`\😶/`, `\😶`, 2},
		{`\😶/`, `\😶/`, 3},
		{`\😶/`, `\😶/`, 4},
	}
	for _, c := range cases {
		test.Equals(t, c.o, truncateUTF8(c.i, c.n))
	}
}
