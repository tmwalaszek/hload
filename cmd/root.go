package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"

	"github.com/tmwalaszek/hload/cmd/cliio"
	"github.com/tmwalaszek/hload/cmd/common"
	"github.com/tmwalaszek/hload/cmd/loader"
	"github.com/tmwalaszek/hload/cmd/tags"
	"github.com/tmwalaszek/hload/cmd/template"
	"github.com/tmwalaszek/hload/cmd/version"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	dbFile         string
	profileFile    string
	profileType    string
	renderTemplate string

	pprofEnabled bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "HLoad",
	Short: "HLoad CLI",
	Long:  "HLoad CLI is a HTTP benchmarking tool written in Golang.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !pprofEnabled {
			return
		}

		switch profileType {
		case "cpu":
			f, err := os.Create(profileFile)
			if err != nil {
				log.Fatal(err)
			}

			err = pprof.StartCPUProfile(f)
			if err != nil {
				log.Fatal(err)
			}
		case "block":
			runtime.SetBlockProfileRate(1)
		case "mutex":
			runtime.SetMutexProfileFraction(1)
		default:
			p := pprof.Lookup(profileType)
			if p == nil {
				log.Fatalf("unknown profile '%s'", profileType)
			}
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			_ = writeProfile()
		}()
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := writeProfile()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func writeProfile() error {
	if !pprofEnabled {
		return nil
	}

	switch profileType {
	case "cpu":
		pprof.StopCPUProfile()
	case "heap":
		runtime.GC()
		fallthrough
	case "block":
		fallthrough
	case "mutex":
		f, err := os.Create(profileFile)
		if err != nil {
			return err
		}

		err = pprof.Lookup(profileType).WriteTo(f, 0)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown profile '%s'", profileType)
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	defaultConfFile := path.Join(home, ".hload", "hload.yaml")
	defaultDbFile := path.Join(home, ".hload", "store.db")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfFile, "config file")
	rootCmd.PersistentFlags().StringVar(&dbFile, "db", defaultDbFile, "Inventory file location")
	rootCmd.PersistentFlags().StringVar(&renderTemplate, "template", "default", "The loader/summary output renderTemplate")
	rootCmd.PersistentFlags().BoolVar(&pprofEnabled, "enable-profile", false, "Enable profiling")
	rootCmd.PersistentFlags().StringVar(&profileFile, "profile-file", "profile.pprof", "Profile file name")
	rootCmd.PersistentFlags().StringVar(&profileType, "profile-type", "cpu", "Profile type (cpu, mem, block, mutex)")

	err = viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("template", rootCmd.PersistentFlags().Lookup("template"))
	if err != nil {
		log.Fatal(err)
	}

	err = common.CreateDBDirectory(viper.GetString("db"))
	if err != nil {
		log.Fatal(err)
	}

	cliIO := cliio.NewStdIO()

	rootCmd.AddCommand(loader.NewLoaderCmd(cliIO))
	rootCmd.AddCommand(tags.NewTagsCmd(cliIO))
	rootCmd.AddCommand(version.NewVersionCmd(cliIO))
	rootCmd.AddCommand(template.NewTemplateCmd(cliIO))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("HLOAD")
	viper.AutomaticEnv() // read in environment variables that match

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".example" (without extension).
		if viper.GetString("CONFIG") != "" {
			viper.SetConfigFile(viper.GetString("CONFIG"))
		}
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return
		}

		log.Fatalf("Could not load hload config file: %v", err)
	}
}
