package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jsiebens/hashi-up/pkg/config"
	"github.com/jsiebens/hashi-up/pkg/operator"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thanhpk/randstr"
)

func InstallBoundaryCommand() *cobra.Command {

	var skipEnable bool
	var skipStart bool
	var binary string
	var version string

	var initDatabase bool
	var configFile string
	var files []string

	var flags = config.BoundaryConfig{}

	var command = &cobra.Command{
		Use:          "install",
		SilenceUsage: true,
	}

	command.Flags().BoolVar(&skipEnable, "skip-enable", false, "If set to true will not enable or start Boundary service")
	command.Flags().BoolVar(&skipStart, "skip-start", false, "If set to true will not start Boundary service")
	command.Flags().StringVarP(&binary, "package", "p", "", "Upload and use this Boundary package instead of downloading")
	command.Flags().StringVarP(&version, "version", "v", "", "Version of Boundary to install")

	command.Flags().BoolVarP(&initDatabase, "init-database", "d", false, "Initialize the Boundary database")
	command.Flags().StringVarP(&configFile, "config-file", "c", "", "Custom Boundary configuration file to upload")
	command.Flags().StringArrayVarP(&files, "file", "f", []string{}, "Additional files, e.g. certificates, to upload")

	command.Flags().StringVar(&flags.ControllerName, "controller-name", "", "Boundary: specifies a unique name of this controller within the Boundary controller cluster.")
	command.Flags().StringVar(&flags.WorkerName, "worker-name", "", "Boundary: specifies a unique name of this worker within the Boundary worker cluster.")
	command.Flags().StringVar(&flags.DatabaseURL, "db-url", "", "Boundary: configures the URL for connecting to Postgres")
	command.Flags().StringVar(&flags.RootKey, "root-key", "", "Boundary: a KEK (Key Encrypting Key) for the scope-specific KEKs (also referred to as the scope's root key).")
	command.Flags().StringVar(&flags.WorkerAuthKey, "worker-auth-key", "", "Boundary: KMS key shared by the Controller and Worker in order to authenticate a Worker to the Controller.")
	command.Flags().StringVar(&flags.RecoveryKey, "recovery-key", "", "Boundary: KMS key is used for rescue/recovery operations that can be used by a client to authenticate almost any operation within Boundary.")
	command.Flags().StringVar(&flags.ApiAddress, "api-addr", "0.0.0.0", "Boundary: address for the API listener")
	command.Flags().StringVar(&flags.ClusterAddress, "cluster-addr", "127.0.0.1", "Boundary: address for the Cluster listener")
	command.Flags().StringVar(&flags.ProxyAddress, "proxy-addr", "0.0.0.0", "Boundary: address for the Proxy listener")
	command.Flags().StringVar(&flags.PublicClusterAddress, "public-cluster-addr", "", "Boundary: specifies the public host or IP address (and optionally port) at which the controller can be reached by workers.")
	command.Flags().StringVar(&flags.PublicAddress, "public-addr", "", "Boundary: specifies the public host or IP address (and optionally port) at which the worker can be reached by clients for proxying.")
	command.Flags().StringArrayVar(&flags.Controllers, "controller", []string{"127.0.0.1"}, "Boundary: a list of hosts/IP addresses and optionally ports for reaching controllers.")

	command.RunE = func(command *cobra.Command, args []string) error {
		if !runLocal && len(sshTargetAddr) == 0 {
			return fmt.Errorf("required ssh-target-addr flag is missing")
		}

		ignoreConfigFlags := len(configFile) != 0

		var generatedConfig string

		if !ignoreConfigFlags {
			generatedConfig = flags.GenerateConfigFile()
		}

		if len(binary) == 0 && len(version) == 0 {
			latest, err := config.GetLatestVersion("boundary")

			if err != nil {
				return errors.Wrapf(err, "unable to get latest version number, define a version manually with the --version flag")
			}

			version = latest
		}

		callback := func(op operator.CommandOperator) error {
			dir := "/tmp/boundary-installation." + randstr.String(6)

			defer op.Execute("rm -rf " + dir)

			_, err := op.Execute("mkdir -p " + dir + "/config")
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			if len(binary) != 0 {
				info("Uploading Boundary package...")
				err = op.UploadFile(binary, dir+"/boundary.zip", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload Boundary package: %s", err)
				}
			}

			if !ignoreConfigFlags {
				info("Uploading generated Boundary configuration...")
				err = op.Upload(strings.NewReader(generatedConfig), dir+"/config/boundary.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload boundary configuration: %s", err)
				}
			} else {
				info(fmt.Sprintf("Uploading %s as boundary.hcl...", configFile))
				err = op.UploadFile(expandPath(configFile), dir+"/config/boundary.hcl", "0640")
				if err != nil {
					return fmt.Errorf("error received during upload boundary configuration: %s", err)
				}
			}

			for _, s := range files {
				if len(s) != 0 {
					info(fmt.Sprintf("Uploading %s...", s))
					_, filename := filepath.Split(expandPath(s))
					err = op.UploadFile(expandPath(s), dir+"/config/"+filename, "0640")
					if err != nil {
						return fmt.Errorf("error received during upload file: %s", err)
					}
				}
			}

			installScript, err := pkger.Open("/scripts/install_boundary.sh")

			if err != nil {
				return err
			}

			defer installScript.Close()

			err = op.Upload(installScript, dir+"/install.sh", "0755")
			if err != nil {
				return fmt.Errorf("error received during upload install script: %s", err)
			}

			info("Installing Boundary...")
			_, err = op.Execute(fmt.Sprintf("cat %s/install.sh | TMP_DIR='%s' INIT_DATABASE='%t' BOUNDARY_VERSION='%s' SKIP_ENABLE='%t' SKIP_START='%t' sh -\n", dir, dir, initDatabase, version, skipEnable, skipStart))
			if err != nil {
				return fmt.Errorf("error received during installation: %s", err)
			}

			return nil
		}

		if runLocal {
			return operator.ExecuteLocal(callback)
		} else {
			return operator.ExecuteRemote(sshTargetAddr, sshTargetUser, sshTargetKey, callback)
		}
	}

	return command
}
