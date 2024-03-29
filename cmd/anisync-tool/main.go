package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nstratos/go-kitsu/kitsu"
	"github.com/nstratos/go-myanimelist/mal"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/nstratos/anisync/anisync"
)

var (
	kitsuUserID = flag.String("kitsuid", "", "Kitsu.io user ID (or set KITSU_USER_ID)")
	malUsername = flag.String("malu", "", "MyAnimeList.net username (or set MAL_USERNAME)")
	malPassword = flag.String("malp", "", "MyAnimeList.net password (or set MAL_PASSWORD)")
	yesFlag     = flag.Bool("y", false, "answer yes in final confirmation")
	helpFlag    = flag.Bool("help", false, "show detailed help message")
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

const help = `anisync-tool: Sync a Kitsu.io anime list back to MyAnimeList.net.

Usage: anisync-tool [options]...

Options:

  -kitsuid Kitsu.io user ID
  -malu    MyAnimeList.net username
  -malp    MyAnimeList.net password
  -y       answer yes in final confirmation
  -help    show detailed help message

By default, the program will ask for any credentials not provided by the
options and will ask for confirmation one final time before syncing. The -y
flag is useful in case the program is intended to be used without user
interaction provided that all the required credentials are  made available
either through options or environment variables.

Examples:

% anisync-tool -kitsuid='AnimeFan'

  Only the Kitsu.io user ID is provided. The program will ask for the
  MyAnimeList.net username, password and confirmation before syncing.

% anisync-tool -y -kitsuid='AnimeFan' -malu='AnimeFan' -malp='password'

  All the credentials are provided through options. The program will not ask
  for confirmation before syncing.

% KITSU_USER_ID='AnimeFan' MAL_USERNAME='AnimeFan' MAL_PASSWORD='password' anisync-tool

  All the credentials are provided through environment variables. The program
  will ask for confirmation before syncing.

`

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: anisync-tool [options]...")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if *helpFlag {
		fmt.Fprint(os.Stderr, help)
		os.Exit(2)
	}

	if *kitsuUserID == "" {
		*kitsuUserID = os.Getenv("KITSU_USER_ID")
	}
	if *malUsername == "" {
		*malUsername = os.Getenv("MAL_USERNAME")
	}
	if *malPassword == "" {
		*malPassword = os.Getenv("MAL_PASSWORD")
	}

	if *kitsuUserID == "" {
		sc := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter Kitsu.io user ID: ")
		sc.Scan()
		*kitsuUserID = sc.Text()
	}

	if *malUsername == "" {
		sc := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter MyAnimeList.net username: ")
		sc.Scan()
		*malUsername = sc.Text()
	}

	resources := anisync.NewResources(
		mal.NewClient(mal.Auth(*malUsername, *malPassword)),
		kitsu.NewClient(nil),
	)
	c := anisync.NewClient(resources)

	myAnimeList, _, err := c.GetMyAnimeList(*malUsername)
	if err != nil {
		return fmt.Errorf("could not get MyAnimeList.net anime list %v", err)
	}

	kitsuList, _, err := c.GetKitsuAnimeList(*kitsuUserID)
	if err != nil {
		return fmt.Errorf("could not get Kitsu.io anime list %v", err)
	}

	diff := anisync.Compare(myAnimeList, kitsuList)

	printDiffReport(*diff)

	if len(diff.Missing) == 0 && len(diff.NeedUpdate) == 0 {
		fmt.Printf("No anime need to be added or updated in MyAnimeList.net account %q.\n", *malUsername)
		return nil
	}

	proceed := false
	if !*yesFlag {
		sc := bufio.NewScanner(os.Stdin)
		fmt.Printf("Do you want to continue? [y/N] ")
		sc.Scan()
		answer := sc.Text()
		proceed = strings.HasPrefix(strings.ToLower(answer), "y")
	} else {
		proceed = true
	}
	// nothing to do
	if !proceed {
		return nil
	}

	if *malPassword == "" {
		fmt.Printf("Enter MyAnimeList.net password for username %v:\n", *malUsername)
		pass, err := terminal.ReadPassword(0)
		if err != nil {
			return fmt.Errorf("reading password: %v", err)
		}
		*malPassword = string(pass)
	}

	if _, _, err := c.VerifyMALCredentials(*malUsername, *malPassword); err != nil {
		return fmt.Errorf("MyAnimeList.net username and password do not match")
	}
	fmt.Println("Verification was successful!")

	fmt.Println("Starting Update...")

	syncResult := c.SyncMALAnime(*diff)

	fmt.Printf("%d updated, %d newly added.\n", len(syncResult.Updates), len(syncResult.Adds))
	if len(syncResult.UpdateFails) != 0 {
		fmt.Printf("%d failed to be updated.\n", len(syncResult.UpdateFails))
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
		fmt.Printf("(===) %7v \t%v\n", u.ID, u.Title)
	}
	for _, u := range diff.Uncertain {
		fmt.Printf("( < ) %7v \t%v\n", u.Anime.ID, u.Anime.Title)
		printAniDiff(u)
	}
	for _, m := range diff.Missing {
		fmt.Printf("(---) %7v \t%v\n", m.ID, m.Title)
	}
	for _, u := range diff.NeedUpdate {
		fmt.Printf("(<<<) %7v \t%v\n", u.Anime.ID, u.Anime.Title)
		printAniDiff(u)
	}
	fmt.Println()
	fmt.Printf("Kitsu entries: %v\n", len(diff.Right))
	fmt.Printf("MyAnimelist entries: %v\n", len(diff.Left))
	fmt.Printf("(===) Up to date: %v\n", len(diff.UpToDate))
	fmt.Printf("( < ) Okay: %v\n", len(diff.Uncertain))
	fmt.Printf("(---) Missing: %v\n", len(diff.Missing))
	fmt.Printf("(<<<) Need update: %v\n", len(diff.NeedUpdate))
	fmt.Println("After this operation, there will be:")
	fmt.Printf("%v updated and %v newly added anime on MyAnimeList.net account %q.\n", len(diff.NeedUpdate), len(diff.Missing), *malUsername)
}

func printAniDiff(d anisync.AniDiff) {
	if d.Status != nil {
		fmt.Printf("\t\t|-> Status: got %v, want %v\n", d.Status.Got, d.Status.Want)
	}
	if d.EpisodesWatched != nil {
		fmt.Printf("\t\t|-> EpisodesWatched: got %v, want %v\n", d.EpisodesWatched.Got, d.EpisodesWatched.Want)
	}
	if d.Rating != nil {
		fmt.Printf("\t\t|-> Rating: got %v, want %v\n", d.Rating.Got, d.Rating.Want)
	}
	if d.Rewatching != nil {
		fmt.Printf("\t\t|-> Rewatching: got %v, want %v\n", d.Rewatching.Got, d.Rewatching.Want)
	}
	if d.LastUpdated != nil {
		fmt.Printf("\t\t|-> LastUpdated: got %v, want %v\n", d.LastUpdated.Got.Local(), d.LastUpdated.Want.Local())
	}
}
