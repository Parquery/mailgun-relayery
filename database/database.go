package database

import (
	"fmt"
	"os"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/golang/protobuf/proto"

	"encoding/binary"
	"github.com/Parquery/mailgun-relayery/dbc"
	"github.com/Parquery/mailgun-relayery/protoed"
	"math"
	"time"
)

const dbChannelName = "channel"
const dbTimestampName = "timestamp"

// Access enumerates different access rights for transactions on the database.
type Access int

const (
	// ControlAccess defines the access rights for the Control server.
	ControlAccess Access = 0
	// RelayAccess defines the access rights for the Relay server.
	RelayAccess Access = 1
)

// Descriptor is a key at which data is stored in the database.
type Descriptor string

// Encode encodes the descriptor to an array of bytes.
func (d Descriptor) Encode() []byte {
	return []byte(d)
}

// DecodeDescriptor decodes an array of bytes to a descriptor.
func DecodeDescriptor(data []byte) Descriptor {
	return Descriptor(data)
}

// Timestamp is a timestamp expressed as milliseconds from epoch
// in UTC.
type Timestamp uint64

// DecodeTimestamp decodes the array of bytes to a timestamp.
func DecodeTimestamp(data []byte) Timestamp {
	decoded := int64(binary.LittleEndian.Uint64(data))
	return Timestamp(decoded)
}

// Encode encodes the timestamp to an array of bytes.
func (t Timestamp) Encode() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(t))
	return bytes
}

// ToTime converts the timestamp to the corresponding time.Time object.
func (t Timestamp) ToTime() time.Time {
	seconds := t / 1000
	millis := t % 1000
	nanos := millis * 1000000
	return time.Unix(int64(seconds), int64(nanos)).In(time.UTC)
}

// TimestampFromTime converts a time.Time to the corresponding timestamp in nanoseconds.
func TimestampFromTime(tm time.Time) Timestamp {
	ts := tm.UnixNano() / int64(time.Millisecond)
	return Timestamp(ts)
}

// NewEnv creates a new database environment object.
// The database directory is assumed to be already initialized.
func NewEnv(access Access, path string) (e *Env, err error) {

	env, err := lmdb.NewEnv()
	if err != nil {
		err = fmt.Errorf("failed to initialize a database "+
			"environment: %s", err)
		return
	}

	err = env.SetMaxDBs(2)
	if err != nil {
		err = fmt.Errorf("failed to set the max. number of DBs to 1: "+
			"%s", err)
		closeErr := env.Close()
		if closeErr != nil {
			err = fmt.Errorf("%s; failed to close the database: "+
				"%s", err, closeErr)
		}
		return
	}

	mapSize := int64(32 * 1024 * 1024 * 1024)
	err = env.SetMapSize(mapSize)
	if err != nil {
		err = fmt.Errorf("failed to set the map size to "+
			"%d: %s", mapSize, err)
		closeErr := env.Close()
		if closeErr != nil {
			err = fmt.Errorf("%s; failed to close the "+
				"database: %s", err, closeErr)
		}
		return
	}

	err = env.Open(path, 0, 0660)

	if err != nil {
		err = fmt.Errorf("failed to open the database in the "+
			"directory %s: %s", path, err)
		closeErr := env.Close()
		if closeErr != nil {
			err = fmt.Errorf("%s; failed to close the database: "+
				"%s", err, closeErr)
		}
		return
	}

	e = &Env{
		Path:   path,
		env:    env,
		Access: access}

	return
}

// Env represents a database of channels.
type Env struct {
	Path   string
	env    *lmdb.Env
	Access Access
}

// Initialize initializes the database environment in the given directory.
// Initialize creates the expected database and should be called only once
// during the deployment.
func Initialize(access Access, path string) (err error) {
	_, err = os.Stat(path)
	if err != nil {
		err = fmt.Errorf("the directory of LMDB environment "+
			"is expected to exist: %s", err.Error())
		return
	}

	env, err := NewEnv(access, path)
	if err != nil {
		return
	}
	defer func() {
		err = env.env.Close()
	}()

	err = env.env.Update(func(txn *lmdb.Txn) (txnErr error) {
		_, txnErr = txn.OpenDBI(dbChannelName, lmdb.Create)
		if txnErr != nil {
			return
		}

		_, txnErr = txn.OpenDBI(dbTimestampName, lmdb.Create)
		if txnErr != nil {
			return
		}
		return
	})

	return
}

