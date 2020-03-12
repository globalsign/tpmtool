package main

import (
	"flag"
	"fmt"
	"os"
)

// command represents an application command.
type command struct {
	name      string
	flagSet   *flag.FlagSet
	cmdFunc   func()
	usageFunc func()
}

// appName is the application name for log and usage messages.
const (
	appName = "tpmtool"
)

// commands are the application commands.
var commands = []command{
	{
		name:    "help",
		cmdFunc: usageMain,
	},
	{
		name:      "caps",
		flagSet:   fCapsSet,
		cmdFunc:   outputCaps,
		usageFunc: usageCaps,
	},
}

// caps command flag set.
var (
	fCapsSet     = flag.NewFlagSet("caps", flag.ExitOnError)
	fCapsDevice  = fCapsSet.String("device", "/dev/tpmrm0", "TPM device name")
	fCapsSeed    = fCapsSet.Int64("seed", 0, "seed for simulated TPM")
	fCapsHandles = fCapsSet.Bool("handles", false, "handles")
	fCapsAlgs    = fCapsSet.Bool("algorithms", false, "algorithms")
	fCapsAll     = fCapsSet.Bool("all", false, "all")
)

func init() {
	for _, cmd := range commands {
		if cmd.flagSet != nil {
			cmd.flagSet.Usage = cmd.usageFunc
		}
	}
}

// usageError outputs a brief usage message to standard error and exits with
// status code 1.
func usageError() {
	fmt.Fprintf(os.Stderr, "usage: %s <command> [options]\n\n", appName)
	fmt.Printf("Use \"%s help\" for a list of commands.\n", appName)
	os.Exit(1)
}

func usageMain() {
	fmt.Printf("usage: %s <command> [options]\n", appName)
	fmt.Println()

	fmt.Printf("%s is a TPM2.0 command line client.\n", appName)
	fmt.Println()

	const fw = 7
	fmt.Println("Commands:")
	fmt.Printf("    %-*s output selected TPM capabilities\n", fw, "caps")
	fmt.Printf("    %-*s show this usage information\n", fw, "help")
	fmt.Println()

	fmt.Printf("Use \"%s <command> -help\" for more information about a command.\n", appName)
	fmt.Println()
}

// usageCaps outputs usage information for the caps command.
func usageCaps() {
	fmt.Printf("usage: %s caps [options]\n", appName)
	fmt.Println()

	fmt.Println("Caps outputs selected TPM capabilities.")
	fmt.Println()

	const fw = 13
	fmt.Println("Options:")
	fmt.Printf("    -%-*s output supported algorithms\n", fw, "algorithms")
	fmt.Printf("    -%-*s output all capabilities\n", fw, "all")
	fmt.Printf("    -%-*s output active handles\n", fw, "handles")
	fmt.Printf("    -%-*s output this usage information\n", fw, "help")
	fmt.Println()
}
