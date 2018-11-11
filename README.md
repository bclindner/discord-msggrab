# msggrab

This is a simple Discord scraper bot written in Go to learn a little about its concurrency model.
It grabs links from a given Discord channel ID and prints them to a file (default msggrab.log).

## Usage

Usage should go something like:
```
msggrab -t <YOUR_BOT_TOKEN> <CHANNELID_TO_SCRAPE> <CHANNELID_TO_SCRAPE>...
```

Full options can be checked with `msggrab -h`.