// GetChannel returns the channel associated with the descriptor in the
// database, if it exists; nil otherwise.
//
// GetChannel requires:
// * t.access == ControlAccess || t.access == RelayAccess
//
// GetChannel ensures:
// * err == nil || channel == nil || channel.Descriptor_ == descriptor
// * count positive if result: !dbc.InTest || !(err == nil && channel != nil) || t.mustCountCh() > 0
// * empty response for empty database: !dbc.InTest || !(err == nil && t.mustCountCh() == 0) || channel == nil
func (t *Txn) GetChannel(descriptor string) (channel *protoed.Channel,
	err error) {
	// Pre-condition
	if !(t.access == ControlAccess || t.access == RelayAccess) {
		panic("Violated: t.access == ControlAccess || t.access == RelayAccess")
	}

	// Post-conditions
	defer func() {
		switch {
		case !(err == nil || channel == nil || channel.Descriptor_ == descriptor):
			panic("Violated: err == nil || channel == nil || channel.Descriptor_ == descriptor")
		case !(!dbc.InTest || !(err == nil && channel != nil) || t.mustCountCh() > 0):
			panic("Violated: count positive if result: !dbc.InTest || !(err == nil && channel != nil) || t.mustCountCh() > 0")
		case !(!dbc.InTest || !(err == nil && t.mustCountCh() == 0) || channel == nil):
			panic("Violated: empty response for empty database: !dbc.InTest || !(err == nil && t.mustCountCh() == 0) || channel == nil")
		default:
			// Pass
		}
	}()

	encodedKey := Descriptor(descriptor).Encode()

	value, getErr := t.lmdbTxn.Get(t.channelDbi, encodedKey)
	switch {
	case getErr == nil:
		// pass
	case lmdb.IsNotFound(getErr):
		// not found, return
		return
	default:
		err = fmt.Errorf("failed to get the channel: %s", getErr.Error())
		return
	}

	channel = &protoed.Channel{}
	err = proto.Unmarshal(value, channel)

	if err != nil {
		err = fmt.Errorf("failed to unmarshal the channel: %s",
			err.Error())
		return
	}

	return
}

// GetTimestamp returns the Timestamp associated with the descriptor in the
// database, if it exists; nil otherwise.
//
// GetTimestamp requires:
// * t.access == ControlAccess || t.access == RelayAccess
//
// GetTimestamp ensures:
// * err == nil || Timestamp == nil || *Timestamp > 0
// * !dbc.InTest || !(err == nil && Timestamp != nil) || t.mustCountTs() > 0
// * empty response for empty database: !dbc.InTest || !(err == nil && t.mustCountTs() == 0) || Timestamp == nil
func (t *Txn) GetTimestamp(descriptor string) (Timestamp *Timestamp, err error) {
	// Pre-condition
	if !(t.access == ControlAccess || t.access == RelayAccess) {
		panic("Violated: t.access == ControlAccess || t.access == RelayAccess")
	}

	// Post-conditions
	defer func() {
		switch {
		case !(err == nil || Timestamp == nil || *Timestamp > 0):
			panic("Violated: err == nil || Timestamp == nil || *Timestamp > 0")
		case !(!dbc.InTest || !(err == nil && Timestamp != nil) || t.mustCountTs() > 0):
			panic("Violated: !dbc.InTest || !(err == nil && Timestamp != nil) || t.mustCountTs() > 0")
		case !(!dbc.InTest || !(err == nil && t.mustCountTs() == 0) || Timestamp == nil):
			panic("Violated: empty response for empty database: !dbc.InTest || !(err == nil && t.mustCountTs() == 0) || Timestamp == nil")
		default:
			// Pass
		}
	}()

	encodedKey := Descriptor(descriptor).Encode()

	value, getErr := t.lmdbTxn.Get(t.timestampDbi, encodedKey)
	switch {
	case getErr == nil:
		// pass
	case lmdb.IsNotFound(getErr):
		// not found, return
		return
	default:
		err = fmt.Errorf("failed to get the Timestamp: %s", getErr.Error())
		return
	}

	ts := DecodeTimestamp(value)
	Timestamp = &ts

	if err != nil {
		err = fmt.Errorf("failed to unmarshal the Timestamp: %s",
			err.Error())
		return
	}

	return
}

