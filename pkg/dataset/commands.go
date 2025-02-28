package dataset

import (
	"errors"
	"strings"

	"github.com/kataras/golog"
	"github.com/lensesio/bite"
	config "github.com/lensesio/lenses-go/pkg/configs"
	"github.com/spf13/cobra"
)

const metadataLong = `Description:
  Lenses can store a user-defined description for a Dataset (i.e. Kafka topics, ES indices).
  Be aware, that you need the "UpdateMetadata" permission to execute the command
`

// NewDatasetGroupCmd Group Cmd
func NewDatasetGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "dataset",
		Short:            "Use the dataset command to set user-defined metadata on Kafka topics and ES indices",
		SilenceErrors:    true,
		TraverseChildren: true,
		Args:             cobra.NoArgs,
	}

	cmd.AddCommand(UpdateDatasetDescriptionCmd())
	cmd.AddCommand(UpdateDatasetTagsCmd())
	cmd.AddCommand(RemoveDatasetDescriptionCmd())
	cmd.AddCommand(RemoveDatasetTagsCmd())
	return cmd
}

// UpdateDatasetDescriptionCmd updates the Dataset Metadata
func UpdateDatasetDescriptionCmd() *cobra.Command {
	var connection, name, description string

	cmd := &cobra.Command{
		Use:              "update-description [CONNECTION] [NAME]",
		Short:            "Set a dataset description",
		Long:             metadataLong,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(strings.TrimSpace(description)) == 0 {
				err := errors.New("--description value cannot be blank")
				golog.Errorf("Failed to update dataset description. [%s]", err.Error())
				return err
			}

			if err := config.Client.UpdateDatasetDescription(connection, name, description); err != nil {
				golog.Errorf("Failed to update dataset description. [%s]", err.Error())
				return err
			}
			return bite.PrintInfo(cmd, "Dataset description has been updated successfully")
		},
	}

	cmd.Flags().StringVar(&connection, "connection", "", "Name of the connection")
	cmd.Flags().StringVar(&name, "name", "", "Name of the dataset")
	cmd.Flags().StringVar(&description, "description", "", "Description of the dataset")
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("connection")
	cmd.MarkFlagRequired("name")

	_ = bite.CanBeSilent(cmd)

	return cmd
}

// UpdateDatasetTagsCmd updates the Dataset Metadata
func UpdateDatasetTagsCmd() *cobra.Command {
	var connection, name string
	var tags []string

	cmd := &cobra.Command{
		Use:   "update-tags [CONNECTION] [NAME]",
		Short: "Set a dataset tags",
		Example: `
		dataset update-tags --connection kafka \
		           --name mytopic \
				   --tag t1 \
				   --tag t2
		`,
		Long:             metadataLong,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(tags) == 0 {
				return errors.New("Tags cannot be empty")
			}

			if err := config.Client.UpdateDatasetTags(connection, name, tags); err != nil {
				return err
			}
			return bite.PrintInfo(cmd, "Dataset tags have been updated successfully")
		},
	}

	cmd.Flags().StringVar(&connection, "connection", "", "Name of the connection")
	cmd.Flags().StringVar(&name, "name", "", "Name of the dataset")
	cmd.Flags().StringArrayVar(&tags, "tag", []string{}, "tag assigned to the connection, can be defined multiple times")
	cmd.MarkFlagRequired("connection")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("tag")

	_ = bite.CanBeSilent(cmd)

	return cmd
}

//RemoveDatasetDescriptionCmd unsets a dataset description
func RemoveDatasetDescriptionCmd() *cobra.Command {
	var connection, name string

	cmd := &cobra.Command{
		Use:              "remove-description [CONNECTION] [NAME] [DESCRIPTION]",
		Short:            "Unsets a dataset description",
		Long:             metadataLong,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			//Setting the description to empty string will result in the field being omitted from the submitted JSON
			//which the backend will handle by unsetting the description record (see `UpdateDatasetDescription`'s
			//`omitempty` annotation)
			if err := config.Client.UpdateDatasetDescription(connection, name, ""); err != nil {
				return err
			}
			return bite.PrintInfo(cmd, "Dataset description has been removed")
		},
	}

	cmd.Flags().StringVar(&connection, "connection", "", "Name of the connection")
	cmd.Flags().StringVar(&name, "name", "", "Name of the dataset")
	cmd.MarkFlagRequired("connection")
	cmd.MarkFlagRequired("name")

	_ = bite.CanBeSilent(cmd)
	return cmd
}

//RemoveDatasetTagsCmd unsets a dataset description
func RemoveDatasetTagsCmd() *cobra.Command {
	var connection, name string

	cmd := &cobra.Command{
		Use:              "remove-tags [CONNECTION] [NAME]",
		Short:            "Remove all tags associated to a dataset",
		Long:             metadataLong,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Client.UpdateDatasetTags(connection, name, []string{}); err != nil {
				return err
			}
			return bite.PrintInfo(cmd, "Dataset tags have been removed")
		},
	}

	cmd.Flags().StringVar(&connection, "connection", "", "Name of the connection")
	cmd.Flags().StringVar(&name, "name", "", "Name of the dataset")
	cmd.MarkFlagRequired("connection")
	cmd.MarkFlagRequired("name")

	_ = bite.CanBeSilent(cmd)
	return cmd
}
