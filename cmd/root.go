package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Version information - these will be set during the build process
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const logo = "  \033[31m" + `
   _        _______        _______  _______  _______  _______ 
  ( (    /|(  ____ \      (  ____ \(  ____ )(  ____ \(  ____ \
  |  \  ( || (    \/      | (    \/| (    )|| (    \/| (    \/
  |   \ | || |      _____ | (_____ | (____)|| (__    | |      
  | (\ \) || | ____(_____)(_____  )|  _____)|  __)   | |      
  | | \   || | \_  )            ) || (      | (      | |      
  | )  \  || (___) |      /\____) || )      | (____/\| (____/\
  |/    )_)(_______)      \_______)|/       (_______/(_______/  
  ` + "\033[0m"

var rootCmd = &cobra.Command{
	Use:   "ng-spec",
	Short: "A simple CLI tool to generate minimal integration test boilerplate for Angular components",
	Long:  `Generate minimal integration test boilerplate for Angular components using the Angular Testing Library.`,
	Example: `
	ng-spec
	ng-spec app
	ng-spec /path/to/component
	`,
	Run: func(cmd *cobra.Command, args []string) {
		var component string

		// Only print component if -v or --version not passed
		if !cmd.Flags().Changed("version") {
			fmt.Println(component)
		}

		if len(args) > 0 {
			component = args[0]
		}

		generateComponentTest(component)
	},
	Version: version,
}

// init function to set version information from build info when available
func init() {
	// Try to get version info from Go module information
	if info, ok := debug.ReadBuildInfo(); ok {
		// Look for version in main module
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}

		// Look for VCS info in settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) > 7 {
					commit = setting.Value[:7]
				} else if setting.Value != "" {
					commit = setting.Value
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					date = t.Format("02/01/2006")
				}
			}
		}
	}
}

func init() {
	versionTemplate := logo + `
  Version                      %s
  Commit                       %s
  Release date	               %s
`
	versionTemplate = fmt.Sprintf(versionTemplate, version, commit, date)

	rootCmd.SetVersionTemplate(versionTemplate)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
