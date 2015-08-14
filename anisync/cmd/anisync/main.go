package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"bitbucket.org/nstratos/anisync/anisync"
)

//go:generate go run generate/includeagent.go

var (
	malUsername = flag.String("mal-username", "", "your MyAnimeList username, alternatively set ANISYNC_MAL_USERNAME")
	malPassword = flag.String("mal-password", "", "your MyAnimeList password, alternatively set ANISYNC_MAL_PASSWORD")
	hbUsername  = flag.String("hb-username", "", "your Hummingbird  username, alternatively set ANISYNC_HB_USERNAME")
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

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: anisync [options] param>\n\n")
		flag.PrintDefaults()
		fmt.Printf("\n")
		fmt.Printf("Example 1: anisync -mal-username='AnimeFan' -mal-password='SecurePassword' -hb-username='AnimeFan'\n\n")
		fmt.Printf("Example 2: ANISYNC_MAL_USERNAME='AnimeFan' ANISYNC_MAL_PASSWORD='SecurePassword' ANISYNC_HB_USERNAME='AnimeFan' anisync\n\n")
	}
	flag.Parse()
	if *malUsername == "" || *malPassword == "" || *hbUsername == "" {
		*malUsername = os.Getenv("ANISYNC_MAL_USERNAME")
		*malPassword = os.Getenv("ANISYNC_MAL_PASSWORD")
		*hbUsername = os.Getenv("ANISYNC_HB_USERNAME")
		if *malUsername == "" || *malPassword == "" || *hbUsername == "" {
			flag.Usage()
			os.Exit(2)
		}
	}

	// malAgent is produced by go generate.
	c := anisync.NewClient(malAgent)
	err := c.VerifyMALCredentials(*malUsername, *malPassword)
	if err != nil {
		log.Fatalf("Could not verify mal credentials for user %s (%s).", *malUsername, err)
	}

	malist, err := c.Anime.ListMAL(*malUsername)
	if err != nil {
		log.Fatalln(err)
	}

	hblist, err := c.Anime.ListHB(*hbUsername)
	if err != nil {
		log.Fatal(err)
	}

	f1, err := os.Create("hb_in_mal.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f1.Close()
	fmt.Fprintln(f1, "*** HB IN MAL ***")
	findAnimeInListByID(hblist, malist, f1)

	f2, err := os.Create("mal_in_hb.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f2.Close()
	fmt.Fprintln(f2, "*** MAL IN HB ***")
	findAnimeInListByID(malist, hblist, f2)

	f3, err := os.Create("mal.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f3.Close()
	for _, mala := range malist {
		fmt.Fprintf(f3, "%7v \t%v\n", mala.ID, mala.Title)
	}
	fmt.Fprintf(f3, "total: %v\n", len(malist))

	f4, err := os.Create("hb.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer f4.Close()
	for _, hba := range hblist {
		fmt.Fprintf(f4, "%7v \t%v\n", hba.ID, hba.Title)
	}
	fmt.Fprintf(f4, "total: %v\n", len(hblist))

}
