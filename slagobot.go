package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

// var channelChIDMap map[string]string

type conf struct {
	Admin   string `yaml:"admin"`
	Channel string `yaml:"channel"`
}

func (c *conf) getConf(configPath string) *conf {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("Read config error: %v\n", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		// log.Fatalf("Unmarshal: %v", err)
		log.Printf("Unmarshal error: %v\n", err)
	}

	return c
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Getting outbound address error: %v\n", err)
		return "can't determinate IP address"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return fmt.Sprintf("%v", localAddr.IP)
}

func main() {
	configPathPtr := flag.String("config-path", "config.yaml", "path to the config file")
	slackTokenPtr := flag.String("slack-token", "xoxb", "slack bot token")
	flag.Parse()
	var slackToken string
	var defaultChannelName string = "random"
	var defaultChannelID string
	var configuration conf
	channelChIDMap := make(map[string]string)
	chIDChannelMap := make(map[string]string)

	configuration.getConf(*configPathPtr)
	outboundIP := getOutboundIP()
	fmt.Printf("Outbound IP: %v\n", outboundIP)

	if *slackTokenPtr == "xoxb" {
		fmt.Println("slack-token flag not passed")
		fmt.Println("checking environment variable...")
		var ok bool
		slackToken, ok = os.LookupEnv("SLACK_TOKEN")
		if !ok {
			fmt.Println("SLACK_TOKEN environment variable is not set")
			os.Exit(1)

		}
	} else {
		slackToken = *slackTokenPtr
	}

	api := slack.New(
		slackToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	channels, err := api.GetChannels(false)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, channel := range channels {
		fmt.Println(channel.Name, channel.ID)
		channelChIDMap[channel.Name] = channel.ID
		chIDChannelMap[channel.ID] = channel.Name
	}
	fmt.Println("channelChIDMap:", channelChIDMap)
	fmt.Println("chIDChannelMap:", chIDChannelMap)

	if configuration.Channel != "" {
		defaultChannelName = configuration.Channel
	}

	defaultChannelID = channelChIDMap[defaultChannelName]

	fmt.Printf("Admin: %v, Default channel: %v\n", configuration.Admin, configuration.Channel)

	users, err := api.GetUsers()
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	for _, user := range users {
		fmt.Println(user.Name, user.RealName, user.ID)
	}

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			fmt.Printf("Hello event: %v\n", ev)
			rtm.SendMessage(rtm.NewOutgoingMessage("slack.HelloEvent msg", defaultChannelID))

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			rtm.SendMessage(rtm.NewOutgoingMessage("slack.ConnectedEvent msg", defaultChannelID))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)
			fmt.Println("Msg:", ev.Msg)
			fmt.Println("Msg.User:", ev.Msg.User)
			fmt.Println("Text:", ev.Text)
			if ev.Msg.User == configuration.Admin && ev.Text == "!ip" {
				rtm.SendMessage(rtm.NewOutgoingMessage("my ip: "+string(outboundIP), defaultChannelID))
			}

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("```Current latency: %v```", ev.Value), defaultChannelID))

		case *slack.DesktopNotificationEvent:
			fmt.Printf("Desktop Notification: %v\n", ev)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			// Other events..
			fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
