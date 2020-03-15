package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/paulgriffiths/pgtpm"
)

// command represents an application command.
type command struct {
	name      string
	flagSet   *flag.FlagSet
	cmdFunc   func()
	usageFunc func()
}

// handleFlag implments flag.Value and contains a TPM handle.
type handleFlag pgtpm.Handle

// Global constants.
const (
	appName          = "tpmtool"
	defaultTPMDevice = "/dev/tpmrm0"
)

// Command name constants.
const (
	capsCommand       = "caps"
	helpCommand       = "help"
	makeCredCommand   = "makecred"
	readPublicCommand = "readpublic"
)

// Flag name constants.
const (
	algsFlagName       = "algorithms"
	allFlagName        = "all"
	credOutFlagName    = "credout"
	handleFlagName     = "handle"
	handlesFlagName    = "handles"
	helpFlagName       = "help"
	inFlagName         = "in"
	outFlagName        = "out"
	publicAreaFlagName = "publicarea"
	pubOutFlagName     = "pubout"
	secretOutFlagName  = "secretout"
	textFlagName       = "text"
	tpmFlagName        = "tpm"
)

// commands are the application commands.
var commands = []command{
	{
		name:    helpCommand,
		cmdFunc: usageMain,
	},
	{
		name:      capsCommand,
		flagSet:   fCapsSet,
		cmdFunc:   outputCaps,
		usageFunc: usageCaps,
	},
	{
		name:      makeCredCommand,
		flagSet:   fMakeCredSet,
		cmdFunc:   makeCred,
		usageFunc: usageMakeCred,
	},
	{
		name:      readPublicCommand,
		flagSet:   fReadPublicSet,
		cmdFunc:   readPublic,
		usageFunc: usageReadPublic,
	},
}

// caps command flag set.
var (
	fCapsSet     = flag.NewFlagSet(capsCommand, flag.ExitOnError)
	fCapsAlgs    = fCapsSet.Bool(algsFlagName, false, "")
	fCapsAll     = fCapsSet.Bool(allFlagName, false, "")
	fCapsHandles = fCapsSet.Bool(handlesFlagName, false, "")
	fCapsHelp    = fCapsSet.Bool(helpFlagName, false, "")
	fCapsTPM     = fCapsSet.String(tpmFlagName, defaultTPMDevice, "")
)

// makecred command flag set.
var (
	fMakeCredSet        = flag.NewFlagSet(makeCredCommand, flag.ExitOnError)
	fMakeCredHandle     handleFlag
	fMakeCredHelp       = fMakeCredSet.Bool(helpFlagName, false, "p")
	fMakeCredIn         = fMakeCredSet.String(inFlagName, "", "")
	fMakeCredCredOut    = fMakeCredSet.String(credOutFlagName, "", "")
	fMakeCredPublicArea = fMakeCredSet.String(publicAreaFlagName, "", "")
	fMakeCredSecretOut  = fMakeCredSet.String(secretOutFlagName, "", "")
	fMakeCredTPM        = fMakeCredSet.String(tpmFlagName, defaultTPMDevice, "")
)

// readpublic command flag set.
var (
	fReadPublicSet    = flag.NewFlagSet(readPublicCommand, flag.ExitOnError)
	fReadPublicHandle handleFlag
	fReadPublicHelp   = fReadPublicSet.Bool(helpFlagName, false, "p")
	fReadPublicIn     = fReadPublicSet.String(inFlagName, "", "")
	fReadPublicOut    = fReadPublicSet.String(outFlagName, "", "")
	fReadPublicPubOut = fReadPublicSet.Bool(pubOutFlagName, false, "")
	fReadPublicText   = fReadPublicSet.Bool(textFlagName, false, "")
	fReadPublicTPM    = fReadPublicSet.String(tpmFlagName, defaultTPMDevice, "")
)

func init() {
	fMakeCredSet.Var(&fMakeCredHandle, handleFlagName, "")
	fReadPublicSet.Var(&fReadPublicHandle, handleFlagName, "")

	for _, cmd := range commands {
		if cmd.flagSet != nil {
			cmd.flagSet.Usage = cmd.usageFunc
		}
	}
}

// String returns a string representation of the flag value.
func (f *handleFlag) String() string {
	if f == nil {
		return ""
	}

	return fmt.Sprintf("0x%08x", *f)
}

// Sets the value of the flag.
func (f *handleFlag) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return err
	}

	if v > uint64(0xffffffff) {
		return errors.New("out of range")
	}

	*f = handleFlag(v)

	return nil
}

