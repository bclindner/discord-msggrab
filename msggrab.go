// msggrab.go - grabs a log of user messages from discord.
// mostly to get memes/image links from a specific channel for download
package main

import (
	"flag"
	"github.com/bwmarrin/discordgo" // to handle Discord
	"log"                           // to log bot functions in console
	"os"                            // to open files, wait for interrupts, etc
	"strings"
	"time" // to wait politely between history requests
	"regexp"
)

var regex regexp.Regexp

func main() {
	regex = *regexp.MustCompile(`https?://[\S]+`)
	// parse args
	botToken := flag.String("t", "", "Bot token to log in with.")
	outFile := flag.String("o", "msggrab.log", "Output file to put the links in")
	amountPerLoop := flag.Int("a", 20, "Amount of messages to get per second.")
	flag.Parse()
	channels := flag.Args()
	// ensure required args are set
	if *botToken == "" {
		log.Fatal("Bot token not specified (specify with -t).")
	}
	if len(channels) == 0 {
		log.Fatal("No channels specified.")
	}
	// open the outfile to write (create if it doesn't exist)
	outFileStream, err := os.OpenFile(*outFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// make it close when we're done
	defer outFileStream.Close()

	// start the bot session
	bot, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatal("Error starting discordgo session:", err)
	}
	// make the bot appear online
	bot.Open()
	defer bot.Close()
	// run the scraper function for each channel
	for _, channel := range channels {
		// use channels to parse stuff because why not?
		lines := make(chan string)
		go ScrapeLinks(bot, channel, *amountPerLoop, lines)
		for line := range lines {
			outFileStream.WriteString(line + "\n")
		}
	}

	// log and quit when we're done
	log.Println("Done parsing messages. Goodbye")
	os.Exit(0)
}

func ScrapeLinks(bot *discordgo.Session, channel string, amt int, lines chan<- string) {
	c, _ := bot.Channel(channel)
	log.Println("Scraping channel", channel, "(#" + c.Name + ")")
	bot.UpdateStatus(1, "#"+c.Name)
	lines <- "-----------BEGIN CHANNEL " + channel + " (#"+ c.Name + ")-----------"
	// initialize a counter for messages parsed (for logging)
	messagesParsed := 0
	linksSent := 0
	// set lastMessage to empty so it starts from the most recent message
	lastMessage := ""
	// initialize the history buffer to ensure the for loop doesn't end early
	history, err := bot.ChannelMessages(channel, amt, lastMessage, "", "")
	if err != nil {
		log.Fatal(err)
	}
	// run this until there are no messages left in the history buffer
	for len(history) != 0 {
		// for every message in the current history buffer
		for _, msg := range history {
			// get the links and print them to the file
			links := GetLinks(msg)
			for _, link := range links {
				lines <- link
				linksSent++
			}
			// set the last id after each message
			// hacky way of getting the last ID but it works for a script this quick
			// still probably quicker than a node or python bodge lmao
			lastMessage = msg.ID
		}
		// wait politely between requests for the channel messages
		// so long as i'm not flooding their servs with requests we should be ok
		time.Sleep(time.Second)

		// add the number of messages parsed to counter & log it
		messagesParsed += len(history)
		log.Println("Messages parsed:", messagesParsed)
		log.Println("Messages saved:", linksSent)
		// reload the history buffer starting after the last thing we got before
		history, err = bot.ChannelMessages(channel, amt, lastMessage, "", "")
		if err != nil {
			log.Fatal(err)
		}
	}
	lines <- "-----------END CHANNEL " + channel + " (#"+ c.Name + ")-----------"
	log.Println("Scraping channel", channel, "(#"+c.Name+")")
	close(lines)
}

func GetLinks(msg *discordgo.Message) (links []string) {
	// if there is an HTTP(S) link in there, print it
	if len(msg.Content) > 0 && strings.Contains(msg.Content, "http") {
		for _, link := range regex.FindAllString(msg.Content, -1) {
			links = append(links, link)
		}
	}
	// also get attachment URLs if available (this will get uploaded stuff)
	if len(msg.Attachments) > 0 {
		for _, att := range msg.Attachments {
			links = append(links, att.URL)
		}
	}
	return
}