// CountChannels returns the number of channels in the database.
//
// CountChannels requires:
// * t.access == ControlAccess || t.access == RelayAccess
//
// CountChannels ensures:
// * !dbc.InTest || err != nil || t.mustCountTs() <= count
func (t *Txn) CountChannels() (count uint64, err error) {
	// Pre-condition
	if !(t.access == ControlAccess || t.access == RelayAccess) {
		panic("Violated: t.access == ControlAccess || t.access == RelayAccess")
	}

	// Post-condition
	defer func() {
		if !(!dbc.InTest || err != nil || t.mustCountTs() <= count) {
			panic("Violated: !dbc.InTest || err != nil || t.mustCountTs() <= count")
		}
	}()

	stat, err := t.lmdbTxn.Stat(t.channelDbi)
	if err != nil {
		err = fmt.Errorf("error while accessing the lmdb stats: %s",
			err.Error())
		return
	}

	count = stat.Entries
	return
}

// CountTimestamps returns the number of timestamps in the database.
//
// CountTimestamps requires:
// * t.access == ControlAccess || t.access == RelayAccess
func (t *Txn) CountTimestamps() (count uint64, err error) {
	// Pre-condition
	if !(t.access == ControlAccess || t.access == RelayAccess) {
		panic("Violated: t.access == ControlAccess || t.access == RelayAccess")
	}

	stat, err := t.lmdbTxn.Stat(t.timestampDbi)
	if err != nil {
		err = fmt.Errorf("error while accessing the lmdb stats: %s",
			err.Error())
		return
	}

	count = stat.Entries
	return
}

// ChannelPage returns the requested `page` of the database entries.
//
// Given the entire database as a list of channels, the elements starting from
// index `page*perPage` are inserted in the list sequentially until either the
// list contains `perPage` elements, or there are no more database entries.
//
// If the database has less than `page*perPage` elements,
// an empty list is returned.
// ChannelPage requires:
// * t.access == ControlAccess
//
// ChannelPage ensures:
// * err != nil || len(channels) <= int(perPage)
// * err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64((page-1)*perPage) || len(channels) == 0)
// * err != nil || !dbc.InTest || !(t.mustCountCh() > uint64((page-1)*perPage) || len(channels) > 0)
// * err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64(page*perPage) || len(channels) == int(perPage))
func (t *Txn) ChannelPage(page uint,
	perPage uint) (channels []*protoed.Channel, err error) {
	// Pre-condition
	if !(t.access == ControlAccess) {
		panic("Violated: t.access == ControlAccess")
	}

	// Post-conditions
	defer func() {
		switch {
		case !(err != nil || len(channels) <= int(perPage)):
			panic("Violated: err != nil || len(channels) <= int(perPage)")
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64((page-1)*perPage) || len(channels) == 0)):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64((page-1)*perPage) || len(channels) == 0)")
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() > uint64((page-1)*perPage) || len(channels) > 0)):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() > uint64((page-1)*perPage) || len(channels) > 0)")
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64(page*perPage) || len(channels) == int(perPage))):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64(page*perPage) || len(channels) == int(perPage))")
		default:
			// Pass
		}
	}()

	pagesRange, err := t.pageRange(page, perPage)
	if err != nil {
		return
	}

	if pagesRange == nil {
		// out-of-bounds page index
		return
	}

	cur, err := t.lmdbTxn.OpenCursor(t.channelDbi)
	if err != nil {
		err = fmt.Errorf("error while accessing the cursor: %s",
			err.Error())
		return
	}
	defer cur.Close()

	for index := uint(0); index < pagesRange.end; index++ {
		_, val, curErr := cur.Get(nil, nil, lmdb.Next)
		if lmdb.IsNotFound(curErr) {
			return
		}

		if curErr != nil {
			err = fmt.Errorf("error while browsing the database with "+
				"the cursor: %s", curErr.Error())
			return
		}

		if index < pagesRange.start {
			continue
		}

		channel := &protoed.Channel{}
		err = proto.Unmarshal(val, channel)

		if err != nil {
			err = fmt.Errorf("failed to unmarshal the channel: %s",
				err.Error())
			return
		}
		channels = append(channels, channel)
	}

	return

}

