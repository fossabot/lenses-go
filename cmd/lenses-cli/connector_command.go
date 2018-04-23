package main

import (
	"fmt"
	"strings"

	"github.com/landoop/lenses-go"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newConnectorsCommand())
	rootCmd.AddCommand(newConnectorGroupCommand())
}

func newConnectorsCommand() *cobra.Command {
	var (
		clusterName string

		namesOnly bool // if true then print only the connector names and not the details as json.
	)

	root := &cobra.Command{
		Use:              "connectors",
		Short:            "List of active connectors' names",
		Aliases:          []string{"connect"},
		Example:          exampleString(`connectors or connectors --clusterName="cluster_name" or --clusterName="*"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			connectorNames := make(map[string][]string) // clusterName:[] connectors names.

			if clusterName == "*" {
				// if * then no clusterName given,
				// fetch the connectors from all known clusters and print them.
				clusters, err := client.GetConnectClusters()
				if err != nil {
					return err
				}

				for _, cluster := range clusters {
					clusterConnectorsNames, err := client.GetConnectors(cluster.Name)
					if err != nil {
						return err
					}
					connectorNames[cluster.Name] = append(connectorNames[cluster.Name], clusterConnectorsNames...)
				}
			} else {
				names, err := client.GetConnectors(clusterName)
				if err != nil {
					return err
				}

				connectorNames[clusterName] = names
			}

			if namesOnly {
				var names []string
				for _, cNames := range connectorNames {
					names = append(names, cNames...)
				}

				return printJSON(cmd, outlineStringResults("name", names))
			}

			connectors := make(map[string][]lenses.Connector, len(connectorNames))

			// else print the entire info.
			for cluster, names := range connectorNames {
				for _, name := range names {
					connector, err := client.GetConnector(cluster, name)
					if err != nil {
						fmt.Fprintf(cmd.OutOrStderr(), "get connector error: %v\n", err)
						continue
					}

					connectors[cluster] = append(connectors[cluster], connector)
				}
			}

			if err := printJSON(cmd, connectors); err != nil {
				return err
			}

			return nil
		},
	}

	root.Flags().BoolVar(&namesOnly, "names", false, `--names`)
	root.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName`)

	canPrintJSON(root)

	// plugins subcommand.
	root.AddCommand(newGetConnectorsPluginsCommand())

	// clusters subcommand.
	root.AddCommand(newGetConnectorsClustersCommand())

	return root
}

func newGetConnectorsPluginsCommand() *cobra.Command {
	var clusterName string

	cmd := &cobra.Command{
		Use:           "plugins",
		Short:         "List of available connectors' plugins",
		Example:       exampleString(`connectors plugins --clusterName="cluster_name"`),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var plugins []lenses.ConnectorPlugin

			if clusterName == "*" {
				// if * then no clusterName given, fetch the plugins from all known clusters and print them.
				clusters, err := client.GetConnectClusters()
				if err != nil {
					return err
				}

				for _, cluster := range clusters {
					clusterPlugins, err := client.GetConnectorPlugins(cluster.Name)
					if err != nil {
						return err
					}
					plugins = append(plugins, clusterPlugins...)
				}
			} else {
				var err error
				plugins, err = client.GetConnectorPlugins(clusterName)
				if err != nil {
					return err
				}
			}

			for _, p := range plugins {
				if p.Version == "null" || p.Version == "" {
					p.Version = "X.X.X"
				}
			}

			return printJSON(cmd, plugins)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName`)

	canPrintJSON(cmd)

	return cmd
}

func newGetConnectorsClustersCommand() *cobra.Command {
	var (
		namesOnly bool
		noNewLine bool // matters when namesOnly is true.
	)

	cmd := &cobra.Command{
		Use:           "clusters",
		Short:         "List of available connectors' clusters",
		Example:       exampleString(`connectors clusters`),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := client.GetConnectClusters()
			if err != nil {
				return err
			}

			if namesOnly {
				var b strings.Builder

				for i, cl := range clusters {
					b.WriteString(fmt.Sprintf("%s", cl.Name))
					if !noNewLine && len(clusters)-1 != i {
						// add new line if enabled and not last, note that we use the fmt.Println below
						// even if newLine is disabled (for unix terminals mostly).
						b.WriteString("\n")
					}
				}

				_, err = fmt.Fprintln(cmd.OutOrStdout(), b.String())
				return err
			}

			return printJSON(cmd, clusters)
		},
	}

	cmd.Flags().BoolVar(&namesOnly, "names", false, `--names`)
	cmd.Flags().BoolVar(&noNewLine, "no-newline", false, "--no-newline Disables line breakers between names, if --names is enabled, defaults to false")

	return cmd
}

func newConnectorGroupCommand() *cobra.Command {
	var clusterName, name string

	root := &cobra.Command{
		Use:              "connector",
		Short:            "Get information about a particular connector based on its name",
		Example:          exampleString(`connector --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			connector, err := client.GetConnector(clusterName, name)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("connector '%s:%s' does not exist", clusterName, name)
				return err
			}
			return printJSON(cmd, connector)
		},
	}

	canPrintJSON(root)

	root.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	root.Flags().StringVar(&name, "name", "", `--name="connector_name"`)

	// subcommands.
	root.AddCommand(newConnectorCreateCommand())
	root.AddCommand(newConnectorUpdateCommand())
	root.AddCommand(newConnectorGetConfigCommand())
	root.AddCommand(newConnectorGetStatusCommand())
	root.AddCommand(newConnectorPauseCommand())
	root.AddCommand(newConnectorResumeCommand())
	root.AddCommand(newConnectorRestartCommand())
	root.AddCommand(newConnectorGetTasksCommand())
	root.AddCommand(newConnectorDeleteCommand())
	// connector.task subcommands.
	root.AddCommand(newConnectorTaskGroupCommand())

	return root
}

