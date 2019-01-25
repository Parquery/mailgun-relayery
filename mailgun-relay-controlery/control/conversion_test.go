package control

import (
	"encoding/json"
	"testing"

	"github.com/golang/protobuf/jsonpb"

	"github.com/Parquery/mailgun-relayery/protoed"
	"strings"
)

func TestJSONToProto(t *testing.T) {
	name1, name2, name3, name4 :=
		"Ludwig van Beethoven",
		"Johannes Brahms",
		"Richard Wagner",
		"Robert Schumann"
	sender := Entity{Name: &name1, Email: "ludwig.van.beethoven@composers.com"}
	recipients := []Entity{
		{Name: &name2, Email: "johannes.brahms@composers.com"}}
	cc := []Entity{
		{Name: &name3, Email: "richard.wagner@composers.com"}}
	bcc := []Entity{
		{Name: &name4, Email: "robert.schumann@composers.com"}}
	domain := "test.maildomain.com"

	jsonChan := Channel{Descriptor: Descriptor("some-channel"),
		Token:  Token("oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU"),
		Sender: sender, Recipients: recipients, Cc: cc, Bcc: bcc,
		Domain: domain, MinPeriod: 0.0001, MaxSize: 10000000}
	converted := JSONToProto(&jsonChan)

	marshaler := jsonpb.Marshaler{OrigName: false}
	protoChanStr, err := marshaler.MarshalToString(converted)
	if err != nil {
		t.Fatalf("failed to marshal the protocol buffer: %s", err.Error())
	}

	expected := dedent(`{"descriptor":"some-channel",
				"token":"oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
				"sender":{"email":"ludwig.van.beethoven@composers.com",
				"name":"Ludwig van Beethoven"},
				"recipients":[{"email":"johannes.brahms@composers.com",
				"name":"Johannes Brahms"}],"cc":[{"email":
				"richard.wagner@composers.com","name":"Richard Wagner"}],
				"bcc":[{"email":"robert.schumann@composers.com","name":
				"Robert Schumann"}],"domain":"test.maildomain.com",
				"minPeriod":0.0001,"maxSize":10000000}`)

	if protoChanStr != expected {
		t.Fatalf("expected %s\n, got %s", expected, protoChanStr)
	}

}

func TestProtoToJSON(t *testing.T) {
	sender := protoed.Entity{Name: "Ludwig van Beethoven",
		Email: "ludwig.van.beethoven@composers.com"}
	recipients := []*protoed.Entity{{Name: "Johannes Brahms",
		Email: "johannes.brahms@composers.com"}}
	cc := []*protoed.Entity{{Name: "Richard Wagner",
		Email: "richard.wagner@composers.com"}}
	bcc := []*protoed.Entity{{Name: "Robert Schumann",
		Email: "robert.schumann@composers.com"}}
	domain := "test.maildomain.com"

	protoChan := &protoed.Channel{Descriptor_: "some-channel",
		Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
		Sender: &sender, Recipients: recipients, Cc: cc, Bcc: bcc,
		Domain: domain, MinPeriod: 0.0001, MaxSize: 10000000}
	converted := ProtoToJSON(protoChan)

	bts, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("failed to marhsal the channel: %s", err.Error())
	}
	jsonChanStr := string(bts)
	expected := dedent(`{"descriptor":"some-channel",
					"token":"oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
					"sender":{"email":"ludwig.van.beethoven@composers.com",
					"name":"Ludwig van Beethoven"},"recipients":[{"email":
					"johannes.brahms@composers.com","name":"Johannes Brahms"}],
					"cc":[{"email":"richard.wagner@composers.com","name":
					"Richard Wagner"}],"bcc":[{"email":"robert.schumann@
					composers.com","name":"Robert Schumann"}],"domain":"test.
					maildomain.com","min_period":0.0001,"max_size":10000000}`)

	if expected != jsonChanStr {
		t.Fatalf("expected %s\n, got %s", expected, jsonChanStr)
	}

}

func dedent(text string) string {
	noTabs := strings.Replace(text, "\t", "", -1)
	return strings.Replace(noTabs, "\n", "", -1)
}
