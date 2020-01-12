package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

// var channelChIDMap map[string]string

type conf struct {
	Admin   string `yaml:"admin"`
	Channel string `yaml:"channel"`
}

func (c *conf) getConf() *conf {

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func main() {

	var c conf
	c.getConf()

	fmt.Printf("Admin: %v, Channel: %v\n", c.Admin, c.Channel)

	channelChIDMap := make(map[string]string)
	chIDChannelMap := make(map[string]string)
	slackToken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		fmt.Printf("SLACK_TOKEN environment variable is not set\n")
		os.Exit(1)
	}

	api := slack.New(
		slackToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello
			fmt.Printf("Message: %v\n", ev)
			rtm.SendMessage(rtm.NewOutgoingMessage("slack.HelloEvent msg", "CL4H16DFB"))

			channels, err := api.GetChannels(false)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			for _, channel := range channels {
				fmt.Println(channel.Name, channel.ID)
				channelChIDMap[channel.Name] = channel.ID
				chIDChannelMap[channel.ID] = channel.Name
				// channel is of type conversation & groupConversation
				// see all available methods in `conversation.go`
			}
			fmt.Println("channelChIDMap:", channelChIDMap)
			fmt.Println("chIDChannelMap:", chIDChannelMap)

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			rtm.SendMessage(rtm.NewOutgoingMessage("slack.ConnectedEvent msg", "CL4H16DFB"))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("```Current latency: %v```", ev.Value), "CL4H16DFB"))

		case *slack.DesktopNotificationEvent:
			fmt.Printf("Desktop Notification: %v\n", ev)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
