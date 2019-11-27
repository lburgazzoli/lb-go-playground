package main

import (
	"errors"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindPFlagsHierarchy(cmd *cobra.Command) error {
	for _, c := range cmd.Commands() {
		if err := bindPFlags(c); err != nil {
			return err
		}

		if err := bindPFlagsHierarchy(c); err != nil {
			return err
		}
	}

	return nil
}

func bindPFlags(cmd *cobra.Command) error {
	prefix := cmd.Name()

	for current := cmd.Parent(); current != nil; current = current.Parent() {
		name := current.Name()
		name = strings.ReplaceAll(name, "_", "-")
		name = strings.ReplaceAll(name, ".", "-")
		prefix = name + "." + prefix
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		name := flag.Name
		name = strings.ReplaceAll(name, "_", "-")
		name = strings.ReplaceAll(name, ".", "-")

		if err := viper.BindPFlag(prefix+"."+name, flag); err != nil {
			log.Fatalf("error binding flag %s with prefix %s to viper", flag.Name, prefix)
		}
	})

	return nil
}

type MainOptions struct {
}

type SubOptions struct {
	Info string `mapstructure:"info"`
}

func main() {
	mopt := MainOptions{}
	sopt := SubOptions{}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", ".", "-", "."))

	mainCmd := &cobra.Command{
		Use:   "main",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.UnmarshalKey("main", &mopt); err != nil {
				return nil
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("mopt: %+v", mopt)
			return nil
		},
	}

	subCmd1 := &cobra.Command{
		Use:   "sub1",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			v := viper.Sub("main.sub1")
			if v == nil {
				return errors.New("no main.sub1")
			}

			if err := v.Unmarshal(&sopt); err != nil {
				return nil
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("sopt: %+v", sopt)
			return nil
		},
	}

	subCmd1.Flags().StringP("info", "i", "", "shows info")
	mainCmd.AddCommand(subCmd1)

	bindPFlagsHierarchy(mainCmd)

	mainCmd.Execute()
}
