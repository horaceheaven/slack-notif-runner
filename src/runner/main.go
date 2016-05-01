package main

import (
	"os/exec"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"os"
	"flag"
)

var log = logrus.New()

var bin = flag.String("bin", "", "binary for the runner to execute")
var scriptPath = flag.String("scriptPath", "~/", "location of the script to be executed")
var scriptArgs = flag.String("scriptArgs", "", "arguments for the script")


// TODO: Add set and tear down steps

func main() {
	flag.Parse()

	if *bin == "" || *scriptPath == "" {
		flag.Usage()
		os.Exit(-1)
	}

	slackNotif("Start of " + *bin + " program runner")
	log.Info("About to execute program runner command...")

	cmd := exec.Command(*bin, *scriptPath, *scriptArgs)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	runErr := cmd.Run()

	log.Info("Finish executing command")

	if runErr != nil {
		slackNotif("Failed to run program, log on to system to review the problem")
		panic(runErr)
	} else {
		log.Info("Successfully executed program")
		slackNotif("Successfully ran program")
	}
}

func slackNotif(msg string) {

	// Does an hard fail if the SLACK_API_KEY environment variables doesn't exist
	api := slack.New(os.Getenv("SLACK_API_KEY"))

	params := slack.PostMessageParameters{}

	_, _, err := api.PostMessage("general", msg, params)

	if err != nil {
		log.Error("Failed to post message to channel")
		panic(err)
	}

	log.Info("Message successfully sent to channel")

}