// pageRange contains the start (inclusive) and end (exclusive) index of a
// pagination.
type pageRange struct {
	start uint
	end   uint
}

// pageRange gives the indices of a pagination, or nil if the index is
// out-of-bounds.
//
// pageRange ensures:
// * err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64(page*perPage)) || pRange == nil
// * err != nil || !dbc.InTest || !(t.mustCountCh() > uint64(page*perPage)) || pRange != nil
// * err != nil || pRange == nil || pRange.start < pRange.end
// * err != nil || pRange == nil || pRange.end-pRange.start <= perPage
// * err != nil || pRange == nil || (page-1)*perPage == pRange.start
// * err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64((page+1)*perPage)) || pRange.start-pRange.end == perPage
func (t *Txn) pageRange(page uint,
	perPage uint) (pRange *pageRange, err error) {
	// Post-conditions
	defer func() {
		switch {
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64(page*perPage)) || pRange == nil):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() <= uint64(page*perPage)) || pRange == nil")
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() > uint64(page*perPage)) || pRange != nil):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() > uint64(page*perPage)) || pRange != nil")
		case !(err != nil || pRange == nil || pRange.start < pRange.end):
			panic("Violated: err != nil || pRange == nil || pRange.start < pRange.end")
		case !(err != nil || pRange == nil || pRange.end-pRange.start <= perPage):
			panic("Violated: err != nil || pRange == nil || pRange.end-pRange.start <= perPage")
		case !(err != nil || pRange == nil || (page-1)*perPage == pRange.start):
			panic("Violated: err != nil || pRange == nil || (page-1)*perPage == pRange.start")
		case !(err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64((page+1)*perPage)) || pRange.start-pRange.end == perPage):
			panic("Violated: err != nil || !dbc.InTest || !(t.mustCountCh() >= uint64((page+1)*perPage)) || pRange.start-pRange.end == perPage")
		default:
			// Pass
		}
	}()

	dbSize, err := t.CountChannels()
	if err != nil {
		return
	}

	if dbSize <= uint64(perPage*(page-1)) {
		// out-of-bounds page index
		return
	}

	from := (page - 1) * perPage
	to := math.Min(float64(page*perPage), float64(dbSize))
	pRange = &pageRange{uint(from), uint(to)}
	return

}

