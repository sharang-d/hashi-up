package cmd

import (
	"fmt"

	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/markbates/pkger"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
)

func UninstallVaultCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:          "uninstall",
		SilenceUsage: true,
	}

	var target = Target{}
	target.prepareCommand(command)

	command.RunE = func(command *cobra.Command, args []string) error {
		if !target.Local && len(target.Addr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}
		
		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/hashi-up." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			installScript, err := pkger.Open("/scripts/uninstall_vault.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/uninstall.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Uninstalling Vault ...")
			_, err = op.Execute(fmt.Sprintf("cat %s/uninstall.sh | sh -\n", dir))
			if err != nil {
				return fmt.Errorf("error received during uninstallation: %s", err)
			}

			info("Done.")

			return nil
		}

		if target.Local {
			return operator.ExecuteLocal(callback)
		} else {
			return operator.ExecuteRemote(target.Addr, target.User, target.Key, callback)
		}
	}

	return command
}
