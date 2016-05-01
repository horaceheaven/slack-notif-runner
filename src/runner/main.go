package main

import (
	"os/exec"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"os"
	"flag"
	"os/signal"
	"syscall"
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

	go func() {
		slackNotif("Start of " + *bin + " program runner", *channel)
		log.Info("About to execute program runner command...")

		cmd := exec.Command(*bin, *scriptPath, *scriptArgs)

		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		runErr := cmd.Run()

		log.Info("Finish executing command")

		if runErr != nil {
			log.Info("Failed to finish program execution ", runErr)
			slackNotif("Failed to run program, log on to system to review the problem", *channel)
		} else {
			log.Info("Successfully executed program")
			slackNotif("Successfully ran program", *channel)
		}

		done <- true
	}()


	go func() {
		for sig := range osSig {
			log.Info(sig)
			slackNotif(sig.String(), *channel)
		}

		done <- true
	}()

	<- done
}

func slackNotif(msg string, channel string) {

	// Does an hard fail if the SLACK_API_KEY environment variables doesn't exist
	api := slack.New(os.Getenv("SLACK_API_KEY"))

	params := slack.PostMessageParameters{}

	_, _, err := api.PostMessage(channel, msg, params)

	if err != nil {
		log.Error("Failed to post message to channel")
		panic(err)
	}

	log.Info("Sent message to slack")

}