// PutChannel inserts a channel in the database, keyed on its descriptor.
//
// PutChannel requires:
// * t.access == ControlAccess
// * channel != nil
//
// PutChannel preamble:
//  var oldHas bool
//  oldCount := uint64(0)
//  if dbc.InTest {
//  	oldCount = t.mustCountCh()
//  	oldHas = t.mustGetCh(channel.Descriptor_) != nil
//  }
//
// PutChannel ensures:
// * !dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) != nil
// * !dbc.InTest || err != nil || oldHas || t.mustCountCh() == oldCount+1
// * !dbc.InTest || err != nil || !oldHas || t.mustCountCh() == oldCount
// * !dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) == nil || t.mustGetCh(channel.Descriptor_).MinPeriod == channel.MinPeriod || t.mustGetTs(channel.Descriptor_) == nil
func (t *Txn) PutChannel(channel *protoed.Channel) (err error) {
	// Pre-conditions
	switch {
	case !(t.access == ControlAccess):
		panic("Violated: t.access == ControlAccess")
	case !(channel != nil):
		panic("Violated: channel != nil")
	default:
		// Pass
	}

	// Preamble starts.
	var oldHas bool
	oldCount := uint64(0)
	if dbc.InTest {
		oldCount = t.mustCountCh()
		oldHas = t.mustGetCh(channel.Descriptor_) != nil
	}
	// Preamble ends.

	// Post-conditions
	defer func() {
		switch {
		case !(!dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) != nil):
			panic("Violated: !dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) != nil")
		case !(!dbc.InTest || err != nil || oldHas || t.mustCountCh() == oldCount+1):
			panic("Violated: !dbc.InTest || err != nil || oldHas || t.mustCountCh() == oldCount+1")
		case !(!dbc.InTest || err != nil || !oldHas || t.mustCountCh() == oldCount):
			panic("Violated: !dbc.InTest || err != nil || !oldHas || t.mustCountCh() == oldCount")
		case !(!dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) == nil || t.mustGetCh(channel.Descriptor_).MinPeriod == channel.MinPeriod || t.mustGetTs(channel.Descriptor_) == nil):
			panic("Violated: !dbc.InTest || err != nil || t.mustGetCh(channel.Descriptor_) == nil || t.mustGetCh(channel.Descriptor_).MinPeriod == channel.MinPeriod || t.mustGetTs(channel.Descriptor_) == nil")
		default:
			// Pass
		}
	}()

	var serialized []byte
	serialized, err = proto.Marshal(channel)
	if err != nil {
		err = fmt.Errorf("failed to serialize the channel: "+
			"%s", err.Error())
		return
	}

	encodedKey := Descriptor(channel.Descriptor_).Encode()

	prevChan, err := t.GetChannel(channel.Descriptor_)
	if err != nil {
		err = fmt.Errorf("failed to fetch the channel: %s", err.Error())
		return
	}

	err = t.lmdbTxn.Put(t.channelDbi, encodedKey, serialized, 0)
	if err != nil {
		err = fmt.Errorf("failed to put the channel: %s", err.Error())
		return
	}

	if prevChan != nil && prevChan.MinPeriod != channel.MinPeriod {
		err = t.removeTimestamp(channel.Descriptor_)
		if err != nil {
			err = fmt.Errorf("failed to erase the timestamp: %s", err.Error())
			return
		}
	}

	return
}

// PutTimestamp inserts a Timestamp in the database, keyed on its descriptor.
//
// PutTimestamp requires:
// * !dbc.InTest || t.access == RelayAccess
// * Timestamp != nil
// * *Timestamp > 0
//
// PutTimestamp preamble:
//  var oldHas bool
//  oldCount := uint64(0)
//  if dbc.InTest {
//  	oldCount = t.mustCountTs()
//  	oldHas = t.mustGetTs(string(descriptor)) != nil
//  }
//
// PutTimestamp ensures:
// * !dbc.InTest || err != nil || t.mustGetTs(string(descriptor)) != nil
// * !dbc.InTest || err != nil || oldHas || t.mustCountTs() == oldCount+1
// * !dbc.InTest || err != nil || !oldHas || t.mustCountTs() == oldCount
func (t *Txn) PutTimestamp(descriptor Descriptor,
	Timestamp *Timestamp) (err error) {
	// Pre-conditions
	switch {
	case !(!dbc.InTest || t.access == RelayAccess):
		panic("Violated: !dbc.InTest || t.access == RelayAccess")
	case !(Timestamp != nil):
		panic("Violated: Timestamp != nil")
	case !(*Timestamp > 0):
		panic("Violated: *Timestamp > 0")
	default:
		// Pass
	}

	// Preamble starts.
	var oldHas bool
	oldCount := uint64(0)
	if dbc.InTest {
		oldCount = t.mustCountTs()
		oldHas = t.mustGetTs(string(descriptor)) != nil
	}
	// Preamble ends.

	// Post-conditions
	defer func() {
		switch {
		case !(!dbc.InTest || err != nil || t.mustGetTs(string(descriptor)) != nil):
			panic("Violated: !dbc.InTest || err != nil || t.mustGetTs(string(descriptor)) != nil")
		case !(!dbc.InTest || err != nil || oldHas || t.mustCountTs() == oldCount+1):
			panic("Violated: !dbc.InTest || err != nil || oldHas || t.mustCountTs() == oldCount+1")
		case !(!dbc.InTest || err != nil || !oldHas || t.mustCountTs() == oldCount):
			panic("Violated: !dbc.InTest || err != nil || !oldHas || t.mustCountTs() == oldCount")
		default:
			// Pass
		}
	}()

	encodedTs := Timestamp.Encode()
	encodedKey := descriptor.Encode()

	err = t.lmdbTxn.Put(t.timestampDbi, encodedKey, encodedTs, 0)
	if err != nil {
		err = fmt.Errorf("failed to put the Timestamp: %s", err.Error())
		return
	}

	return
}

