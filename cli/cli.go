package cli

import (
	"fmt"
	"iter"
	"maps"
	"os"
	"strconv"
)

type parsingState int

const (
	expectCommand = iota
	expectDbFile
	expectServerAddr

	userWantsHelp

	unknownCommand
)

type command struct {
	Flag        parsingState
	Description string
}

var commands = map[string]command{
	"--help": {userWantsHelp, "Displays this list"},
	"--db":   {expectDbFile, "Sets database source file. Expects path"},
	"--addr": {expectServerAddr, "Sets HTTP server address. Expects 'ip:port'"},
}

func findMaxLen(ss iter.Seq[string]) (max int) {
	for s := range ss {
		l := len(s)
		if max < l {
			max = l
		}
	}
	return
}

func printUserHelp() {
	cmdFormat := "%-" + strconv.Itoa(4+findMaxLen(maps.Keys(commands))) + "s %-4s\n"
	fmt.Println("Supported commands:")
	for key, cmd := range commands {
		fmt.Printf(cmdFormat, key, cmd.Description)
	}
}

func parseCommand(arg string) parsingState {
	cmd, exists := commands[arg]
	if exists {
		return cmd.Flag
	}
	return unknownCommand
}

// Returns true if user wants to continue, false otherwise
func Parse(dbFile, serverAddr *string) bool {
	var state parsingState = expectCommand

	fmt.Println(os.Args)

	for i, arg := range os.Args {
		if i == 0 {
			continue // Skip path to executable
		}
		if state == userWantsHelp {
			break
		}

		switch state {
		case expectCommand:
			state = parseCommand(arg)
		case expectDbFile:
			*dbFile = arg
		case expectServerAddr:
			*serverAddr = arg
		case unknownCommand:
			fmt.Println("Unknown command use '--help' to see supported commands list")
			fmt.Println(os.Args)
			return false
		}
	}

	if state == userWantsHelp {
		printUserHelp()
		return false
	}
	return true
}
