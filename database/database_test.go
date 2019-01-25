package database

import (
	"bytes"
	"os"
	"testing"

	"fmt"
	"github.com/Parquery/mailgun-relayery/protoed"
	"io/ioutil"
	"strconv"
	"time"
)

func TestEncodeDecode(t *testing.T) {
	descriptors := []string{"client-1/pipeline-2", "client-5/pipeline-1"}

	data := [][]byte{
		{0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2d, 0x31, 0x2f, 0x70,
			0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2d, 0x32},
		{0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2d, 0x35, 0x2f, 0x70,
			0x69, 0x70, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2d, 0x31},
	}

	for i, expected := range descriptors {
		key := Descriptor(expected)
		encoded := key.Encode()

		if !bytes.Equal(data[i], encoded) {
			t.Fatalf("expected %#v, got %#v", data[i], encoded)
		}

		k := DecodeDescriptor(encoded)

		if k != key {
			t.Fatalf("expected %v, got %v", key, k)
		}
	}

	timestamps := []uint64{
		12312312312123123,
		383838383123,
		234092304923234234,
		23982300,
		0,
		1,
	}

	data = [][]byte{
		{0xf3, 0x4e, 0xba, 0x99, 0xfb, 0xbd, 0x2b, 0x0},
		{0x13, 0xec, 0x8c, 0x5e, 0x59, 0x0, 0x0, 0x0},
		{0xba, 0xf, 0x9a, 0x7, 0xb6, 0xa9, 0x3f, 0x3},
		{0xdc, 0xf0, 0x6d, 0x1, 0x0, 0x0, 0x0, 0x0},
		{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	}

	for i, expected := range timestamps {
		key := Timestamp(expected)
		encoded := key.Encode()

		if !bytes.Equal(data[i], encoded) {
			t.Fatalf("expected %#v, got %#v", data[i], encoded)
		}

		k := DecodeTimestamp(encoded)

		if k != key {
			t.Fatalf("expected %v, got %v", key, k)
		}
	}
}

func TestFromToTime(t *testing.T) {
	timestamps := []uint64{
		1545396245000,
		1542394683321,
		723945600000,
	}

	data := []time.Time{
		time.Date(2018, 12, 21, 12, 44, 5, 0, time.UTC),
		time.Date(2018, 11, 16, 18, 58, 3, 321000000, time.UTC),
		time.Date(1992, 12, 10, 0, 0, 0, 0, time.UTC),
	}

	for i, expected := range timestamps {
		tm := Timestamp(expected).ToTime()

		if tm != data[i] {
			t.Fatalf("expected %s, got %s", data[i].String(), tm.String())
		}

		ts := TimestampFromTime(tm)

		if ts != Timestamp(expected) {
			t.Fatalf("expected %v, got %v", ts, expected)
		}
	}
}

func TestTxn_CountChannels(t *testing.T) {
	d, err := emptyDatabase(ControlAccess)
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

	descriptor := "some-channel"
	sender := protoed.Entity{Name: "Ludwig van Beethoven",
		Email: "ludwig.van.beethoven@composers.com"}
	recipients := []*protoed.Entity{
		{Name: "Johannes Brahms", Email: "johannes.brahms@composers.com"}}
	cc := []*protoed.Entity{
		{Name: "Richard Wagner", Email: "richard.wagner@composers.com"}}
	bcc := []*protoed.Entity{
		{Name: "Robert Schumann", Email: "robert.schumann@composers.com"}}

	channel := protoed.Channel{Descriptor_: descriptor,
		Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
		Sender: &sender, Recipients: recipients,
		Cc: cc, Bcc: bcc, MinPeriod: 0.0001, MaxSize: 10000000}

	// check that count is zero
	err = d.View(func(txn *Txn) (txnerr error) {
		var count uint64
		count, txnerr = txn.CountChannels()
		if count != 0 {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got count = %d", descriptor, count)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put a couple of channels
	channelCount := 20
	err = d.Update(func(txn *Txn) (txnerr error) {
		for i := 0; i < channelCount; i++ {
			channel.Descriptor_ = descriptor + strconv.Itoa(i)
			txnerr = txn.PutChannel(&channel)
			if txnerr != nil {
				t.Fatalf(txnerr.Error())
			}
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check correctness of count
	err = d.View(func(txn *Txn) (txnerr error) {
		var count uint64
		count, txnerr = txn.CountChannels()
		if count != uint64(channelCount) {
			t.Fatalf("Expected count to be %d, but got %d",
				channelCount, count)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTxn_CountTimestamps(t *testing.T) {
	d, err := emptyDatabase(RelayAccess)
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

	descriptor := "some-channel"
	timestamp := Timestamp(1222221123111)

	// check that count is zero
	err = d.View(func(txn *Txn) (txnerr error) {
		var count uint64
		count, txnerr = txn.CountTimestamps()
		if count != 0 {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got count = %d", descriptor, count)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put a couple of timestamps
	timestampCount := 20
	err = d.Update(func(txn *Txn) (txnerr error) {
		for i := 0; i < timestampCount; i++ {
			desc := descriptor + strconv.Itoa(i)
			txnerr = txn.PutTimestamp(Descriptor(desc), &timestamp)
			if txnerr != nil {
				t.Fatalf(txnerr.Error())
			}
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check correctness of count
	err = d.View(func(txn *Txn) (txnerr error) {
		var count uint64
		count, txnerr = txn.CountTimestamps()
		if count != uint64(timestampCount) {
			t.Fatalf("Expected count to be %d, but got %d",
				timestampCount, count)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTxn_ChannelPage(t *testing.T) {
	d, err := emptyDatabase(ControlAccess)
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

	descriptor := "some-channel"
	sender := protoed.Entity{Name: "Ludwig van Beethoven",
		Email: "ludwig.van.beethoven@composers.com"}
	recipients := []*protoed.Entity{
		{Name: "Johannes Brahms", Email: "johannes.brahms@composers.com"}}
	cc := []*protoed.Entity{
		{Name: "Richard Wagner", Email: "richard.wagner@composers.com"}}
	bcc := []*protoed.Entity{
		{Name: "Robert Schumann", Email: "robert.schumann@composers.com"}}

	channel := protoed.Channel{Descriptor_: descriptor,
		Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
		Sender: &sender, Recipients: recipients,
		Cc: cc, Bcc: bcc, MinPeriod: 0.0001, MaxSize: 10000000}

	// empty database calls
	err = d.View(func(txn *Txn) (txnerr error) {
		channels, txnerr := txn.ChannelPage(uint(1), uint(10))
		if txnerr != nil {
			t.Fatalf(txnerr.Error())
		}
		if len(channels) != 0 {
			t.Fatalf("Expected no channel pages but got count = %d",
				len(channels))
		}

		channels, txnerr = txn.ChannelPage(uint(2), uint(3))
		if txnerr != nil {
			t.Fatalf(txnerr.Error())
		}
		if len(channels) != 0 {
			t.Fatalf("Expected no channel pages but got count = %d",
				len(channels))
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put a couple of channels
	channelCount := 20
	err = d.Update(func(txn *Txn) (txnerr error) {
		for i := 0; i < channelCount; i++ {
			desc := descriptor
			if i < 10 {
				desc += "0"
			}
			desc += strconv.Itoa(i)
			channel.Descriptor_ = desc
			txnerr = txn.PutChannel(&channel)
			if txnerr != nil {
				t.Fatalf(txnerr.Error())
			}
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check correctness of pagination
	type testcase struct {
		page        uint
		perPage     uint
		descriptors []string
	}
	testcases := []testcase{
		{page: 1, perPage: 2,
			descriptors: []string{descriptor + "00", descriptor + "01"}},
		{page: 5, perPage: 1,
			descriptors: []string{descriptor + "04"}},
		{page: 1, perPage: 30,
			descriptors: []string{descriptor + "00",
				descriptor + "01", descriptor + "02", descriptor + "03",
				descriptor + "04", descriptor + "05", descriptor + "06",
				descriptor + "07", descriptor + "08", descriptor + "09",
				descriptor + "10", descriptor + "11", descriptor + "12",
				descriptor + "13", descriptor + "14", descriptor + "15",
				descriptor + "16", descriptor + "17", descriptor + "18",
				descriptor + "19"}},
		{page: 10, perPage: 4,
			descriptors: []string{}},
		{page: 1, perPage: 9,
			descriptors: []string{descriptor + "00",
				descriptor + "01", descriptor + "02", descriptor + "03",
				descriptor + "04", descriptor + "05", descriptor + "06",
				descriptor + "07", descriptor + "08"}},
		{page: 2, perPage: 9,
			descriptors: []string{descriptor + "09",
				descriptor + "10", descriptor + "11", descriptor + "12",
				descriptor + "13", descriptor + "14", descriptor + "15",
				descriptor + "16", descriptor + "17"}},
		{page: 3, perPage: 9,
			descriptors: []string{descriptor + "18", descriptor + "19"}},
	}

	for _, testcase := range testcases {
		err = d.View(func(txn *Txn) (txnerr error) {
			channels, txnerr := txn.ChannelPage(uint(testcase.page),
				uint(testcase.perPage))
			if txnerr != nil {
				t.Fatalf(txnerr.Error())
			}
			if len(channels) != len(testcase.descriptors) {
				t.Errorf("Expected %d channels, got %d",
					len(testcase.descriptors), len(channels))
			} else {
				for i, channel := range channels {
					if channel.Descriptor_ != testcase.descriptors[i] {
						t.Errorf("Expected descriptor %s, got %s",
							channel.Descriptor_, testcase.descriptors[i])
					}
				}
			}
			return
		})
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestTxn_PutChannel(t *testing.T) {
	d, err := emptyDatabase(ControlAccess)
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

	descriptor := "some-channel"
	sender := protoed.Entity{Name: "Ludwig van Beethoven",
		Email: "ludwig.van.beethoven@composers.com"}
	recipients := []*protoed.Entity{
		{Name: "Johannes Brahms", Email: "johannes.brahms@composers.com"}}
	cc := []*protoed.Entity{
		{Name: "Richard Wagner", Email: "richard.wagner@composers.com"}}
	bcc := []*protoed.Entity{
		{Name: "Robert Schumann", Email: "robert.schumann@composers.com"}}

	channel := protoed.Channel{Descriptor_: descriptor,
		Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
		Sender: &sender, Recipients: recipients,
		Cc: cc, Bcc: bcc, MinPeriod: 0.0001, MaxSize: 10000000}

	// check absence of the channel
	err = d.View(func(txn *Txn) (txnerr error) {
		var channelPtr *protoed.Channel
		channelPtr, txnerr = txn.GetChannel(descriptor)
		if channelPtr != nil {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got: %#v", descriptor, *channelPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put the channel
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutChannel(&channel)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check presence of the channel
	err = d.View(func(txn *Txn) (txnerr error) {
		var channelPtr *protoed.Channel
		channelPtr, txnerr = txn.GetChannel(descriptor)
		if channelPtr == nil {
			t.Fatalf("Expected channel to be present " +
				"in the database, but got nil")
		} else {
			CompareChannels(&channel, channelPtr, t)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put a timestamp for the channel with Relay access
	relayEnv, err := NewEnv(RelayAccess, d.Path)
	if err != nil {
		t.Fatalf("Failed to open the database with Relay access: %s",
			err.Error())
	}
	timestamp := Timestamp(1231231231231)
	err = relayEnv.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutTimestamp(Descriptor(descriptor), &timestamp)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// overwrite the channel with one having the same min_period
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutChannel(&channel)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check that the timestamp was kept
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr == nil {
			t.Fatalf("Expected timestamp to be present " +
				"in the database, but got nil")
		} else if *tsPtr != timestamp {
			t.Fatalf("Expected timestamp to be %d, got %d.",
				timestamp, *tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// overwrite the channel with a different min_period
	err = d.Update(func(txn *Txn) (txnerr error) {
		channel.MinPeriod += 2
		txnerr = txn.PutChannel(&channel)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check that the timestamp was deleted
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr != nil {
			t.Fatalf("Expected timestamp to be absent from the "+
				"database, but got %v", tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTxn_PutTimestamp(t *testing.T) {
	d, err := emptyDatabase(RelayAccess)
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

	descriptor := "some-channel"
	timestamp := Timestamp(12391924913323423)

	// check absence of the timestamp
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr != nil {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got: %#v", descriptor, *tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// put the timestamp
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutTimestamp(Descriptor(descriptor), &timestamp)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check presence of the timestamp
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr == nil {
			t.Fatalf("Expected timestamp to be present " +
				"in the database, but got nil")
		} else if *tsPtr != timestamp {
			t.Fatalf("Expected timestamp to be %d, got %d.",
				timestamp, *tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTxn_RemoveChannel(t *testing.T) {
	d, err := emptyDatabase(ControlAccess) //TODO test with put timestamp as well
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

	descriptor := "some-other-channel"
	sender := protoed.Entity{Name: "Johann Sebastian Bach",
		Email: "johann.bach@composers.com"}
	recipients := []*protoed.Entity{
		{Name: "CPE Bach", Email: "cpe.bach@composers.com"}}
	cc := []*protoed.Entity{
		{Name: "Fryderyk Chopin", Email: "fryderyk.chopin@composers.com"}}
	bcc := []*protoed.Entity{
		{Name: "Franz Schubert", Email: "franz.schubert@composers.com"}}

	channel := protoed.Channel{Descriptor_: descriptor,
		Token:  "oqiwdJKNsdKIUwezd92DNQsndkDERDFKJNQWSwq3rODIU",
		Sender: &sender, Recipients: recipients, Cc: cc, Bcc: bcc,
		MinPeriod: 0.0001, MaxSize: 10000000}

	// put the channel
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutChannel(&channel)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// verify presence of the channel
	err = d.View(func(txn *Txn) (txnerr error) {
		var channelPtr *protoed.Channel
		channelPtr, txnerr = txn.GetChannel(descriptor)
		if channelPtr == nil {
			t.Fatalf("Expected channel to be present in the database, " +
				"but got nil")
		} else {
			CompareChannels(&channel, channelPtr, t)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// erase the channel
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.RemoveChannel(descriptor)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check absence of the channel
	err = d.View(func(txn *Txn) (txnerr error) {
		var channelPtr *protoed.Channel
		channelPtr, txnerr = txn.GetChannel(descriptor)
		if channelPtr != nil {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got: %#v", descriptor, *channelPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestTxn_RemoveTimestamp(t *testing.T) {
	d, err := emptyDatabase(ControlAccess)
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

	descriptor := "some-other-channel"
	timestamp := Timestamp(1242342323423423)

	// put the timestamp
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.PutTimestamp(Descriptor(descriptor), &timestamp)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// verify presence of the timestamp
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr == nil {
			t.Fatalf("Expected timestamp to be present in the " +
				"database, but got nil")
		} else if *tsPtr != timestamp {
			t.Fatalf("Expected timestamp to be %d, got %d.",
				timestamp, *tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// erase the timestamp
	err = d.Update(func(txn *Txn) (txnerr error) {
		txnerr = txn.removeTimestamp(descriptor)
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// check absence of the timestamp
	err = d.View(func(txn *Txn) (txnerr error) {
		var tsPtr *Timestamp
		tsPtr, txnerr = txn.GetTimestamp(descriptor)
		if tsPtr != nil {
			t.Fatalf("Expected no entry in the database keyed on "+
				"descriptor %s, but got: %#v", descriptor, *tsPtr)
		}
		return
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}

// emptyDatabase creates an empty database.
func emptyDatabase(access Access) (e *Env, err error) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return
	}

	err = Initialize(access, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	e, err = NewEnv(access, tmpdir)
	if err != nil {
		removeErr := os.RemoveAll(tmpdir)
		if removeErr != nil {
			err = fmt.Errorf("%s; %s", err.Error(), removeErr.Error())
		}
		return
	}

	return
}
