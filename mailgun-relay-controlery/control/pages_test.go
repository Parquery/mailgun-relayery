package control

import (
	"strconv"
	"testing"

	"fmt"
	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/protoed"
	"io/ioutil"
	"os"
)

func TestPaginateChannels_EmptyDatabase(t *testing.T) {
	d, err := emptyDatabase()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		err = os.RemoveAll(d.Path)
		if err != nil {
			t.Fatal(err.Error())
		}
	}()
	defer func() {
		err = d.Close()
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	channelsPage, err := paginateChannels(1, 100, d)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if channelsPage.Page != 1 {
		t.Errorf("Expected page equals %d, got %d",
			1, channelsPage.Page)
	}
	if channelsPage.PageCount != 0 {
		t.Errorf("Expected page count equals %d, got %d",
			0, channelsPage.PageCount)
	}
	if channelsPage.PerPage != 100 {
		t.Errorf("Expected perPage equals %d, got %d",
			100, channelsPage.PageCount)
	}
	if len(channelsPage.Channels) > 0 {
		t.Errorf("Expected empty channels array, got %#v",
			channelsPage.Channels)
	}
}
func TestPaginateChannels(t *testing.T) {
	d, populatedChannels, err := populatedDatabase()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		err = os.RemoveAll(d.Path)
		if err != nil {
			t.Fatal(err.Error())
		}
	}()
	defer func() {
		err = d.Close()
		if err != nil {
			t.Fatal(err.Error())
		}
	}()

	// typical usecase: first page, 100 items per page
	channelsPage, err := paginateChannels(1, 100, d)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedMap := populatedChannels
	if channelsPage.Page != 1 {
		t.Errorf("Expected page equals %d, got %d",
			1, channelsPage.Page)
	}
	if channelsPage.PageCount != 1 {
		t.Errorf("Expected page count equals %d, got %d",
			1, channelsPage.PageCount)
	}
	if channelsPage.PerPage != 100 {
		t.Errorf("Expected perPage equals %d, got %d",
			100, channelsPage.PageCount)
	}
	channelsPageMap := make(map[string]*protoed.Channel)
	for _, chann := range channelsPage.Channels {
		protoChan := JSONToProto(&chann)
		channelsPageMap[protoChan.Descriptor_] = protoChan
	}
	if len(channelsPageMap) != len(expectedMap) {
		t.Errorf("Expected %d items, got %d",
			len(expectedMap), len(channelsPageMap))
	} else {
		for desc, expected := range expectedMap {
			database.CompareChannels(expected, channelsPageMap[desc], t)
		}
	}

	// custom usecase: page 10, 1 item per page
	channelsPage, err = paginateChannels(10, 1, d)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedMap = make(map[string]*protoed.Channel)
	expectedMap["channel-10"] = populatedChannels["channel-10"]
	if channelsPage.Page != 10 {
		t.Errorf("Expected page equals %d, got %d",
			10, channelsPage.Page)
	}
	if channelsPage.PageCount != 20 {
		t.Errorf("Expected page count equals %d, got %d",
			20, channelsPage.PageCount)
	}
	if channelsPage.PerPage != 1 {
		t.Errorf("Expected perPage equals %d, got %d",
			1, channelsPage.PageCount)
	}
	channelsPageMap = make(map[string]*protoed.Channel)

	for _, chann := range channelsPage.Channels {
		protoChan := JSONToProto(&chann)
		channelsPageMap[protoChan.Descriptor_] = protoChan
	}
	if len(channelsPageMap) != len(expectedMap) {
		t.Errorf("Expected %d items, got %d",
			len(expectedMap), len(channelsPageMap))
	} else {
		for desc, expected := range expectedMap {
			database.CompareChannels(expected, channelsPageMap[desc], t)
		}
	}

	// another custom usecase: page 4, 5 items per page
	channelsPage, err = paginateChannels(4, 5, d)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expectedMap = make(map[string]*protoed.Channel)
	expectedMap["channel-16"] = populatedChannels["channel-16"]
	expectedMap["channel-17"] = populatedChannels["channel-17"]
	expectedMap["channel-18"] = populatedChannels["channel-18"]
	expectedMap["channel-19"] = populatedChannels["channel-19"]
	expectedMap["channel-20"] = populatedChannels["channel-20"]
	if channelsPage.Page != 4 {
		t.Errorf("Expected page equals %d, got %d",
			4, channelsPage.Page)
	}
	if channelsPage.PageCount != 4 {
		t.Errorf("Expected page count equals %d, got %d",
			4, channelsPage.PageCount)
	}
	if channelsPage.PerPage != 5 {
		t.Errorf("Expected perPage equals %d, got %d",
			5, channelsPage.PageCount)
	}
	channelsPageMap = make(map[string]*protoed.Channel)

	for _, chann := range channelsPage.Channels {
		protoChan := JSONToProto(&chann)
		channelsPageMap[protoChan.Descriptor_] = protoChan
	}
	if len(channelsPageMap) != len(expectedMap) {
		t.Errorf("Expected %d items, got %d",
			len(expectedMap), len(channelsPageMap))
	} else {
		for desc, expected := range expectedMap {
			database.CompareChannels(expected, channelsPageMap[desc], t)
		}
	}
}

// emptyDatabase creates an empty database.
func emptyDatabase() (e *database.Env, err error) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return
	}

	err = database.Initialize(database.ControlAccess, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	e, err = database.NewEnv(database.ControlAccess, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	return
}

// populatedDatabase creates an empty database and populates it of dumnmy data.
func populatedDatabase() (e *database.Env,
	channelsMap map[string]*protoed.Channel, err error) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return
	}

	err = database.Initialize(database.ControlAccess, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	e, err = database.NewEnv(database.ControlAccess, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	channelsMap = make(map[string]*protoed.Channel)
	for i := 1; i <= 20; i++ {
		descriptor := "channel-"
		if i < 10 {
			descriptor += "0"
		}
		descriptor += strconv.Itoa(i)
		sender := protoed.Entity{Name: "Johann Sebastian Bach",
			Email: "johann.bach@composers.com"}
		recipients := []*protoed.Entity{{Name: "CPE Bach",
			Email: "cpe.bach@composers.com"}}
		cc := []*protoed.Entity{{Name: "Fryderyk Chopin",
			Email: "fryderyk.chopin@composers.com"}}
		bcc := []*protoed.Entity{{Name: "Franz Schubert",
			Email: "franz.schubert@composers.com"}}
		channel := &protoed.Channel{Descriptor_: descriptor,
			Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
			Sender: &sender, Recipients: recipients, Cc: cc,
			Bcc: bcc, MinPeriod: 0.0001, MaxSize: 10000000}
		channelsMap[descriptor] = channel

		err = e.Update(func(txn *database.Txn) (txnerr error) {
			txnerr = txn.PutChannel(channel)
			if txnerr != nil {
				return
			}
			return
		})
		if err != nil {
			return
		}
	}

	return
}
