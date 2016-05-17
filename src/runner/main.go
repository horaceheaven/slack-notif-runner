package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"github.com/robfig/cron"
)

var log = logrus.New()

var channel = flag.String("channel", "general", "slack channel to post notifications")
var bin = flag.String("bin", "", "binary for the runner to execute")
var scriptPath = flag.String("scriptPath", "~/", "location of the script to be executed")
var scriptArgs = flag.String("scriptArgs", "", "arguments for the script")

// TODO: Add setup and tear down steps

func main() {
	flag.Parse()

	if *bin == "" || *scriptPath == "" {
		flag.Usage()
		os.Exit(-1)
	}

	osSig := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(osSig, syscall.SIGINT, syscall.SIGTERM)

	slackClient := getSlackClient()

	cron := cron.New()

	cron.AddFunc("5 * * * * *", func() {
		sendSlackMessage(slackClient, "Start of "+*bin+" program runner", *channel)
		log.Info("About to execute program runner command...")

		cmd := exec.Command(*bin, *scriptPath, *scriptArgs)

		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		runErr := cmd.Run()

		log.Info("Finish executing command")

		if runErr != nil {
			log.Info("Failed to finish program execution ", runErr)
			sendSlackMessage(slackClient, "Failed to run program, log on to system to review the problem", *channel)
		} else {
			log.Info("Successfully executed program")
			sendSlackMessage(slackClient, "Successfully ran program", *channel)
		}

		done <- true
	})

	cron.Start()

	go func() {
		for sig := range osSig {
			log.Info(sig)
			sendSlackMessage(slackClient, "Runner stopped by the following signal: "+sig.String(), *channel)
		}

		done <- true
	}()

	<-done
}

func getSlackClient() *slack.Client {
	apiKey, exist := os.LookupEnv("SLACK_API_KEY")

	if !exist {
		log.Error("Please set the SLACK_API_KEY")
	}

	return slack.New(apiKey)
}

func sendSlackMessage(client *slack.Client, msg string, channel string) {

	params := slack.PostMessageParameters{}

	_, _, err := client.PostMessage(channel, msg, params)

	if err != nil {
		log.Error("Failed to post message to channel")
		panic(err)
	} else {
		log.Debug(msg)
	}

	log.Info("Sent message to slack")
}
