# Golang Slack Bot

![Go](https://github.com/vensder/slack-go-bot/workflows/Go/badge.svg) [![Docker](https://img.shields.io/docker/cloud/build/vensder/slack-go-bot)](https://hub.docker.com/r/vensder/slack-go-bot)

## How to run pre-build docker

```bash
docker run --rm --name slagobot -d vensder/slack-go-bot -slack-token xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx
```

Or use environment variable "SLACK_TOKEN":

```bash
export SLACK_TOKEN='xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx'
docker run --rm --name slagobot -d vensder/slack-go-bot
```

![alt text](./img/screenshot_1.png?raw=true)


## How to run from source code

```bash
git clone https://github.com/vensder/slack-go-bot.git
cd slack-go-bot
```

Run this bot and pass your Slack token:

```bash
go run slagobot.go -slack-token xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx
```

Or use environment variable "SLACK_TOKEN":

```bash
export SLACK_TOKEN='xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx'
go run slagobot.go
```

## How to build executalbe on Linux and run it

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o slagobot -v -x slagobot.go
chmod +x slagobot
./slagobot.go -slack-token xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx
```

## How to build docker image and run container

```bash
docker build -t slagobot .
docker run --rm --name slagobot -d slagobot -slack-token  xoxb-000000-xxxxxxxx-xxxxxxxxxxxxxxx
```
