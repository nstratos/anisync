/*
Package hb provides a client for accessing the Hummingbird.me API.

Construct a new client, then use one of the client's services to access the
different Hummingbird API methods. For example, to get the currently watching
anime entries that are contained in the library of the user "cybrox":

	c := hb.NewClient(nil)

	entries, _, err := c.User.Library("cybrox", hb.StatusCurrentlyWatching)
	// handle err

	// do something with entries
*/
package hb
