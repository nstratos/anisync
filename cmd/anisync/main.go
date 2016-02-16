package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/howeyc/gopass"

	"bitbucket.org/nstratos/anisync/anisync"
)

//go:generate go run generate/includeagent.go

var (
	malUsername = flag.String("mal-username", "", "Your MyAnimeList `username`, alternatively set MAL_USERNAME.")
	malPassword = flag.String("mal-password", "", "Your MyAnimeList `password`, alternatively set MAL_PASSWORD.")
	hbUsername  = flag.String("hb-username", "", "Your Hummingbird `username`, alternatively set HB_USERNAME.")
)

func findAnimeInListByID(anime, list []anisync.Anime, w io.Writer) {
	matches := 0
	for _, a := range anime {
		found := anisync.FindByID(list, a.ID)
		if found != nil {
			fmt.Fprintf(w, "+++ %7v \t%v\n", a.ID, a.Title)
			matches++
		} else {
			fmt.Fprintf(w, "--- %7v \t%v\n", a.ID, a.Title)
		}
	}
	fmt.Fprintf(w, "Total titles: %d\n", len(anime))
	fmt.Fprintf(w, "Found %d matches on list out of %d\n", matches, len(list))
}

const (
	usageExample1 = `Example 1: anisync -mal-username='AnimeFan' -mal-password='SecurePassword' -hb-username='AnimeFan'`
	usageExample2 = `Example 2: MAL_USERNAME='AnimeFan' MAL_PASSWORD='SecurePassword' HB_USERNAME='AnimeFan' anisync`
)

func customUsage() {
	fmt.Println("Usage: anisync [OPTION]...")
	fmt.Println("Sync your myanimelist.net list with your hummingbird.me anime list.")
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println(usageExample1)
	fmt.Println()
	fmt.Println(usageExample2)
}

func main() {
	flag.Usage = customUsage
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var emptyCredentials = func() bool {
		if *malUsername == "" || *malPassword == "" || *hbUsername == "" {
			return true
		}
		return false
	}
	if emptyCredentials() {
		*malUsername = os.Getenv("MAL_USERNAME")
		*malPassword = os.Getenv("MAL_PASSWORD")
		*hbUsername = os.Getenv("HB_USERNAME")
		if emptyCredentials() {
			flag.Usage()
			return fmt.Errorf("no credentials were provided")
		}
	}

	// malAgent is produced by go generate.
	c := anisync.NewDefaultClient(malAgent)

	malist, _, err := c.GetMyAnimeList(*malUsername)
	if err != nil {
		return fmt.Errorf("could not get myAnimeList %v", err)
	}

	hblist, _, err := c.GetHBAnimeList(*hbUsername)
	if err != nil {
		return fmt.Errorf("could not get Hummingbird list %v", err)
	}

	diff := anisync.Compare(malist, hblist)

	printDiffReport(*diff)

	if len(diff.Missing) == 0 && len(diff.NeedUpdate) == 0 {
		fmt.Println("No anime need to be added or updated in your MyAnimeList.")
		return nil
	}

	fmt.Printf("Proceed with updating and adding missing anime to your MyAnimeList? (y/n) ")
	var answer string
	if _, err := fmt.Scanf("%v\n", &answer); err != nil {
		return err
	}
	proceed := strings.HasPrefix(strings.ToLower(answer), "y")
	if !proceed {
		return nil
	}
	if *malPassword == "" {

		fmt.Printf("Enter MyAnimeList password for user %v:\n", *malUsername)
		pass := gopass.GetPasswdMasked()
		*malPassword = string(pass)
	}

	err = c.VerifyMALCredentials(*malUsername, *malPassword)
	if err != nil {
		return fmt.Errorf("could not verify MAL credentials for user %s", *malUsername)
	}
	fmt.Println("Verification was successful!")
	fmt.Println("Starting Update...")

	syncResult := c.SyncMALAnime(*diff)

	fmt.Printf("%d updated, %d newly added.\n", len(syncResult.Adds), len(syncResult.Updates))
	if len(syncResult.UpdateFails) != 0 {
		fmt.Println("%d failed to be updated.", len(syncResult.UpdateFails))
		for i, updf := range syncResult.UpdateFails {
			fmt.Printf("#%d failed to update (%v %v): %v\n", i+1, updf.Anime.ID, updf.Anime.Title, updf.Error)
		}
	}
	if len(syncResult.AddFails) != 0 {
		fmt.Printf("%d failed to be added.\n", len(syncResult.AddFails))
		for i, addf := range syncResult.AddFails {
			fmt.Printf("#%d failed to add (%v %v): %v\n", i+1, addf.Anime.ID, addf.Anime.Title, addf.Error)
		}
	}

	return nil
}

func printDiffReport(diff anisync.Diff) {
	for _, u := range diff.UpToDate {
		fmt.Printf(">>> %7v \t%v\n", u.ID, u.Title)
	}
	for _, m := range diff.Missing {
		fmt.Printf("--- %7v \t%v\n", m.ID, m.Title)
	}
	for _, u := range diff.NeedUpdate {
		fmt.Printf("<<< %7v \t%v\n", u.Anime.ID, u.Anime.Title)
		printAniDiff(u)
	}
	fmt.Println()
	fmt.Printf("Hummingbird entries: %v\n", len(diff.Right))
	fmt.Printf("MyAnimelist entries: %v\n", len(diff.Left))
	fmt.Printf("(>>>) Up to date: %v\n", len(diff.UpToDate))
	fmt.Printf("(---) Missing: %v\n", len(diff.Missing))
	fmt.Printf("(<<<) Need update: %v\n", len(diff.NeedUpdate))
}

func printAniDiff(d anisync.AniDiff) {
	if d.Status != nil {
		fmt.Printf("    Status: got %v, want %v\n", d.Status.Got, d.Status.Want)
	}
	if d.EpisodesWatched != nil {
		fmt.Printf("    EpisodesWatched: got %v, want %v\n", d.EpisodesWatched.Got, d.EpisodesWatched.Want)
	}
	if d.Rating != nil {
		fmt.Printf("    Rating: got %v, want %v\n", d.Rating.Got, d.Rating.Want)
	}
	if d.Rewatching != nil {
		fmt.Printf("    Rewatching: got %v, want %v\n", d.Rewatching.Got, d.Rewatching.Want)
	}
	if d.LastUpdated != nil {
		fmt.Printf("    LastUpdated: got %v, want %v\n", d.LastUpdated.Got.Local(), d.LastUpdated.Want.Local())
	}
}
