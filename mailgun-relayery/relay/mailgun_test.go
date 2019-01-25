package relay

import (
	"testing"

	"github.com/Parquery/mailgun-relayery/mailgun-relay-controlery/control"
)

func TestEntityToMailGunEmail(t *testing.T) {
	cpeBach := "CPE Bach"
	fryderykChopin := "Fryderyk Chopin"
	entities := []control.Entity{
		{Email: "johann.bach@composers.com"},
		{Name: &cpeBach, Email: "cpe.bach@composers.com"},
		{Name: &fryderykChopin, Email: "fryderyk.chopin@composers.com"},
	}
	expected := []string{
		"johann.bach@composers.com",
		"CPE Bach <cpe.bach@composers.com>",
		"Fryderyk Chopin <fryderyk.chopin@composers.com>",
	}

	for i, entity := range entities {
		got := entityToMailgunEmail(entity)
		if expected[i] != got {
			t.Errorf("for entity %#v, expected %s, got %s",
				entity, expected[i], got)
		}
	}
}