// RemoveChannel removes a channel from the database.
//
// RemoveChannel requires:
// * t.access == ControlAccess
//
// RemoveChannel preamble:
//  var oldHasCh, oldHasTs bool
//  oldCountCh := uint64(0)
//  oldCountTs := uint64(0)
//  if dbc.InTest {
//  	oldCountTs = t.mustCountTs()
//  	oldCountCh = t.mustCountCh()
//  	oldHasCh = t.mustGetCh(descriptor) != nil
//  	oldHasTs = t.mustGetTs(descriptor) != nil
//  }
//
// RemoveChannel ensures:
// * !dbc.InTest || err != nil || t.mustGetCh(descriptor) == nil && t.mustGetTs(descriptor) == nil
// * !dbc.InTest || err != nil || !oldHasCh || t.mustCountCh() == oldCountCh-1
// * !dbc.InTest || err != nil || oldHasCh || t.mustCountCh() == oldCountCh
// * !dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCountTs-1
// * !dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCountTs
func (t *Txn) RemoveChannel(descriptor string) (err error) {
	// Pre-condition
	if !(t.access == ControlAccess) {
		panic("Violated: t.access == ControlAccess")
	}

	// Preamble starts.
	var oldHasCh, oldHasTs bool
	oldCountCh := uint64(0)
	oldCountTs := uint64(0)
	if dbc.InTest {
		oldCountTs = t.mustCountTs()
		oldCountCh = t.mustCountCh()
		oldHasCh = t.mustGetCh(descriptor) != nil
		oldHasTs = t.mustGetTs(descriptor) != nil
	}
	// Preamble ends.

	// Post-conditions
	defer func() {
		switch {
		case !(!dbc.InTest || err != nil || t.mustGetCh(descriptor) == nil && t.mustGetTs(descriptor) == nil):
			panic("Violated: !dbc.InTest || err != nil || t.mustGetCh(descriptor) == nil && t.mustGetTs(descriptor) == nil")
		case !(!dbc.InTest || err != nil || !oldHasCh || t.mustCountCh() == oldCountCh-1):
			panic("Violated: !dbc.InTest || err != nil || !oldHasCh || t.mustCountCh() == oldCountCh-1")
		case !(!dbc.InTest || err != nil || oldHasCh || t.mustCountCh() == oldCountCh):
			panic("Violated: !dbc.InTest || err != nil || oldHasCh || t.mustCountCh() == oldCountCh")
		case !(!dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCountTs-1):
			panic("Violated: !dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCountTs-1")
		case !(!dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCountTs):
			panic("Violated: !dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCountTs")
		default:
			// Pass
		}
	}()

	encodedKey := Descriptor(descriptor).Encode()

	delErr := t.lmdbTxn.Del(t.channelDbi, encodedKey, []byte{0})
	switch {
	case delErr == nil:
		// pass
	case lmdb.IsNotFound(delErr):
		// not found, pass
	default:
		err = fmt.Errorf("failed to erase the channel: %s", delErr.Error())
		return
	}

	err = t.removeTimestamp(descriptor)
	if err != nil {
		err = fmt.Errorf("failed to erase the timestamp: "+
			"%s", err.Error())
		return
	}

	return
}

