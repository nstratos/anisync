package anisync

import "time"

// Diff represents the difference of two anime lists (left and right). It
// contains the orignal lists, the missing anime, the anime that need to
// be updated and the ones that are up to date. It is assuming that right
// list is larger than left list. Typically the left list will be the
// MyAnimeList and the right list will be the Hummingbird list.
type Diff struct {
	Left       []Anime
	Right      []Anime
	Missing    []Anime
	NeedUpdate []AniDiff
	UpToDate   []Anime
	Uncertain  []AniDiff
}

// Compare compares two anime lists and returns the difference containing
// the orignal lists, the missing anime, the anime that need to be updated
// and the ones that are up to date. It is assuming that right list is
// larger than left list. Typically the left list will be the MyAnimeList
// and the right list will be the Hummingbird list.
func Compare(left, right []Anime) *Diff {
	diff := &Diff{Left: left, Right: right}
	var (
		missing    []Anime
		needUpdate []AniDiff
		upToDate   []Anime
		uncertain  []AniDiff
	)
	for _, a := range right {
		found := FindByID(left, a.ID)
		if found != nil {
			needsUpdate, isUncertain, diff := compare(*found, a)
			switch {
			case needsUpdate:
				needUpdate = append(needUpdate, diff)
			case isUncertain:
				uncertain = append(uncertain, diff)
			default:
				upToDate = append(upToDate, a)
			}
		} else {
			missing = append(missing, a)
		}
	}
	diff.Missing = missing
	diff.NeedUpdate = needUpdate
	diff.UpToDate = upToDate
	diff.Uncertain = uncertain
	return diff
}

type AniDiff struct {
	Anime           Anime
	Status          *StatusDiff
	EpisodesWatched *EpisodesWatchedDiff
	Rating          *RatingDiff
	Rewatching      *RewatchingDiff
	LastUpdated     *LastUpdatedDiff
}

type StatusDiff struct {
	Got  string
	Want string
}

type EpisodesWatchedDiff struct {
	Got  int
	Want int
}

type RatingDiff struct {
	Got  string
	Want string
}

type RewatchingDiff struct {
	Got  bool
	Want bool
}

type LastUpdatedDiff struct {
	Got  time.Time
	Want time.Time
}

func compare(left, right Anime) (bool, bool, AniDiff) {
	needsUpdate, uncertain := false, false

	diff := AniDiff{Anime: right}
	if got, want := left.Status, right.Status; got != want {
		diff.Status = &StatusDiff{got, want}
		// fmt.Printf("->Status got %v, want %v\n", got, want)
		needsUpdate = true
	}
	if got, want := left.EpisodesWatched, right.EpisodesWatched; got != want {
		//fmt.Printf("->EpisodesWatched got %v, want %v\n", got, want)
		diff.EpisodesWatched = &EpisodesWatchedDiff{got, want}
		needsUpdate = true
	}
	if got, want := left.Rating, right.Rating; got != want {
		//fmt.Printf("->Rating got %v, want %v\n", got, want)
		// MyAnimeList API always sends score 0 even if the user hasn't entered
		// a score. So if we get "0.0" but Hummingbird has "" then we skip.
		// The second (reverse) case practically can never happen especially
		// since compare is meant to be used with left as MAL and right as HB
		// but just in case it used in reverse in the future.
		if !(got == "0.0" && want == "") && !(got == "" && want == "0.0") {
			diff.Rating = &RatingDiff{got, want}
			needsUpdate = true
		}
	}
	if got, want := left.Rewatching, right.Rewatching; got != want {
		//fmt.Printf("->Rewatching got %v, want %v\n", got, want)
		diff.Rewatching = &RewatchingDiff{got, want}
		needsUpdate = true
	}
	if left.LastUpdated != nil && right.LastUpdated != nil {
		// MAL API does not return comments so we cannot compare with notes.
		// It does not return times rewatched either. The only thing we can do
		// is compare the last updates. The problem is that MAL does not
		// always change last update when a change happens.
		c := compareLastUpdate(left, right)
		if got, want := left.LastUpdated, right.LastUpdated; c == -1 {
			diff.LastUpdated = &LastUpdatedDiff{*got, *want}
			if needsUpdate == false {
				uncertain = true
			}
		}
	}

	return needsUpdate, uncertain, diff
}

// compareLastUpdate compares the LastUpdated time values of two Anime.
//
// If left anime was updated before right, it returns -1.
// If left and right anime have equal update times, it returns 0.
// If left anime was updated after right, it returns 1.
//
// In the typical case, left will be a MyAnimeList anime and right will be a
// HummingBird anime. The application does not handle the case where the
// MyAnimeList anime has been updated after the HummingBird anime as that
// would mean being able to sync from MyAnimeList to HummingBird.
func compareLastUpdate(left, right Anime) int {
	switch {
	// Left anime was updated before right. Left needs update.
	case left.LastUpdated.Before(*right.LastUpdated):
		return -1
	// Left and right anime have equal update times. Up to date.
	case left.LastUpdated.Equal(*right.LastUpdated):
		return 0
	// Left anime was updated after right. Right needs update.
	default:
		return 1
	}
}
