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

func findAnimeInListByTitle(anime, list []anisync.Anime, w io.Writer) {
	matches := 0
	for _, a := range anime {
		found := anisync.FindByTitle(list, a.Title)
		if found != nil {
			fmt.Fprintf(w, "+++ %v\n", a.Title)
			matches++
		} else {
			fmt.Fprintf(w, "--- %v\n", a.Title)
		}
	}
	fmt.Fprintf(w, "Total titles: %d\n", len(anime))
	fmt.Fprintf(w, "Found %d matches on list out of %d\n", matches, len(list))
}

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
	c := anisync.NewClient(malAgent)

	malist, err := c.Anime.ListMAL(*malUsername)
	if err != nil {
		return fmt.Errorf("could not get myAnimeList %v", err)
	}

	hblist, err := c.Anime.ListHB(*hbUsername)
	if err != nil {
		return fmt.Errorf("could not get Hummingbird list %v", err)
	}

	diff := anisync.Compare(malist, hblist)

	printDiffReport(*diff)

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

	err = c.Anime.UpdateMAL(*diff)
	if err != nil {
		return err
	}

	return nil
}

func printDiffReport(diff anisync.Diff) {
	fmt.Println("Missing:")
	for _, m := range diff.Missing {
		fmt.Printf("--- %7v \t%v\n", m.ID, m.Title)
	}
	fmt.Println("Need update:")
	for _, u := range diff.NeedUpdate {
		fmt.Printf("<<< %7v \t%v\n", u.ID, u.Title)
	}
	fmt.Println("Up to date:")
	for _, u := range diff.UpToDate {
		fmt.Printf(">>> %7v \t%v\n", u.ID, u.Title)
	}
	fmt.Println()
	fmt.Printf("Hummingbird entries: %v\n", len(diff.Right))
	fmt.Printf("MyAnimelist entries: %v\n", len(diff.Left))
	fmt.Printf("Missing: %v\n", len(diff.Missing))
	fmt.Printf("Need update: %v\n", len(diff.NeedUpdate))
	fmt.Printf("Up to date: %v\n", len(diff.UpToDate))
}
