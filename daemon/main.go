package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mackerelio/go-osstat/uptime"
)

var debug bool
var state_file string
var timeout time.Duration
var command string

func log_debug(msg string) {
	if debug {
		log.Printf("DEBUG: %s\n", msg)
	}
}

func log_warn(msg string) {
	log.Printf("WARN: %s\n", msg)
}

func log_info(msg string) {
	log.Printf("INFO: %s\n", msg)
}

func fail(err error) {
	log.Printf("ERROR: %s\n", err)
	os.Exit(1)
}

func failOnError(err error) {
	if err != nil {
		fail(err)
	}
}

func load_envs() {
	debugEnv := os.Getenv("DEBUG")
	if debugEnv == "1" {
		debug = true
		log_debug("Enabling debug log")
	}

	state_file = os.Getenv("STATE_FILE")
	if state_file == "" {
		fail(errors.New("STATE_FILE env is empty"))
	}
	log_debug(fmt.Sprintf("env STATE_FILE=%q", state_file))

	timeoutEnv := os.Getenv("TIMEOUT")
	if timeoutEnv == "" {
		fail(errors.New("TIMEOUT env is empty"))
	}
	var err error
	timeout, err = time.ParseDuration(timeoutEnv)
	failOnError(err)
	log_debug(fmt.Sprintf("env TIMEOUT=%q (%d seconds)", timeout, int(timeout.Seconds())))

	command = os.Getenv("COMMAND")
	if command == "" {
		fail(errors.New("COMMAND env is empty"))
	}
	log_debug(fmt.Sprintf("env COMMAND=%q", command))
}

func runCommand() {
	log_info(fmt.Sprintf("Running command: %q", command))
	cmd := exec.Command("bash", "-c", command)
	if output, err := cmd.CombinedOutput(); err != nil {
		log_warn(fmt.Sprintf("%s: %s", err.Error(), output))
	}
}

func get_last_unlock() time.Time {
	if fileInfo, err := os.Stat(state_file); errors.Is(err, os.ErrNotExist) {
		up, err := uptime.Get()
		failOnError(err)
		bootTime := time.Now().Add(-1 * up) // .Sub takes a date object, .Add a duration ...
		log_warn(fmt.Sprintf("State file does not exist. Using uptime %q (booted at %q)", up, bootTime))
		return bootTime
	} else {
		log_debug(fmt.Sprintf("State file was modified at %q", fileInfo.ModTime()))
		return fileInfo.ModTime()
	}
}

func main() {
	// dont' log timestampe
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	load_envs()
	log_info("Successfully started")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log_info("Shutting down...")
		os.Exit(0)
	}()

	for {
		log_debug(fmt.Sprintf("Checking inactivity (%s)", state_file))

		lastUnlocked := get_last_unlock()
		triggerCommandAfter := lastUnlocked.Add(timeout)
		log_debug(fmt.Sprintf("Triggering command after %q (timeout %s)", triggerCommandAfter, timeout))

		now := time.Now()
		log_debug(fmt.Sprintf("Now we're at %q", now))

		if now.After(triggerCommandAfter) {
			runCommand()
		}

		if timeout.Seconds() > 5*60 {
			// prod run, check every 10 minutes
			// this value should be not too small
			// when you wake up in the morning and open your Laptop,
			// this will always trigger. You want to have some time
			// to enter the password, before killswitching again
			time.Sleep(time.Minute * 10)
		} else {
			// debug run, check every 15 seconds
			time.Sleep(time.Second * 15)
		}

	}

}
