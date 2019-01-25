package control

import (
	"fmt"
	"math"

	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/dbc"
	"github.com/Parquery/mailgun-relayery/protoed"
)

// paginateChannels computes the response for a channels page listing
// request. If the page is out of bounds, the response contains no channels.
//
// paginateChannels requires:
// * db != nil
// * page != 0
// * perPage != 0
//
// paginateChannels ensures:
// * err != nil || channelsPage.Page == int32(page)
// * err != nil || channelsPage.PerPage == int32(perPage)
// * err != nil || channelsPage.PageCount >= 0
// * err != nil || len(channelsPage.Channels) <= int(perPage)
// * err != nil || !dbc.InTest || mustCount(db) != 0 || (page == 1 && len(channelsPage.Channels) == 0 && channelsPage.PageCount == 0)
func paginateChannels(page uint, perPage uint,
	db *database.Env) (channelsPage ChannelsPage, err error) {
	// Pre-conditions
	switch {
	case !(db != nil):
		panic("Violated: db != nil")
	case !(page != 0):
		panic("Violated: page != 0")
	case !(perPage != 0):
		panic("Violated: perPage != 0")
	default:
		// Pass
	}

	// Post-conditions
	defer func() {
		switch {
		case !(err != nil || channelsPage.Page == int32(page)):
			panic("Violated: err != nil || channelsPage.Page == int32(page)")
		case !(err != nil || channelsPage.PerPage == int32(perPage)):
			panic("Violated: err != nil || channelsPage.PerPage == int32(perPage)")
		case !(err != nil || channelsPage.PageCount >= 0):
			panic("Violated: err != nil || channelsPage.PageCount >= 0")
		case !(err != nil || len(channelsPage.Channels) <= int(perPage)):
			panic("Violated: err != nil || len(channelsPage.Channels) <= int(perPage)")
		case !(err != nil || !dbc.InTest || mustCount(db) != 0 || (page == 1 && len(channelsPage.Channels) == 0 && channelsPage.PageCount == 0)):
			panic("Violated: err != nil || !dbc.InTest || mustCount(db) != 0 || (page == 1 && len(channelsPage.Channels) == 0 && channelsPage.PageCount == 0)")
		default:
			// Pass
		}
	}()

	channels := []Channel{}
	pageCount := int32(0)
	channelCount := uint64(0)
	dbErr := db.View(func(txn *database.Txn) (txnErr error) {
		channelCount, txnErr = txn.CountChannels()
		if txnErr != nil {
			return
		}
		pageCount = int32(math.Ceil(float64(channelCount) / float64(perPage)))

		if channelCount > 0 {
			var channelsProto []*protoed.Channel
			channelsProto, txnErr = txn.ChannelPage(page, perPage)
			if txnErr != nil {
				return
			}
			for _, protoChan := range channelsProto {
				jsonChannel := ProtoToJSON(protoChan)
				channels = append(channels, *jsonChannel)
			}
		}

		return
	})
	if dbErr != nil {
		err = fmt.Errorf("error while retrieving the channels in the "+
			"database: %s", dbErr.Error())
		return
	}

	channelsPage = ChannelsPage{PageCount: pageCount, PerPage: int32(perPage),
		Page: int32(page), Channels: channels}
	return
}

// mustCount returns the Count of entries in the database. In case of error,
// it panics.
func mustCount(db *database.Env) uint64 {
	var count uint64
	err := db.View(func(txn *database.Txn) (txnerr error) {
		count, txnerr = txn.CountChannels()
		return
	})
	if err != nil {
		panic(err.Error())
	}

	return count
}