// removeTimestamp removes a Timestamp from the database.
//
// removeTimestamp requires:
// * t.access == ControlAccess
//
// removeTimestamp preamble:
//  var oldHasTs bool
//  oldCount := uint64(0)
//  if dbc.InTest {
//  	oldCount = t.mustCountTs()
//  	oldHasTs = t.mustGetTs(descriptor) != nil
//  }
//
// removeTimestamp ensures:
// * !dbc.InTest || err != nil || t.mustGetTs(descriptor) == nil
// * !dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCount
// * !dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCount-1
func (t *Txn) removeTimestamp(descriptor string) (err error) {
	// Pre-condition
	if !(t.access == ControlAccess) {
		panic("Violated: t.access == ControlAccess")
	}

	// Preamble starts.
	var oldHasTs bool
	oldCount := uint64(0)
	if dbc.InTest {
		oldCount = t.mustCountTs()
		oldHasTs = t.mustGetTs(descriptor) != nil
	}
	// Preamble ends.

	// Post-conditions
	defer func() {
		switch {
		case !(!dbc.InTest || err != nil || t.mustGetTs(descriptor) == nil):
			panic("Violated: !dbc.InTest || err != nil || t.mustGetTs(descriptor) == nil")
		case !(!dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCount):
			panic("Violated: !dbc.InTest || err != nil || oldHasTs || t.mustCountTs() == oldCount")
		case !(!dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCount-1):
			panic("Violated: !dbc.InTest || err != nil || !oldHasTs || t.mustCountTs() == oldCount-1")
		default:
			// Pass
		}
	}()

	encodedKey := Descriptor(descriptor).Encode()

	delErr := t.lmdbTxn.Del(t.timestampDbi, encodedKey, []byte{0})
	switch {
	case delErr == nil:
		// pass
	case lmdb.IsNotFound(delErr):
		// not found, return
		return
	default:
		err = fmt.Errorf("failed to erase the Timestamp: %s", delErr.Error())
		return
	}

	return
}

// mustCountCh returns the Count of entries in the channel database.
// In case of error,it panics.
func (t *Txn) mustCountCh() uint64 {

	count, getErr := t.CountChannels()
	if getErr != nil {
		panic(fmt.Sprintf("failed to get the channels count: %s", getErr.Error()))
	}

	return count
}

// mustCountTs returns the Count of entries in the timestamp database.
// In case of error, it panics.
func (t *Txn) mustCountTs() uint64 {

	count, getErr := t.CountTimestamps()
	if getErr != nil {
		panic(fmt.Sprintf("failed to get the timestamps count: %s", getErr.Error()))
	}

	return count
}

// mustGetCh returns the element associated with the descriptor in the
// channel database.
// If such element doesn't exist, it returns nil. In case of error, it panics.
func (t *Txn) mustGetCh(descriptor string) *protoed.Channel {

	elem, getErr := t.GetChannel(descriptor)
	if getErr != nil {
		panic(fmt.Sprintf("failed to get the channel: %s", getErr.Error()))
	}

	return elem
}

// mustGetTs returns the element associated with the descriptor in the
// timestamp database.
// If such element doesn't exist, it returns nil. In case of error, it panics.
func (t *Txn) mustGetTs(descriptor string) *Timestamp {

	elem, getErr := t.GetTimestamp(descriptor)
	if getErr != nil {
		panic(fmt.Sprintf("failed to get the timestamp: %s", getErr.Error()))
	}

	return elem
}

// Update executes a read-write transaction.
func (e *Env) Update(fn func(txn *Txn) error) error {
	return e.env.Update(func(lmdbTxn *lmdb.Txn) error {
		channelDbi, err := lmdbTxn.OpenDBI(dbChannelName, 0)
		if err != nil {
			return err
		}

		timestampDbi, err := lmdbTxn.OpenDBI(dbTimestampName, 0)
		if err != nil {
			return err
		}

		txn := &Txn{lmdbTxn: lmdbTxn,
			channelDbi: channelDbi, timestampDbi: timestampDbi,
			env: e, access: e.Access}
		return fn(txn)
	})
}

// View executes a read-only transaction.
func (e *Env) View(fn func(txn *Txn) error) error {
	return e.env.View(func(lmdbTxn *lmdb.Txn) error {
		channelDbi, err := lmdbTxn.OpenDBI(dbChannelName, 0)
		if err != nil {
			return err
		}

		timestampDbi, err := lmdbTxn.OpenDBI(dbTimestampName, 0)
		if err != nil {
			return err
		}

		txn := &Txn{lmdbTxn: lmdbTxn,
			channelDbi: channelDbi, timestampDbi: timestampDbi,
			env: e, access: e.Access}
		return fn(txn)
	})
}

// Close closes the database.
func (e *Env) Close() error {
	return e.env.Close()
}

// Txn represents a transaction over the entries of the database.
type Txn struct {
	lmdbTxn      *lmdb.Txn
	channelDbi   lmdb.DBI
	timestampDbi lmdb.DBI
	env          *Env
	access       Access
}
