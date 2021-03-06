package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use: "hashi-up",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		SilenceErrors: true,
	}

	var certificate = &cobra.Command{
		Use: "cert",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var consul = &cobra.Command{
		Use: "consul",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var nomad = &cobra.Command{
		Use: "nomad",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var vault = &cobra.Command{
		Use: "vault",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	certificate.AddCommand(CreateCertificateCommand())

	nomad.AddCommand(InstallNomadCommand())
	nomad.AddCommand(UninstallNomadCommand())
	nomad.AddCommand(GetCommand("nomad"))

	consul.AddCommand(InstallConsulCommand())
	consul.AddCommand(UninstallConsulCommand())
	consul.AddCommand(GetCommand("consul"))

	vault.AddCommand(InstallVaultCommand())
	vault.AddCommand(UninstallVaultCommand())
	vault.AddCommand(GetCommand("vault"))

	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(CompletionCommand())
	rootCmd.AddCommand(certificate)
	rootCmd.AddCommand(nomad)
	rootCmd.AddCommand(consul)
	rootCmd.AddCommand(vault)

	return rootCmd.Execute()
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}

func info(message string) {
	fmt.Println("[INFO] " + message)
}
