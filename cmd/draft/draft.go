// Copyright (c) Microsoft Corporation. All rights reserved.
//
// Licensed under the MIT license.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Azure/draft/pkg/draft/draftpath"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	homeEnvVar      = "DRAFT_HOME"
	hostEnvVar      = "HELM_HOST"
	namespaceEnvVar = "TILLER_NAMESPACE"
)

var (
	// flagDebug is a signal that the user wants additional output.
	flagDebug bool
	// draftHome depicts the home directory where all Draft config is stored.
	draftHome string
	//rootCmd is the root command handling `draft`. It's used in other parts of package cmd to add/search the command tree.
	rootCmd *cobra.Command
	// globalConfig is the configuration stored in $DRAFT_HOME/config.toml
	globalConfig DraftConfig
)

var globalUsage = `The application deployment tool for Kubernetes.`

// DraftConfig is the configuration stored in $DRAFT_HOME/config.toml
type DraftConfig map[string]string

func init() {
	rootCmd = newRootCmd(os.Stdout, os.Stdin)
}

func newRootCmd(out io.Writer, in io.Reader) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "draft",
		Short:        globalUsage,
		Long:         globalUsage,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if flagDebug {
				log.SetLevel(log.DebugLevel)
			}
			os.Setenv(homeEnvVar, draftHome)
			globalConfig, err = ReadConfig()
			return
		},
	}
	p := cmd.PersistentFlags()
	p.StringVar(&draftHome, "home", defaultDraftHome(), "location of your Draft config. Overrides $DRAFT_HOME")

	cmd.AddCommand(
		//newConfigCmd(out),
		newCreateCmd(out),
		newHomeCmd(out),
		newInitCmd(out, in),
		//newUpCmd(out),
		//newVersionCmd(out),
		newPluginCmd(out),
		//newConnectCmd(out),
		//newDeleteCmd(out),
		//newLogsCmd(out),
		//newHistoryCmd(out),
		//newPackCmd(out),
	)

	// Find and add plugins
	//loadPlugins(cmd, draftpath.Home(homePath()), out, in)

	return cmd
}

func defaultDraftHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".draft")
}

func homePath() string {
	return os.ExpandEnv(draftHome)
}

func debug(format string, args ...interface{}) {
	if flagDebug {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Printf(format, args...)
	}
}

func validateArgs(args, expectedArgs []string) error {
	if len(args) != len(expectedArgs) {
		return fmt.Errorf("This command needs %v argument(s): %v", len(expectedArgs), expectedArgs)
	}
	return nil
}

// ReadConfig reads in global configuration from $DRAFT_HOME/config.toml
func ReadConfig() (DraftConfig, error) {
	var data DraftConfig
	h := draftpath.Home(draftHome)
	f, err := os.Open(h.Config())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Could not open file %s: %s", h.Config(), err)
	}
	defer f.Close()
	if _, err := toml.DecodeReader(f, &data); err != nil {
		return nil, fmt.Errorf("Could not decode config %s: %s", h.Config(), err)
	}
	return data, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
