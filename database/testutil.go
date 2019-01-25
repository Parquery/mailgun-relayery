package database

import (
	"testing"

	"github.com/Parquery/mailgun-relayery/protoed"
)

// CompareChannels runs a deep comparison between two proto Channel messages,
// raising errors upon differences.
func CompareChannels(
	expected *protoed.Channel,
	got *protoed.Channel,
	t *testing.T) {

	if expected == nil || got == nil {
		t.Fatalf("unexpected nil channel. expected=%#v, got=%#v",
			expected, got)
	}

	if expected.String() != got.String() {
		t.Errorf("expected %s, got %s",
			expected.String(), got.String())
	}

}
