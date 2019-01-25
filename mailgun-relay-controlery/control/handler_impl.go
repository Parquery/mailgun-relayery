package control

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Parquery/mailgun-relayery/database"
	"github.com/Parquery/mailgun-relayery/protoed"
)

// HandlerImpl implements the Handler.
type HandlerImpl struct {
	LogErr *log.Logger
	LogOut *log.Logger
	Env    *database.Env
}

// PutChannel implements Handler.PutChannel.
func (h *HandlerImpl) PutChannel(w http.ResponseWriter,
	r *http.Request,
	channel Channel) {

	protoChan := JSONToProto(&channel)
	dbErr := h.Env.Update(func(txn *database.Txn) (txnErr error) {
		txnErr = txn.PutChannel(protoChan)
		return
	})
	if dbErr != nil {
		http.Error(w,
			"Failed to store the channel.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to store the channel in the "+
			"database: %s\n", r.URL.String(), dbErr.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(fmt.Sprintf("The channel with descriptor"+
		" %s was correctly stored.", protoChan.Descriptor_)))
	if err != nil {
		h.LogErr.Printf("%s: Error while writing to the response "+
			"writer: %s\n", r.URL.String(), err.Error())
	}
	h.LogOut.Printf("%s: The channel with descriptor %s "+
		"was correctly stored in the database.\n",
		r.URL.String(), protoChan.Descriptor_)
}

// DeleteChannel implements Handler.DeleteChannel.
func (h *HandlerImpl) DeleteChannel(w http.ResponseWriter,
	r *http.Request,
	descriptor Descriptor) {

	descriptorStr := string(descriptor)
	var protoChan *protoed.Channel
	err := h.Env.View(func(txn *database.Txn) (txnErr error) {
		protoChan, txnErr = txn.GetChannel(descriptorStr)
		return
	})
	if err != nil {
		http.Error(w, "Failed to check for the channel's presence.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to check the channel's presence "+
			"in the database: %s\n", r.URL.String(), err.Error())
		return
	}
	if protoChan == nil {
		w.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf("No channel associated to "+
			"the descriptor %s was found.", descriptorStr)

		_, err = w.Write([]byte(msg))
		if err != nil {
			h.LogErr.Printf("%s: Error while writing to the response "+
				"writer: %s\n", r.URL.String(), err.Error())
		}
		h.LogOut.Printf("%s: %s.\n", r.URL.String(), msg)
		return
	}

	err = h.Env.Update(func(txn *database.Txn) (txnErr error) {
		txnErr = txn.RemoveChannel(descriptorStr)
		return
	})
	if err != nil {
		http.Error(w, "Failed to erase the channel.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to erase the channel from the "+
			"database: %s\n", r.URL.String(), err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("The channel with "+
		"descriptor %s was correctly erased.", descriptorStr)))
	if err != nil {
		h.LogErr.Printf("%s: Error while writing to the response "+
			"writer: %s\n", r.URL.String(), err.Error())
	}
	h.LogOut.Printf("%s: The channel with descriptor %s was correctly "+
		"erased from the database.\n",
		r.URL.String(), descriptorStr)

}

// ListChannels implements Handler.ListChannels.
func (h *HandlerImpl) ListChannels(w http.ResponseWriter,
	r *http.Request,
	page *int32,
	perPage *int32) {

	pageNr := uint(1)
	if page != nil {
		if *page <= 0 {
			http.Error(w, "Page index smaller than 1 is not allowed.",
				http.StatusBadRequest)
			h.LogErr.Printf("%s: Received a page number smaller than "+
				"1 (%d)\n", r.URL.String(), *page)
			return
		}
		pageNr = uint(*page)
	}

	perPageNr := uint(100)
	if perPage != nil {
		if *perPage <= 0 {
			http.Error(w,
				"perPage variable smaller than 1 is not allowed.",
				http.StatusBadRequest)
			h.LogErr.Printf("%s: Received a perPage variable smaller "+
				"than 1 (%d)\n", r.URL.String(), *perPage)
			return
		}
		perPageNr = uint(*perPage)
	}

	channelsPage, err := paginateChannels(pageNr, perPageNr, h.Env)
	if err != nil {
		http.Error(w, "Failed to fetch the channel listing response.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to fetch the channel listing "+
			"from the database: %s\n", r.URL.String(), err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&channelsPage)

	if err != nil {
		http.Error(w, "Failed to marshal the channel listing response.",
			http.StatusInternalServerError)
		h.LogErr.Printf("%s: Failed to marshal the channel listing "+
			"response: %s\n", r.URL.String(), err.Error())
	}

}
