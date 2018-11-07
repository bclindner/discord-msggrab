// msggrab.go - grabs a log of user messages from discord.
// mostly to get memes/image links from a specific channel for download
package main

import (
	"github.com/bwmarrin/discordgo" // to handle Discord
	"log" // to log bot functions in console
	"time" // to wait politely between history requests
	"os" // to open files, wait for interrupts, etc
	"os/signal" // to wait for interrupts (keeping the program open)
	"encoding/json"
	"io/ioutil"
	"strings"
	"flag"
)

type Config struct {
	// Channels to scrape.
	Channels []string `json:"channels"`

	// Token for the bot to scrape with.
	BotToken string `json:"botToken"`

	// Amount of posts to parse per loop.
	AmountPerLoop int `json:"amountPerLoop"`

	// Time to wait for each loop (set this to at least 1, please!)
	TimeToWait time.Duration `json:"timeToWait"`
}

func main() {
	// parse args
	configFile := flag.String("conf", "msggrab.json", "Config file for the bot")
	outFile := flag.String("out", "msggrab.log", "Output file to put the links in")
	flag.Parse()

	// read the config file
	configFileStream, err := ioutil.ReadFile(*configFile)
	if err != nil { log.Fatal(err) }
	// parse it
	config := Config{}
	err = json.Unmarshal(configFileStream, &config)
	if err != nil { log.Fatal(err) }

	// open the outfile to write (create if it doesn't exist)
	outFileStream, err := os.OpenFile(*outFile, os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil { log.Fatal(err) }
	// make it close when we're done
	defer outFileStream.Close()

	// start the bot session
	bot, err := discordgo.New(config.BotToken)
	if err != nil { log.Fatal(err) }

	// run the scraper function for each channel
	for _, channel := range config.Channels {
		ScrapeLinksToFile(bot, channel, config.AmountPerLoop, config.TimeToWait, outFileStream)
	}

	// log and quit when we're done
	log.Println("Done parsing messages. Goodbye")
	os.Exit(0)
	// block until there's an OS interrupt or kill
	// this gives our function time to work
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}

func ScrapeLinksToFile(bot *discordgo.Session, channel string, amt int, waittime time.Duration, file *os.File) {
	log.Println("Scraping channel with ID",channel)
	file.WriteString("-----BEGIN CHANNEL "+channel+"-----\n")
	// initialize a counter for messages parsed (for logging)
	messagesParsed := 0
	// set lastMessage to empty so it starts from the most recent message
	lastMessage := ""
	// initialize the history buffer to ensure the for loop doesn't end early
	history, err := bot.ChannelMessages(channel, amt, lastMessage, "", "")
	if err != nil { log.Fatal(err) }

	// run this until there are no messages left in the history buffer
	for len(history) != 0 {
		// for every message in the current history buffer
		for _, msg := range history {
			// get the links and print them to the file
			links := GetLinks(msg)
			for _, link := range links {
				file.WriteString(link+"\n")
			}
			// set the last id after each message
			// hacky way of getting the last ID but it works for a script this quick
			// still probably quicker than a node or python bodge lmao
			lastMessage = msg.ID
		}
		// wait politely between requests for the channel messages
		// 2 secs seems ok, hell 1 is probably fine
		// so long as i'm not flooding their servs with requests we should be ok
		time.Sleep(waittime * time.Second)

		// add the number of messages parsed to counter & log it
		messagesParsed += len(history)
		log.Println("Messages parsed:",messagesParsed)
		// reload the history buffer starting after the last thing we got before
		history, err = bot.ChannelMessages(channel, amt, lastMessage, "", "")
		if err != nil { log.Fatal(err) }
	}
	file.WriteString("-----END CHANNEL "+channel+"-----\n")
}

func GetLinks(msg *discordgo.Message) (links []string){
	// if there is an HTTP(S) link in there, print it
	if len(msg.Content) > 0 && strings.Contains(msg.Content, "http") {
		links = append(links, msg.Content)
	}
	// also get attachment URLs if available (this will get uploaded stuff)
	if len(msg.Attachments) > 0 {
		for _, att := range msg.Attachments {
			links = append(links, att.URL)
		}
	}
	return
}