func newConnectorCreateCommand() *cobra.Command {
	var clusterName, name, configRaw string

	cmd := &cobra.Command{
		Use:              "create",
		Short:            "Create a new connector",
		Example:          exampleString(`connector create --clusterName="cluster_name" --name="connector_name" --config="{\"key\": \"value\"}" or connector create ./connector.yml`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			connector := lenses.CreateUpdateConnectorPayload{
				ClusterAlias: clusterName,
				Name:         name,
				Config:       make(lenses.ConnectorConfig),
			}

			if len(args) > 0 {
				// load from file.
				if err := loadFile(cmd, args[0], &connector); err != nil {
					return err
				}
			} else {
				// try load only the config from flag or file if possible.
				if err := tryReadFile(configRaw, &connector.Config); err != nil {
					return err
				}
			}

			if err := connector.ApplyAndValidateName(); err != nil {
				return err
			}

			if err := checkRequiredFlags(cmd, flags{"clusterName": connector.ClusterAlias, "name": connector.Name}); err != nil {
				return err
			}

			_, err := client.CreateConnector(connector.ClusterAlias, connector.Name, connector.Config)
			if err != nil {
				return err
			}

			if silent {
				return nil
			}

			return echo(cmd, "Connector %s created", connector.Name)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().StringVar(&configRaw, "config", "", `--config="{\"key\": \"value\"}"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}

func newConnectorUpdateCommand() *cobra.Command { // almost the same as `newConnectorCreateCommand` but keep them separate, in future this may change.
	var clusterName, name, configRaw string

	cmd := &cobra.Command{
		Use:              "update",
		Short:            "Update a connector's configuration",
		Example:          exampleString(`connector update --clusterName="cluster_name" --name="connector_name" --config="{\"key\": \"value\"}" or connector update ./connector.yml`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			connector := lenses.CreateUpdateConnectorPayload{
				ClusterAlias: clusterName,
				Name:         name,
				Config:       make(lenses.ConnectorConfig),
			}

			if len(args) > 0 {
				// load from file.
				if err := loadFile(cmd, args[0], &connector); err != nil {
					return err
				}
			} else {
				// try load only the config from flag or file if possible.
				if err := tryReadFile(configRaw, &connector.Config); err != nil {
					return err
				}
			}

			if err := connector.ApplyAndValidateName(); err != nil {
				return err
			}

			if err := checkRequiredFlags(cmd, flags{"clusterName": connector.ClusterAlias, "name": connector.Name}); err != nil {
				return err
			}

			// for any case.
			existingConnector, err := client.GetConnector(connector.ClusterAlias, connector.Name)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("connector '%s:%s' does not exist", connector.ClusterAlias, connector.Name)
				return err
			}

			if existingConnector.Config != nil {
				if existingNameValue := existingConnector.Config["name"]; existingNameValue != connector.Name {
					return fmt.Errorf(`connector config["name"] '%s' does not match with the existing one '%s'`, connector.Name, existingNameValue)
				}
			}

			updatedConnector, err := client.UpdateConnector(connector.ClusterAlias, connector.Name, connector.Config)
			if err != nil {
				return err
			}

			if silent {
				return nil
			}

			echo(cmd, "Connector %s updated\n\n", connector.Name)

			return printJSON(cmd, updatedConnector) // why we print it back? Because of the connector.Tasks.
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().StringVar(&configRaw, "config", "", `--config="{\"key\": \"value\"}"`)

	canPrintJSON(cmd)

	return cmd
}

func newConnectorGetConfigCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "config",
		Short:            "Get connector config",
		Example:          exampleString(`connector config --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			cfg, err := client.GetConnectorConfig(clusterName, name)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to retrieve config, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return printJSON(cmd, cfg)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)

	canPrintJSON(cmd)

	return cmd
}

func newConnectorGetStatusCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "status",
		Short:            "Get connector status",
		Example:          exampleString(`connector status --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			cs, err := client.GetConnectorStatus(clusterName, name)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to retrieve status, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return printJSON(cmd, cs)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)

	canPrintJSON(cmd)

	return cmd
}

func newConnectorPauseCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "pause",
		Short:            "Pause a connector",
		Example:          exampleString(`connector pause --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			if err := client.PauseConnector(clusterName, name); err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to pause, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return echo(cmd, "Connector %s:%s paused", clusterName, name)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}

func newConnectorResumeCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "resume",
		Short:            "Resume a paused connector",
		Example:          exampleString(`connector resume --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			if err := client.ResumeConnector(clusterName, name); err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to resume, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return echo(cmd, "Connector %s:%s resumed", clusterName, name)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}

func newConnectorRestartCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "restart",
		Short:            "Restart a connector",
		Example:          exampleString(`connector restart --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			if err := client.RestartConnector(clusterName, name); err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to restart, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return echo(cmd, "Connector %s:%s restarted", clusterName, name)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}

func newConnectorGetTasksCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "tasks",
		Short:            "List of connector tasks",
		Example:          exampleString(`connector tasks --clusterName="cluster_name" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			tasksMap, err := client.GetConnectorTasks(clusterName, name)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to retrieve tasks, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return printJSON(cmd, tasksMap)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)

	canPrintJSON(cmd)

	return cmd
}

func newConnectorTaskGroupCommand() *cobra.Command {
	rootSub := &cobra.Command{
		Use:              "task",
		Short:            "Work with a particular connector task, see connector task --help for details",
		Example:          exampleString(`connector task status --clusterName="cluster_name" --name="connector_name" --task=1`),
		SilenceErrors:    true,
		TraverseChildren: true,
	}

	rootSub.AddCommand(newConnectorGetCurrentTaskStatusCommand())
	rootSub.AddCommand(newConnectorTaskRestartCommand())

	return rootSub
}

func newConnectorGetCurrentTaskStatusCommand() *cobra.Command {
	var (
		clusterName, name string
		taskID            int
	)

	cmd := &cobra.Command{
		Use:              "status",
		Short:            "Get current status of a task",
		Example:          exampleString(`connector task status --clusterName="cluster_name" --name="connector_name" --task=1`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			cst, err := client.GetConnectorTaskStatus(clusterName, name, taskID)
			if err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("task does not exist")
				return err
			}

			return printJSON(cmd, cst)
		},
	}

	cmd.Flags().IntVar(&taskID, "task", 0, "--task=1 The Task ID")
	cmd.MarkFlagRequired("task")

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)

	canPrintJSON(cmd)

	return cmd
}

func newConnectorTaskRestartCommand() *cobra.Command {
	var (
		clusterName, name string
		taskID            int
	)

	cmd := &cobra.Command{
		Use:              "restart",
		Short:            "Restart a connector task",
		Example:          exampleString(`connector task restart --clusterName="cluster_name" --name="connector_name" --task=1`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			if err := client.RestartConnectorTask(clusterName, name, taskID); err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("task does not exist")
				return err
			}

			return echo(cmd, "Connector task %s:%s:%d restarted", clusterName, name, taskID)
		},
	}

	cmd.Flags().IntVar(&taskID, "task", 0, "--task=1 The Task ID")
	cmd.MarkFlagRequired("task")
	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}

func newConnectorDeleteCommand() *cobra.Command {
	var clusterName, name string

	cmd := &cobra.Command{
		Use:              "delete",
		Short:            "Delete a running connector",
		Example:          exampleString(`connector delete --clusterName="" --name="connector_name"`),
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkRequiredFlags(cmd, flags{"clusterName": clusterName, "name": name}); err != nil {
				return err
			}

			if err := client.DeleteConnector(clusterName, name); err != nil {
				errResourceNotFoundMessage = fmt.Sprintf("unable to delete, connector '%s:%s' does not exist", clusterName, name)
				return err
			}

			return echo(cmd, "Connector %s:%s deleted", clusterName, name)
		},
	}

	cmd.Flags().StringVar(&clusterName, "clusterName", "", `--clusterName="cluster_name"`)
	cmd.Flags().StringVar(&name, "name", "", `--name="connector_name"`)
	cmd.Flags().BoolVar(&silent, "silent", false, "run in silent mode. No printing info messages for CRUD except errors, defaults to false")

	return cmd
}
