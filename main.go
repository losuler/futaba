package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"
	"syscall"
	"io/ioutil"
	"regexp"
	"github.com/go-yaml/yaml"
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Discord Discord `yaml:"discord"`
	Users []Users `yaml:"users"`
}

type Discord struct {
	Token string `yaml:"token"`
}

type Users struct {
	Username string `yaml:"username"`
	UserID string `yaml:"userid"`
	Timezone string	`yaml:"timezone"`
	Nicknames string `yaml:"nicknames"`
}

func readConfig(configFile string) Config {
	var c Config
    
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	} 

	err = yaml.Unmarshal([]byte(raw), &c)

	return c
}

func getAcc(c Config, userMatch []string) Users {
	var user Users

	for _, user := range c.Users {
		// Check all submatches.
		for _, subMatch := range userMatch {
			if subMatch == user.Username {
				return user
			}
		}
	}
	return user
}

func getTime(account Users) string {
	tz, err := time.LoadLocation(account.Timezone)
	if err != nil {
		fmt.Println(err)
	}

	timeNow := time.Now().In(tz)
	dayTime := timeNow.Format("Monday 3:04PM")

	return dayTime
}

func messageRecieve(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check the message is not from the bot.
	if m.Author.ID == s.State.User.ID {
		return
	}

	c := readConfig("config.yml")

	timeReg := regexp.MustCompile(`t(ime)\.(.*)`)

	if timeReg.MatchString(m.Content) {
		userMatch := timeReg.FindStringSubmatch(m.Content)
		account := getAcc(c, userMatch)
		
		// If no account was returned.
		if account.Username != "" {
			dayTime := getTime(account)
			msg := fmt.Sprintf("It's %s where %s is.", dayTime, account.Username)
			
			_, err := s.ChannelMessageSend(m.ChannelID, msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func main() {
	c := readConfig("config.yml")

	dg, err := discordgo.New("Bot " + c.Discord.Token)
	if err != nil {
		panic(err)
	}

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	dg.AddHandler(messageRecieve)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sigChan

	dg.Close()
}