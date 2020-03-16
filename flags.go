package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paulgriffiths/pgtpm"
)

// command represents an application command.
type command struct {
	name      string
	flagSet   *flag.FlagSet
	cmdFunc   func() error
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
	activateCommand      = "activate"
	capsCommand          = "caps"
	createCommand        = "create"
	createPrimaryCommand = "createprimary"
	evictCommand         = "evict"
	flushCommand         = "flush"
	helpCommand          = "help"
	makeCredCommand      = "makecred"
	readPublicCommand    = "readpublic"
)

// Flag name constants.
const (
	algsFlagName              = "algorithms"
	allFlagName               = "all"
	credInFlagName            = "credin"
	credOutFlagName           = "credout"
	handleFlagName            = "handle"
	handlesFlagName           = "handles"
	helpFlagName              = "help"
	inFlagName                = "in"
	outFlagName               = "out"
	ownerFlagName             = "owner"
	ownerPasswordFlagName     = "ownerpass"
	parentFlagName            = "parent"
	parentPasswordFlagName    = "parentpass"
	passwordFlagName          = "pass"
	persistentFlagName        = "persistent"
	privOutFlagName           = "privout"
	protectorFlagName         = "protector"
	protectorPasswordFlagName = "protectorpass"
	publicAreaFlagName        = "publicarea"
	pubOutFlagName            = "pubout"
	secretInFlagName          = "secretin"
	secretOutFlagName         = "secretout"
	templateFlagName          = "template"
	textFlagName              = "text"
	tpmFlagName               = "tpm"
)

