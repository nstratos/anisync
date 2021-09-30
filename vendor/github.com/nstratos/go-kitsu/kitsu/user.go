package kitsu

import (
	"fmt"
)

// Possible values for User.RatingSystem.
const (
	UserRatingSystemAdvanced = "advanced"
	UserRatingSystemRegular  = "regular"
	UserRatingSystemSimple   = "simple"
)

// Possible values for User.Theme.
const (
	UserThemeLight = "light"
	UserThemeDark  = "dark"
)

// UserService handles communication with the user related methods of the
// Kitsu API.
//
// Kitsu API docs:
// http://docs.kitsu.apiary.io/#reference/users/users
type UserService service

// User represents a Kitsu user.
type User struct {
	ID string `jsonapi:"primary,users"`

	// --- Attributes ---

	// ISO 8601 date and time, e.g. 2017-07-27T22:21:26.824Z
	CreatedAt string `jsonapi:"attr,createdAt,omitempty"`

	// ISO 8601 of last modification, e.g. 2017-07-27T22:47:45.129Z
	UpdatedAt string `jsonapi:"attr,updatedAt,omitempty"`

	// e.g. vikhyat
	Name string `jsonapi:"attr,name,omitempty"`

	PastNames []string `jsonapi:"attr,pastNames,omitempty"`

	// Unique slug used for page URLs, e.g. vikhyat
	Slug string `jsonapi:"attr,slug,omitempty"`

	// Max length of 500 characters, e.g.
	//
	// Co-founder of Hummingbird. Obsessed with Gumi.
	About string `jsonapi:"attr,about,omitempty"`

	// e.g. Seattle, WA
	Location string `jsonapi:"attr,location,omitempty"`

	// e.g. Waifu
	WaifuOrHusbando string `jsonapi:"attr,waifuOrHusbando,omitempty"`

	// e.g. 1716
	FollowersCount int `jsonapi:"attr,followersCount,omitempty"`

	// e.g. 2031
	FollowingCount int `jsonapi:"attr,followingCount,omitempty"`

	Birthday string `jsonapi:"attr,birthday,omitempty"`
	Gender   string `jsonapi:"attr,gender,omitempty"`

	CommentsCount       int `jsonapi:"attr,commentsCount,omitempty"`
	FavoritesCount      int `jsonapi:"attr,favoritesCount,omitempty"`
	LikesGivenCount     int `jsonapi:"attr,likesGivenCount,omitempty"`
	ReviewsCount        int `jsonapi:"attr,reviewsCount,omitempty"`
	LikesReceivedCount  int `jsonapi:"attr,likesReceivedCount,omitempty"`
	PostsCount          int `jsonapi:"attr,postsCount,omitempty"`
	RatingsCount        int `jsonapi:"attr,ratingsCount,omitempty"`
	MediaReactionsCount int `jsonapi:"attr,mediaReactionsCount,omitempty"`

	// e.g. 2015-01-30T16:49:35.173Z
	ProExpiresAt string `jsonapi:"attr,proExpiresAt,omitempty"`

	Title string `jsonapi:"attr,title,omitempty"`

	ProfileCompleted bool `jsonapi:"attr,profileCompleted,omitempty"`
	FeedCompleted    bool `jsonapi:"attr,feedCompleted,omitempty"`

	// e.g.
	//
	// "tiny": "https://media.kitsu.io/users/avatars/1/tiny.jpg?1434087646"
	//
	// "small": "https://media.kitsu.io/users/avatars/1/small.jpg?1434087646"
	//
	// "medium": "https://media.kitsu.io/users/avatars/1/medium.jpg?1434087646"
	//
	// "large": "https://media.kitsu.io/users/avatars/1/large.jpg?1434087646"
	//
	// "original": "https://media.kitsu.io/users/avatars/1/original.jpg?1434087646"
	//
	// It may also contain a meta object with additional dimensions objects for
	// each previous Avatar type.
	Avatar     map[string]interface{} `jsonapi:"attr,avatar,omitempty"`
	CoverImage map[string]interface{} `jsonapi:"attr,coverImage,omitempty"`

	// Possible valued described by UserRatingSystem constants.
	RatingSystem string `jsonapi:"attr,ratingSystem,omitempty"`

	// Possible valued described by UserTheme constants.
	Theme string `jsonapi:"attr,theme,omitempty"`

	FacebookID string `jsonapi:"attr,facebookId,omitempty"`

	// --- Relationships ---

	Waifu          *Character      `jsonapi:"relation,waifu,omitempty"`
	LibraryEntries []*LibraryEntry `jsonapi:"relation,libraryEntries,omitempty"`
}

// Show returns details for a specific User by providing the ID of the user
// e.g. 29745.
func (s *UserService) Show(userID string, opts ...URLOption) (*User, *Response, error) {
	u := fmt.Sprintf(defaultAPIVersion+"users/%s", userID)

	req, err := s.client.NewRequest("GET", u, nil, opts...)
	if err != nil {
		return nil, nil, err
	}

	user := new(User)
	resp, err := s.client.Do(req, user)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

// List returns a list of Users. Optional parameters can be specified to filter
// the search results and control pagination, sorting etc.
func (s *UserService) List(opts ...URLOption) ([]*User, *Response, error) {
	u := defaultAPIVersion + "users"

	req, err := s.client.NewRequest("GET", u, nil, opts...)
	if err != nil {
		return nil, nil, err
	}

	var users []*User
	resp, err := s.client.Do(req, &users)
	if err != nil {
		return nil, resp, err
	}

	return users, resp, nil
}
