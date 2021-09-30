package kitsu

import (
	"fmt"
)

// The possible library entry statuses. They are convenient when creating a
// LibraryEntry or for making comparisons with LibraryEntry.Status.
const (
	LibraryEntryStatusCurrent   = "current"
	LibraryEntryStatusPlanned   = "planned"
	LibraryEntryStatusCompleted = "completed"
	LibraryEntryStatusOnHold    = "on_hold"
	LibraryEntryStatusDropped   = "dropped"
)

// LibraryService handles communication with the library entry related methods
// of the Kitsu API.
type LibraryService service

// LibraryEntry represents a Kitsu user's library entry.
type LibraryEntry struct {
	ID             string `jsonapi:"primary,libraryEntries"`
	Status         string `jsonapi:"attr,status,omitempty"`         // Status for related media. Can be compared with LibraryEntryStatus constants.
	Progress       int    `jsonapi:"attr,progress,omitempty"`       // How many episodes/chapters have been consumed, e.g. 22.
	Reconsuming    bool   `jsonapi:"attr,reconsuming,omitempty"`    // Whether the media is being reconsumed, e.g. false.
	ReconsumeCount int    `jsonapi:"attr,reconsumeCount,omitempty"` // How many times the media has been reconsumed, e.g. 0.
	Notes          string `jsonapi:"attr,notes,omitempty"`          // Note attached to this entry, e.g. Very Interesting!
	Private        bool   `jsonapi:"attr,private,omitempty"`        // Whether this entry is hidden from the public, e.g. false.
	Rating         string `jsonapi:"attr,rating,omitempty"`         // User rating out of 5.0.
	UpdatedAt      string `jsonapi:"attr,updatedAt,omitempty"`      // When the entry was last updated, e.g. 2016-11-12T03:35:00.064Z.

	// Relationships.

	User  *User       `jsonapi:"relation,user,omitempty"`
	Anime *Anime      `jsonapi:"relation,anime,omitempty"`
	Media interface{} `jsonapi:"relation,media,omitempty"`
}

// Show returns details for a specific LibraryEntry by providing a unique identifier
// of the library entry, e.g. 5269457.
func (s *LibraryService) Show(libraryEntryID string, opts ...URLOption) (*LibraryEntry, *Response, error) {
	u := fmt.Sprintf(defaultAPIVersion+"library-entries/%s", libraryEntryID)

	req, err := s.client.NewRequest("GET", u, nil, opts...)
	if err != nil {
		return nil, nil, err
	}

	e := new(LibraryEntry)
	resp, err := s.client.Do(req, e)
	if err != nil {
		return nil, resp, err
	}

	return e, resp, nil
}

// List returns a list of Library entries. Optional parameters can be specified
// to filter the search results and control pagination, sorting etc.
func (s *LibraryService) List(opts ...URLOption) ([]*LibraryEntry, *Response, error) {
	u := defaultAPIVersion + "library-entries"

	req, err := s.client.NewRequest("GET", u, nil, opts...)
	if err != nil {
		return nil, nil, err
	}

	var entries []*LibraryEntry
	resp, err := s.client.Do(req, &entries)
	if err != nil {
		return nil, resp, err
	}

	return entries, resp, nil
}

// Create creates a library entry. This method needs authentication.
func (s *LibraryService) Create(e *LibraryEntry, opts ...URLOption) (*LibraryEntry, *Response, error) {
	u := defaultAPIVersion + "library-entries"

	req, err := s.client.NewRequest("POST", u, e, opts...)
	if err != nil {
		return nil, nil, err
	}

	var entry = new(LibraryEntry)
	resp, err := s.client.Do(req, entry)
	if err != nil {
		return nil, resp, err
	}

	return entry, resp, nil
}

// Delete deletes a library entry. This method needs authentication.
func (s *LibraryService) Delete(id string, opts ...URLOption) (*Response, error) {
	u := defaultAPIVersion + "library-entries/" + id

	req, err := s.client.NewRequest("DELETE", u, nil, opts...)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