// commands are the application commands.
var commands = []command{
	{
		name:    helpCommand,
		cmdFunc: usageMain,
	},
	{
		name:      activateCommand,
		flagSet:   fActivateSet,
		cmdFunc:   activateCred,
		usageFunc: usageActivate,
	},
	{
		name:      capsCommand,
		flagSet:   fCapsSet,
		cmdFunc:   outputCaps,
		usageFunc: usageCaps,
	},
	{
		name:      createCommand,
		flagSet:   fCreateSet,
		cmdFunc:   createObject,
		usageFunc: usageCreate,
	},
	{
		name:      createPrimaryCommand,
		flagSet:   fCreatePrimarySet,
		cmdFunc:   createPrimary,
		usageFunc: usageCreatePrimary,
	},
	{
		name:      evictCommand,
		flagSet:   fEvictSet,
		cmdFunc:   evictObject,
		usageFunc: usageEvict,
	},
	{
		name:      flushCommand,
		flagSet:   fFlushSet,
		cmdFunc:   flushContext,
		usageFunc: usageFlush,
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

// activate command flag set.
var (
	fActivateSet               = flag.NewFlagSet(activateCommand, flag.ExitOnError)
	fActivateCredIn            = fActivateSet.String(credInFlagName, "", "")
	fActivateHandle            handleFlag
	fActivatePassword          = fActivateSet.String(passwordFlagName, "", "")
	fActivateProtector         handleFlag
	fActivateProtectorPassword = fActivateSet.String(protectorPasswordFlagName, "", "")
	fActivateHelp              = fActivateSet.Bool(helpFlagName, false, "")
	fActivateSecretIn          = fActivateSet.String(secretInFlagName, "", "")
	fActivateTPM               = fActivateSet.String(tpmFlagName, defaultTPMDevice, "")
)

// caps command flag set.
var (
	fCapsSet     = flag.NewFlagSet(capsCommand, flag.ExitOnError)
	fCapsAlgs    = fCapsSet.Bool(algsFlagName, false, "")
	fCapsAll     = fCapsSet.Bool(allFlagName, false, "")
	fCapsHandles = fCapsSet.Bool(handlesFlagName, false, "")
	fCapsHelp    = fCapsSet.Bool(helpFlagName, false, "")
	fCapsTPM     = fCapsSet.String(tpmFlagName, defaultTPMDevice, "")
)

// createprimary command flag set.
var (
	fCreatePrimarySet           = flag.NewFlagSet(createPrimaryCommand, flag.ExitOnError)
	fCreatePrimaryPersistent    handleFlag
	fCreatePrimaryHelp          = fCreatePrimarySet.Bool(helpFlagName, false, "")
	fCreatePrimaryOwnerPassword = fCreatePrimarySet.String(ownerPasswordFlagName, "", "")
	fCreatePrimaryPassword      = fCreatePrimarySet.String(passwordFlagName, "", "")
	fCreatePrimaryTemplate      = fCreatePrimarySet.String(templateFlagName, "", "")
	fCreatePrimaryTPM           = fCreatePrimarySet.String(tpmFlagName, defaultTPMDevice, "")
)

// create command flag set.
var (
	fCreateSet            = flag.NewFlagSet(createCommand, flag.ExitOnError)
	fCreatePersistent     handleFlag
	fCreateHelp           = fCreateSet.Bool(helpFlagName, false, "")
	fCreateOwnerPassword  = fCreateSet.String(ownerPasswordFlagName, "", "")
	fCreateParent         handleFlag
	fCreateParentPassword = fCreateSet.String(parentPasswordFlagName, "", "")
	fCreatePassword       = fCreateSet.String(passwordFlagName, "", "")
	fCreatePublicOut      = fCreateSet.String(pubOutFlagName, "", "")
	fCreatePrivateOut     = fCreateSet.String(privOutFlagName, "", "")
	fCreateTemplate       = fCreateSet.String(templateFlagName, "", "")
	fCreateTPM            = fCreateSet.String(tpmFlagName, defaultTPMDevice, "")
)

// evict command flag set.
var (
	fEvictSet           = flag.NewFlagSet(evictCommand, flag.ExitOnError)
	fEvictHandle        handleFlag
	fEvictHelp          = fEvictSet.Bool(helpFlagName, false, "")
	fEvictOwnerPassword = fEvictSet.String(ownerPasswordFlagName, "", "")
	fEvictTPM           = fEvictSet.String(tpmFlagName, defaultTPMDevice, "")
)

// flush command flag set.
var (
	fFlushSet    = flag.NewFlagSet(flushCommand, flag.ExitOnError)
	fFlushHandle handleFlag
	fFlushHelp   = fFlushSet.Bool(helpFlagName, false, "")
	fFlushTPM    = fFlushSet.String(tpmFlagName, defaultTPMDevice, "")
)

// makecred command flag set.
var (
	fMakeCredSet        = flag.NewFlagSet(makeCredCommand, flag.ExitOnError)
	fMakeCredHandle     handleFlag
	fMakeCredHelp       = fMakeCredSet.Bool(helpFlagName, false, "")
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
	fReadPublicHelp   = fReadPublicSet.Bool(helpFlagName, false, "")
	fReadPublicIn     = fReadPublicSet.String(inFlagName, "", "")
	fReadPublicOut    = fReadPublicSet.String(outFlagName, "", "")
	fReadPublicPubOut = fReadPublicSet.Bool(pubOutFlagName, false, "")
	fReadPublicText   = fReadPublicSet.Bool(textFlagName, false, "")
	fReadPublicTPM    = fReadPublicSet.String(tpmFlagName, defaultTPMDevice, "")
)

func init() {
	fActivateSet.Var(&fActivateHandle, handleFlagName, "")
	fActivateSet.Var(&fActivateProtector, protectorFlagName, "")
	fCreateSet.Var(&fCreateParent, parentFlagName, "")
	fCreateSet.Var(&fCreatePersistent, persistentFlagName, "")
	fCreatePrimarySet.Var(&fCreatePrimaryPersistent, persistentFlagName, "")
	fEvictSet.Var(&fEvictHandle, handleFlagName, "")
	fFlushSet.Var(&fFlushHandle, handleFlagName, "")
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
func ensureExactlyOnePassed(set *flag.FlagSet, names ...string) error {
	if len(names) == 0 {
		panic("at least one name must be passed to ensureExactlyOneOf")
	}

	if countFlagsPassed(set, names...) != 1 {
		if len(names) == 1 {
			return fmt.Errorf("-%s must be provided", names[0])
		}

		return fmt.Errorf("exactly one of %s must be provided", listifyFlagNames(names...))
	}

	return nil
}

// ensureAllPassed logs a failure message unless all of the named flags were
// passed at the command line.
func ensureAllPassed(set *flag.FlagSet, names ...string) error {
	if len(names) == 0 {
		panic("at least one name must be passed to ensureAllPassed")
	}

	if countFlagsPassed(set, names...) != len(names) {
		if len(names) == 1 {
			return fmt.Errorf("-%s must be provided", names[0])
		}

		return fmt.Errorf("%s must all be provided", listifyFlagNames(names...))
	}

	return nil
}

// ensureAllOrNonePassed logs a failure message unless all or none of the named
// flags were passed at the command line.
func ensureAllOrNonePassed(set *flag.FlagSet, names ...string) error {
	if len(names) < 0 {
		panic("at least two names must be passed to ensureAllOrNonePassed")
	}

	if c := countFlagsPassed(set, names...); c != 0 && c != len(names) {
		return fmt.Errorf("all or none of %s must be provided", listifyFlagNames(names...))
	}

	return nil
}

// usageError outputs a brief usage message to standard error and exits with
// status code 1.
func usageError() {
	fmt.Fprintf(os.Stderr, "usage: %s <command> [options]\n\n", appName)
	fmt.Printf("Use \"%s help\" for a list of commands.\n", appName)
	os.Exit(1)
}

// usageMain outputs a full usage message for the application.
func usageMain() error {
	fmt.Printf("usage: %s <command> [options]\n", appName)
	fmt.Println()

	fmt.Printf("%s is a TPM2.0 command line client.\n", appName)
	fmt.Println()

	const fw = 16
	fmt.Println("Commands:")
	fmt.Printf("    %-*s activate a credential\n", fw, activateCommand)
	fmt.Printf("    %-*s output selected TPM capabilities\n", fw, capsCommand)
	fmt.Printf("    %-*s create an object\n", fw, createCommand)
	fmt.Printf("    %-*s create a primary object\n", fw, createPrimaryCommand)
	fmt.Printf("    %-*s evict a persistent object\n", fw, evictCommand)
	fmt.Printf("    %-*s flush a transient object\n", fw, flushCommand)
	fmt.Printf("    %-*s show this usage information\n", fw, helpCommand)
	fmt.Printf("    %-*s make an activation credential\n", fw, makeCredCommand)
	fmt.Printf("    %-*s read a TPM object's public area\n", fw, readPublicCommand)
	fmt.Println()

	fmt.Printf("Use \"%s <command> -help\" for more information about a command.\n", appName)
	fmt.Println()

	return nil
}

// usageActivate outputs usage information for the activate command.
func usageActivate() {
	fmt.Printf("usage: %s %s [options]\n", appName, activateCommand)
	fmt.Println()

	fmt.Printf("The %s command activates a credential.\n", activateCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s credential blob input file\n", fw, credInFlagName+" <path>")
	fmt.Printf("    -%-*s persistent object handle of key\n", fw, handleFlagName+" <integer>")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s key password\n", fw, passwordFlagName+" <string>")
	fmt.Printf("    -%-*s persistent object handle of protecting key\n", fw, protectorFlagName+" <integer>")
	fmt.Printf("    -%-*s protecting key password\n", fw, protectorPasswordFlagName+" <string>")
	fmt.Printf("    -%-*s encrypted secret input file\n", fw, secretInFlagName+" <path>")
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
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

// usageCreate outputs usage information for the create command.
func usageCreate() {
	fmt.Printf("usage: %s %s [options]\n", appName, createCommand)
	fmt.Println()

	fmt.Printf("The %s command creates an object.\n", createCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s owner password\n", fw, ownerPasswordFlagName+" <string>")
	fmt.Printf("    -%-*s persistent handle of parent object\n", fw, parentFlagName+" <integer>")
	fmt.Printf("    -%-*s parent password\n", fw, parentPasswordFlagName+" <string>")
	fmt.Printf("    -%-*s object password\n", fw, passwordFlagName+" <string>")
	fmt.Printf("    -%-*s persistent object handle\n", fw, persistentFlagName+" <integer>")
	fmt.Printf("    -%-*s public area output file\n", fw, pubOutFlagName+" <path>")
	fmt.Printf("    -%-*s private area output file\n", fw, privOutFlagName+" <path>")
	fmt.Printf("    -%-*s template\n", fw, templateFlagName+" <path>")
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}

// usageCreatePrimary outputs usage information for the createprimary command.
func usageCreatePrimary() {
	fmt.Printf("usage: %s %s [options]\n", appName, createPrimaryCommand)
	fmt.Println()

	fmt.Printf("The %s command creates a primary object.\n", createPrimaryCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s owner password\n", fw, ownerPasswordFlagName+" <string>")
	fmt.Printf("    -%-*s object password\n", fw, passwordFlagName+" <string>")
	fmt.Printf("    -%-*s persistent object handle\n", fw, persistentFlagName+" <integer>")
	fmt.Printf("    -%-*s template\n", fw, templateFlagName+" <path>")
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}

// usageEvict outputs usage information for the evict command.
func usageEvict() {
	fmt.Printf("usage: %s %s [options]\n", appName, evictCommand)
	fmt.Println()

	fmt.Printf("The %s command evicts a persistent object.\n", evictCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s persistent object handle\n", fw, handleFlagName+" <integer>")
	fmt.Printf("    -%-*s output this usage information\n", fw, helpFlagName)
	fmt.Printf("    -%-*s owner password\n", fw, ownerPasswordFlagName+" <string>")
	fmt.Printf("    -%-*s TPM device (default: %s)\n", fw, tpmFlagName+" <path>|<hostname:port>", defaultTPMDevice)
	fmt.Println()
}

// usageFlush outputs usage information for the flush command.
func usageFlush() {
	fmt.Printf("usage: %s %s [options]\n", appName, flushCommand)
	fmt.Println()

	fmt.Printf("The %s command flushes a transient object.\n", flushCommand)
	fmt.Println()

	const fw = 29
	fmt.Println("Options:")
	fmt.Printf("    -%-*s transient object handle\n", fw, handleFlagName+" <integer>")
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
