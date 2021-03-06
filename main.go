package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	token         string
	quotes        map[string]string
	pathQuotes    string = "data\\quotes.txt"
	currentAuthor string
	currentQuoute string
	userRank      map[discordgo.User]int
	randomWord    string
	f             *os.File
)

func init() {
	flag.StringVar(&token, "t", "", "Bot token")
	flag.Parse()
}

func main() {
	quotes = ParseFile(pathQuotes)
	fmt.Printf("\nlength: %d\n", len(quotes))
	userRank = make(map[discordgo.User]int)
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session %v", err)
	}
	randomWord = "Okay"
	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection %v", err)
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, "!SetRandomWord") {
		randomWord = (strings.TrimSpace(m.Content[15:len(m.Content)]))
	}
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn((15 - 1) + 1)
	if number == 1 {
		sendChannelMessage(s, m, randomWord)
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!quote" {
		fmt.Printf("1\n")
		currentQuoute, currentAuthor = FindRandomQuote()
		fmt.Printf("\ncurrentQuoute: %s\n", currentQuoute)
		fmt.Printf("currentAuthor: %s\n", currentAuthor)
		fmt.Printf("2\n")
		output := fmt.Sprintf("--- Guess the quote! ---  \n%s", currentQuoute)
		fmt.Printf("3\n")
		sendChannelMessage(s, m, output)

	} else if strings.HasPrefix(m.Content, "!quote guess") {
		if m.Content == "!quote guess" {
			sendChannelMessage(s, m, "You have to guess, dummy!")
		} else {

			guess := strings.Title(strings.TrimSpace(m.Content[13:len(m.Content)]))

			if strings.EqualFold(guess, currentAuthor) {
				output := fmt.Sprintf("Congratulations, %s is the correct person! :)", currentAuthor)
				GiveUserPoint(m.Author)
				currentAuthor = ""
				currentQuoute = ""

				sendChannelMessage(s, m, output)
			} else {
				output := fmt.Sprintf("I'm sorry, %s is not the correct person :(", guess)

				sendChannelMessage(s, m, output)
			}
		}
	} else if m.Content == "!quote all" {
		sendChannelMessage(s, m, GetAllQuotes())
	} else if m.Content == "!quote help" {
		output := fmt.Sprintf("Hello %s and welcome to G??gler bot!\n\n"+
			"The following commands are supported at the moment:\n"+
			"!quote : Returns a random quote\n"+
			"!quote guess <name> : A guess to the last random quote\n"+
			"!quote all : Returns all quotes (but without the authors, otherwise that would be cheating!)\n"+
			"!tue : Returns \"Yo, fuck Tue!\"\n"+
			"!tue send : Sends \"Yo, fuck Tue!\" to Tue in a private dm", m.Author.Username)

		sendChannelMessage(s, m, output)
	} else if m.Content == "!quote rank" {
		sendChannelMessage(s, m, DisplayRanks(s, m))

	} else if m.Content == "!tue" {
		sendChannelMessage(s, m, "Yo, fuck Tue!")
	} else if m.Content == "!tue send" {
		a, err := s.UserChannelCreate("245253768021540864")
		if err != nil {
			fmt.Printf("Error in sending to user: %v", err)
		}
		s.ChannelMessageSend(a.ID, "Yo, fuck Tue!")
		sendChannelMessage(s, m, "Successfully sent message to Tue")

	} else if strings.HasPrefix(m.Content, "!quote ") {
		authorToSearchQuotesFor := strings.TrimSpace(m.Content[7:len(m.Content)])
		sendChannelMessage(s, m, GetAllQuotesFromAuthor(authorToSearchQuotesFor))
	}
	if strings.HasPrefix(m.Content, "!add quote ") {
		f.WriteString(m.Content[13:len(m.Content)])
	}

}

func sendChannelMessage(s *discordgo.Session, m *discordgo.MessageCreate, output string) {
	_, err := s.ChannelMessageSend(m.ChannelID, output)
	if err != nil {
		log.Fatal(err)
	}
}

func GiveUserPoint(author *discordgo.User) {
	if _, ok := userRank[*author]; ok {
		userRank[*author]++
	} else {
		userRank[*author] = 1
	}
}

func DisplayRanks(s *discordgo.Session, m *discordgo.MessageCreate) string {
	var ranks strings.Builder
	ranks.WriteString("--- Quote rankings! ---\n")
	for key, value := range userRank {
		ranks.WriteString(fmt.Sprintf("%s = %d\n", key.Username, value))
	}

	if ranks.Len() == 1 {
		return "Doesn't seem like there are any users with points :c"
	} else {
		return ranks.String()
	}

}

func ParseFile(path string) map[string]string {
	f, err := os.OpenFile("data\\quotes.txt", os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		log.Fatalf("Error in opening file: %v", err)
	}

	defer f.Close()

	data, err := ioutil.ReadFile("data\\quotes.txt")
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	fmt.Printf("\nLength: %d bytes", len(data))
	fmt.Printf("\nData: %s", data)
	fmt.Printf("\nError: %v", err)

	sc := bufio.NewScanner(f)
	sc.Split(bufio.ScanLines)

	quouteToAuthor := make(map[string]string)

	for sc.Scan() {

		line := sc.Text()
		split := strings.Split(line, ":")

		quouteToAuthor[split[0]] = split[1]
	}
	fmt.Printf("\nquoteToAuthor: %s\n", quouteToAuthor)
	return quouteToAuthor
}

func GetAllQuotes() string {
	var allQuotes strings.Builder
	for key := range quotes {
		allQuotes.WriteString(fmt.Sprintf("\"%s \"\n", key))
	}
	return allQuotes.String()
}

func GetAllQuotesAndAuthors() string {
	var allQuotes strings.Builder
	for key, value := range quotes {
		allQuotes.WriteString(fmt.Sprintf("\"%s \" -%s \n", key, value))
	}
	return allQuotes.String()
}

func GetAllQuotesFromAuthor(author string) string {
	var allQuotes strings.Builder
	allQuotes.WriteString(fmt.Sprintf("%s has said all of these quotes:\n", strings.Title(author)))

	for key, value := range quotes {
		if strings.EqualFold(value, author) {
			allQuotes.WriteString(fmt.Sprintf("\"%s \"\n", key))
		}
	}

	if allQuotes.Len() == 1 {
		return fmt.Sprintf("Hmm, doesn't look like %s has any quotes :(", author)
	} else {
		return allQuotes.String()
	}
}

func FindRandomQuote() (string, string) {
	rand.Seed(time.Now().UnixNano())

	var list []string
	for key := range quotes {
		list = append(list, key)
	}
	fmt.Printf("\nlist length: %d\n", len(list))
	randomNumber := rand.Intn((len(list) - 0) + 1)

	// Returns quote and author tuple
	fmt.Printf("\nauthor: %s", list[randomNumber])
	return fmt.Sprintf("\"%s\"", list[randomNumber]), quotes[list[randomNumber]]
}
