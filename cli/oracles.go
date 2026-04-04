package cli

import (
	"fmt"
	"path/filepath"
	"soloterm/config"
	"soloterm/database"
	"soloterm/domain/oracle"
	"soloterm/shared/dirs"

	"github.com/spf13/cobra"
)

var oraclesCmd = &cobra.Command{
	Use:   "oracles",
	Short: "List available oracle tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir, err := dirs.ConfigDir()
		if err != nil {
			return err
		}
		if configFlag != "" {
			configDir = filepath.Dir(configFlag)
		}

		dataDir, err := dirs.DataDir()
		if err != nil {
			return err
		}

		var cfg config.Config
		loadedCfg, err := cfg.Load(configDir)
		if err != nil {
			return err
		}

		dbPath := database.ResolveDBPath(loadedCfg.DatabaseDir, dataDir)
		if dbFlag != "" {
			dbPath = dbFlag
		}
		db, err := database.Setup(dbPath)
		if err != nil {
			return err
		}
		defer db.Connection.Close()

		oracles, err := oracle.NewService(oracle.NewRepository(db)).GetAll()
		if err != nil {
			return err
		}

		for _, o := range oracles {
			fmt.Printf("%s/%s\n", o.Category, o.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(oraclesCmd)
}
