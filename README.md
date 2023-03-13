# Anisync

A web application and command line tool that can transfer a [Kitsu.io](https://kitsu.io) anime list
to [MyAnimeList.net](https://myanimelist.net/).

Live demo: https://ani-sync.appspot.com/

## Screenshots

![anisync checking](/screenshots/anisync-check.png?raw=true "Checking for updates")

![anisync after sync](/screenshots/anisync-sync.png?raw=true "After pressing the Sync button")

## Installation

Install Cloud SDK: https://cloud.google.com/sdk/docs/install

    gcloud auth login
    gcloud components install app-engine-go

## Deployment

    gcloud app deploy --project=ani-sync
    gcloud app browse --project=ani-sync

Note: On Windows the path for the gcloud command by default only works on
Command Prompt and not PowerShell.