// isFlagPassed checked if the named flag was passed.
func isFlagPassed(set *flag.FlagSet, name string) bool {
	if set == nil {
		return false
	}

	found := false
	set.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

// countFlagsPassed determines how many of the named flags were passed at
// the command line.
func countFlagsPassed(set *flag.FlagSet, names ...string) int {
	var count int

	for _, name := range names {
		if isFlagPassed(set, name) {
			count++
		}
	}

	return count
}

// listifyFlagNames returns a string representation of a comma-separated
// list of flag names, each prepended with the '-' character.
func listifyFlagNames(names ...string) string {
	if len(names) == 0 {
		panic("at least one name must be passed to listifyFlagNames")
	}

	var builder strings.Builder

	for i := range names {
		if i != 0 {
			if i == len(names)-1 {
				builder.WriteString(" or ")
			} else {
				builder.WriteString(", ")
			}
		}
		builder.WriteString("-" + names[i])
	}

	return builder.String()
}

// ensureExactlyOnePassed logs a failure message unless exactly one of the
// named flags was passed at the command line.
func ensureExactlyOnePassed(set *flag.FlagSet, names ...string) {
	if len(names) == 0 {
		panic("at least one name must be passed to ensureExactlyOneOf")
	}

	if countFlagsPassed(set, names...) != 1 {
		if len(names) == 1 {
			log.Fatalf("-%s must be provided", names[0])
		} else {
			log.Fatalf("exactly one of %s must be provided", listifyFlagNames(names...))
		}
	}
}

// ensureAllPassed logs a failure message unless all of the named flags were
// passed at the command line.
func ensureAllPassed(set *flag.FlagSet, names ...string) {
	if len(names) == 0 {
		panic("at least one name must be passed to ensureAllPassed")
	}

	if countFlagsPassed(set, names...) != len(names) {
		if len(names) == 1 {
			log.Fatalf("-%s must be provided", names[0])
		} else {
			log.Fatalf("%s must all be provided", listifyFlagNames(names...))
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

// usageMain outputs a full usage message for the application.
func usageMain() {
	fmt.Printf("usage: %s <command> [options]\n", appName)
	fmt.Println()

	fmt.Printf("%s is a TPM2.0 command line client.\n", appName)
	fmt.Println()

	const fw = 13
	fmt.Println("Commands:")
	fmt.Printf("    %-*s output selected TPM capabilities\n", fw, capsCommand)
	fmt.Printf("    %-*s show this usage information\n", fw, helpCommand)
	fmt.Printf("    %-*s read a TPM object's public area\n", fw, readPublicCommand)
	fmt.Println()

	fmt.Printf("Use \"%s <command> -help\" for more information about a command.\n", appName)
	fmt.Println()
}

// usageCaps outputs usage information for the caps command.
func usageCaps() {
	fmt.Printf("usage: %s %s [options]\n", appName, capsCommand)
	fmt.Println()

	fmt.Printf("The %s command outputs selected TPM capabilities.\n", capsCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s output supported algorithms\n", fw, algsFlagName)
	fmt.Printf("    -%-*s output all capabilities\n", fw, allFlagName)
	fmt.Printf("    -%-*s output active handles\n", fw, handlesFlagName)
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}

// usageMakeCred outputs usage information for the makecred command.
func usageMakeCred() {
	fmt.Printf("usage: %s %s [options]\n", appName, makeCredCommand)
	fmt.Println()

	fmt.Printf("The %s command creates an activation credential.\n", makeCredCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s credential blob output file\n", fw, credOutFlagName+" <path>")
	fmt.Printf("    -%-*s persistent object handle of protecting key\n", fw, handleFlagName+" <integer>")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s input file containing credential (default: stdin)\n", fw, inFlagName+" <path>")
	fmt.Printf("    -%-*s public area input file\n", fw, publicAreaFlagName+" <path>")
	fmt.Printf("    -%-*s encrypted secret output file\n", fw, secretOutFlagName+" <path>")
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}

// usageReadPublic outputs usage information for the readpublic command.
func usageReadPublic() {
	fmt.Printf("usage: %s %s [options]\n", appName, readPublicCommand)
	fmt.Println()

	fmt.Printf("The %s command reads the public area for a TPM object.\n", readPublicCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s persistent object handle\n", fw, handleFlagName+" <integer>")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s input file\n", fw, inFlagName+" <path>")
	fmt.Printf("    -%-*s output file (default: stdout)\n", fw, outFlagName+" <path>")
	fmt.Printf("    -%-*s output public key in PEM format\n", fw, pubOutFlagName)
	fmt.Printf("    -%-*s print the public area in text form\n", fw, textFlagName)
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}
