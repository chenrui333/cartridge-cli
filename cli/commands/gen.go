package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/common"
)

var (
	completionsDirName = "completion"

	bashCompFilePath string
	zshCompFilePath  string

	defaultBashCompFilePath string
	defaultZshCompFilePath  string

	skipBash bool
	skipZsh  bool
)

/*
 * `cartridge gen` command is used to generate shell
 * autocompletions for Bash and Zsh.
 *
 * Autocompletion is generated by cobra, see
 * https://github.com/spf13/cobra/blob/master/shell_completions.md.
 *
 * Bash completion is delivered with the RPM and DEB packages
 * (see .goreleaser.yml).
 *
 * On installation from `brew` both Bash and Zsh completions are installed
 * automatically.
 *
 * It can be used to generate completion for manual installation.
 */

func init() {
	defaultBashCompFilePath = filepath.Join(completionsDirName, "bash", rootCmd.Name())
	defaultZshCompFilePath = filepath.Join(completionsDirName, "zsh", fmt.Sprintf("_%s", rootCmd.Name()))

	var genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate some useful things",
		Args:  cobra.MaximumNArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			cutFlagsDescription(rootCmd)
		},
	}

	rootCmd.AddCommand(genCmd)

	var genCompletionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generate shell autocompletion scripts",
		Args:  cobra.MaximumNArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			cutFlagsDescription(rootCmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := genCompletion(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	genCompletionCmd.Flags().StringVar(&bashCompFilePath, "bash", defaultBashCompFilePath, "Bash completion file path")
	genCompletionCmd.Flags().StringVar(&zshCompFilePath, "zsh", defaultZshCompFilePath, "Zsh completion file path")

	genCompletionCmd.Flags().BoolVar(&skipBash, "skip-bash", false, "Do not generate bash completion")
	genCompletionCmd.Flags().BoolVar(&skipZsh, "skip-zsh", false, "Do not generate zsh completion")

	genSubCommands := []*cobra.Command{
		genCompletionCmd,
	}

	for _, cmd := range genSubCommands {
		genCmd.AddCommand(cmd)
		configureFlags(cmd)
	}
}

// cutFlagsDescription cuts command usage on first '\n'
// it's needed to make zsh comletion for flags prettier
func cutFlagsDescription(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.VisitAll(func(f *pflag.Flag) {
		f.Usage = strings.SplitN(f.Usage, "\n", 2)[0]
	})

	for _, subCmd := range cmd.Commands() {
		cutFlagsDescription(subCmd)
	}
}

func genCompletion(cmd *cobra.Command, args []string) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Cailed to get current directory path: %s", err)
	}

	bashCompFilePath := filepath.Join(curDir, bashCompFilePath)
	zshCompFilePath := filepath.Join(curDir, zshCompFilePath)

	// create directories
	bashCompFileDir := filepath.Dir(bashCompFilePath)
	if err := os.MkdirAll(bashCompFileDir, 0755); err != nil {
		return fmt.Errorf("Failed to create bash completion directory: %s", err)
	}

	zshCompFileDir := filepath.Dir(zshCompFilePath)
	if err := os.MkdirAll(zshCompFileDir, 0755); err != nil {
		return fmt.Errorf("Failed to create zsh completion directory: %s", err)
	}

	// gen completions
	if !skipBash {
		if err := os.RemoveAll(bashCompFilePath); err != nil {
			return fmt.Errorf("Failed to remove existent bash completion: %s", err)
		}

		if err := cmd.Root().GenBashCompletionFile(bashCompFilePath); err != nil {
			return fmt.Errorf("Failed to generate bash completion: %s", err)
		}

		// bash: remove flags duplicates (e.g. '--name', '--name=')
		twoWordsFlagRgx := regexp.MustCompile(`(two_word_flags\+=\("--[\w-]+"\))`)
		if err := common.ReplaceFileLinesByRe(bashCompFilePath, twoWordsFlagRgx, "# $1"); err != nil {
			return fmt.Errorf("Failed to comment two words flags in the bash completion: %s", err)
		}
	}

	if !skipZsh {
		if err := os.RemoveAll(zshCompFilePath); err != nil {
			return fmt.Errorf("Failed to remove existent zsh completion: %s", err)
		}

		if err := cmd.Root().GenZshCompletionFile(zshCompFilePath); err != nil {
			return fmt.Errorf("Failed to generate zsh completion: %s", err)
		}
	}

	return nil
}
