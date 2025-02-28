package topic

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kataras/golog"
	"github.com/lensesio/bite"
	"github.com/lensesio/lenses-go/pkg/api"
	config "github.com/lensesio/lenses-go/pkg/configs"
	"github.com/spf13/cobra"
)

type topicView struct {
	api.Topic `yaml:",inline" header:"inline"`
	// for machine view-only.
	ValueSchema json.RawMessage `json:"valueSchema" yaml:"-"`
	KeySchema   json.RawMessage `json:"keySchema" yaml:"-"`
}

//NewTopicsGroupCommand creates `topics` command
func NewTopicsGroupCommand() *cobra.Command {
	var namesOnly, unwrap bool

	root := &cobra.Command{
		Use:           "topics",
		Short:         "List all available topics",
		Example:       "topics",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client

			if namesOnly {
				topicNames, err := client.GetTopicsNames()
				if err != nil {
					return err
				}
				sort.Strings(topicNames)

				if unwrap {
					for _, name := range topicNames {
						fmt.Fprintln(cmd.OutOrStdout(), name)
					}
					return nil
				}

				// return printJSON(cmd, outlineStringResults("name", topicNames))
				return bite.PrintObject(cmd, bite.OutlineStringResults(cmd, "name", topicNames))
			}

			topics, err := client.GetTopics()
			if err != nil {
				return err
			}

			sort.Slice(topics, func(i, j int) bool {
				return topics[i].TopicName < topics[j].TopicName
			})

			topicsView := make([]topicView, len(topics))
			for i, topic := range topics {
				topicsView[i] = newTopicView(cmd, client, topic)
			}

			// return printJSON(cmd, topics)
			return bite.PrintObject(cmd, topicsView, func(t topicView) bool {
				return !t.IsControlTopic // on JSON we print everything.
			})
		},
	}

	root.Flags().BoolVar(&namesOnly, "names", false, "Print topic names only")
	root.Flags().BoolVar(&unwrap, "unwrap", false, "--unwrap")

	bite.CanPrintJSON(root)

	root.AddCommand(NewGetAvailableTopicConfigKeysCommand())
	root.AddCommand(NewTopicsMetadataSubgroupCommand())

	return root
}

//NewGetAvailableTopicConfigKeysCommand creates `topics keys` command
func NewGetAvailableTopicConfigKeysCommand() *cobra.Command {
	var unwrap bool

	cmd := &cobra.Command{
		Use:           "keys",
		Short:         "List all available config keys for topics",
		Example:       "topics keys",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			keys, err := config.Client.GetAvailableTopicConfigKeys()
			if err != nil {
				return err
			}

			sort.Strings(keys)

			if unwrap {
				for _, key := range keys {
					fmt.Fprintln(cmd.OutOrStdout(), key)
				}

				return nil
			}

			return bite.PrintObject(cmd, bite.OutlineStringResults(cmd, "key", keys))
		},
	}

	cmd.Flags().BoolVar(&unwrap, "unwrap", false, "--unwrap Display the names separated by new lines, disables the Table or JSON view")

	bite.CanPrintJSON(cmd)

	return cmd
}

//NewTopicsMetadataSubgroupCommand cfreates `topics metadata` command
func NewTopicsMetadataSubgroupCommand() *cobra.Command {
	var topicName string

	rootSub := &cobra.Command{
		Use:           "metadata",
		Short:         "List all available topics metadata",
		Example:       "topics metadata",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client
			if topicName != "" {
				// view single.
				meta, err := client.GetTopicMetadata(topicName)
				if err != nil {
					return err
				}

				viewMeta, err := newTopicMetadataView(meta)
				if err != nil {
					return err
				}

				return bite.PrintObject(cmd, viewMeta)
			}

			metas, err := client.GetTopicsMetadata()
			if err != nil {
				return err
			}

			sort.Slice(metas, func(i, j int) bool {
				return metas[i].TopicName < metas[j].TopicName
			})

			viewMetas := make([]topicMetadataView, len(metas), len(metas))

			for i, m := range metas {
				viewMetas[i], err = newTopicMetadataView(m)
				if err != nil {
					return err
				}
			}

			return bite.PrintObject(cmd, viewMetas)
		},
	}

	rootSub.Flags().StringVar(&topicName, "name", "", "Topic to return metadata for")

	bite.CanPrintJSON(rootSub)

	rootSub.AddCommand(NewTopicMetadataDeleteCommand())
	rootSub.AddCommand(NewTopicMetadataCreateCommand())

	return rootSub
}

//NewTopicMetadataDeleteCommand creates `topics metadata delete` command
func NewTopicMetadataDeleteCommand() *cobra.Command {
	var topicName string

	cmd := &cobra.Command{
		Use:              "delete",
		Short:            "Delete a topic's metadata",
		Example:          `topics metadata delete --name="topicName"`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": topicName}); err != nil {
				return err
			}

			if err := config.Client.DeleteTopicMetadata(topicName); err != nil {
				golog.Errorf("Failed to delete topic metadata [%s]. [%s]", topicName, err.Error())
				return err
			}

			return bite.PrintInfo(cmd, "Metadata for topic [%s] deleted", topicName)
		},
	}

	cmd.Flags().StringVar(&topicName, "name", "", "Topic to delete")

	bite.CanBeSilent(cmd)

	return cmd
}

//NewTopicMetadataCreateCommand creates `topics metadata create` command
func NewTopicMetadataCreateCommand() *cobra.Command {
	var meta api.TopicMetadata

	cmd := &cobra.Command{
		Use:              "set",
		Aliases:          []string{"create", "update"},
		Short:            "Create or update an existing topic metadata",
		Example:          `topics metadata set ./topic_metadata.yml`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": meta.TopicName}); err != nil {
				return err
			}

			if err := client.CreateOrUpdateTopicMetadata(meta); err != nil {
				return fmt.Errorf("Failed to update topic metadata for [%s]. [%s]", meta.TopicName, err.Error())
			}

			return bite.PrintInfo(cmd, "Metadata for topic [%s] created/updated", meta.TopicName)
		},
	}

	cmd.Flags().StringVar(&meta.TopicName, "name", "", "Topic name to update/create metadata for")
	cmd.Flags().StringVar(&meta.KeyType, "key-type", "", "Topic keyType")
	cmd.Flags().StringVar(&meta.ValueType, "value-type", "", "Topic's value type")
	cmd.Flags().StringVar(&meta.KeySchemaRaw, "key-schema", "", "Topic's key schema")
	cmd.Flags().StringVar(&meta.ValueSchemaRaw, "value-schema", "", "Topic's value schema")
	bite.CanBeSilent(cmd)

	bite.Prepend(cmd, bite.FileBind(&meta))

	return cmd
}

//NewTopicGroupCommand creates `topic` command
func NewTopicGroupCommand() *cobra.Command {
	var topicName string

	root := &cobra.Command{
		Use:              "topic",
		Short:            "Manage particular topic based on the topic name, retrieve it or create a new one",
		Example:          `topic --name="existing_topic_name" or topic create --name="topic1" --replication=1 --partitions=1 --configs="{\"key\": \"value\"}"`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": topicName}); err != nil {
				return err
			}

			// default is the retrieval of the particular topic info.
			topic, err := client.GetTopic(topicName)
			if err != nil {
				golog.Errorf("Failed to retrieve topic [%s]. [%s]", topic.TopicName, err.Error())
				return err
			}

			return bite.PrintObject(cmd, newTopicView(cmd, client, topic))
		},
	}

	root.Flags().StringVar(&topicName, "name", "", "Topic name")
	bite.CanPrintJSON(root)

	// subcommands
	root.AddCommand(NewTopicCreateCommand())
	root.AddCommand(NewTopicDeleteCommand())
	root.AddCommand(NewTopicUpdateCommand())

	return root
}

//NewTopicCreateCommand creates `topic create` command
func NewTopicCreateCommand() *cobra.Command {
	var (
		configsRaw string
		topic      = api.CreateTopicPayload{
			Replication: 1,
			Partitions:  1,
			Configs:     api.KV{},
		}
	)

	cmd := &cobra.Command{
		Use:              "create",
		Short:            "Create a new topic",
		Example:          `topic create --name="topic1" --replication=1 --partitions=1 --configs="{\"max.message.bytes\": \"1000010\"}"`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": topic.TopicName}); err != nil {
				return err
			}

			if configsRaw != "" {
				if err := bite.TryReadFile(configsRaw, &topic.Configs); err != nil {
					// from flag as json.
					if err = json.Unmarshal([]byte(configsRaw), &topic.Configs); err != nil {
						return fmt.Errorf("Unable to unmarshal the configs: [%v]", err)
					}
				}
			}

			if err := config.Client.CreateTopic(topic.TopicName, topic.Replication, topic.Partitions, topic.Configs); err != nil {
				golog.Errorf("Failed to create topic [%s]. [%s]", topic.TopicName, err.Error())
				return err
			}

			return bite.PrintInfo(cmd, "Topic [%s] created", topic.TopicName)
		},
	}

	cmd.Flags().StringVar(&topic.TopicName, "name", "", "Topic name")
	cmd.Flags().IntVar(&topic.Replication, "replication", topic.Replication, "Topic replication factor")
	cmd.Flags().IntVar(&topic.Partitions, "partitions", topic.Partitions, "Number of partitions")
	cmd.Flags().StringVar(&configsRaw, "configs", "", `Topic configs .e.g. "{\"max.message.bytes\": \"1000010\"}"`)
	bite.CanBeSilent(cmd)
	bite.Prepend(cmd, bite.FileBind(&topic))

	return cmd
}

//NewTopicDeleteCommand creates `topic delete` command
func NewTopicDeleteCommand() *cobra.Command {
	var (
		topicName string
		// and for records with offset.
		fromPartition int
		toOffset      int64
	)

	cmd := &cobra.Command{
		Use:              "delete",
		Short:            "Delete a topic",
		Example:          `topic delete --name="topic1" [--partition=0 --offset=1260]`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": topicName}); err != nil {
				return err
			}

			if fromPartition >= 0 && toOffset >= 0 {
				// delete records.
				if err := client.DeleteTopicRecords(topicName, fromPartition, toOffset); err != nil {
					golog.Errorf("Failed to delete records topic [%s]. [%s]", topicName, err.Error())
					return err
				}

				return bite.PrintInfo(cmd, "Records from topic [%s] and partition [%d] up to offset [%d], are marked for deletion. This may take a few moments to have effect", topicName, fromPartition, toOffset)
			}

			if err := client.DeleteTopic(topicName); err != nil {
				golog.Errorf("Failed to delete topic [%s]. [%s]", topicName, err.Error())
				return err
			}

			return bite.PrintInfo(cmd, "Topic [%s] marked for deletion. This may take a few moments to have effect", topicName)
		},
	}

	cmd.Flags().StringVar(&topicName, "name", "", "Topic name to delete from")

	// negative default values because 0 is valid value.
	cmd.Flags().IntVar(&fromPartition, "partition", -1, "Deletes records from a specific partition (offset must set)")
	cmd.Flags().Int64Var(&toOffset, "offset", -1, "Deletes records from a specific offset (partition must set)")
	bite.CanBeSilent(cmd)

	return cmd
}

//NewTopicUpdateCommand creates `topic update` command
func NewTopicUpdateCommand() *cobra.Command {
	var (
		configsRaw string
		topic      = api.CreateTopicPayload{Configs: api.KV{}}
		partitions int
	)

	cmd := &cobra.Command{
		Use:              "update",
		Short:            "Update a topic's configs (as an array of config key-value map)",
		Example:          `topic update --name="topic1" --configs="{\"key\": \"max.message.bytes\", \"value\": \"1000020\"} or topic update ./topic.yml`,
		SilenceErrors:    true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"name": topic.TopicName}); err != nil {
				return err
			}

			confs := []api.KV{topic.Configs}

			if configsRaw != "" {
				var cfg api.KV
				if err := bite.TryReadFile(configsRaw, &topic.Configs); err != nil {
					// from flag as json.
					if err = json.Unmarshal([]byte(configsRaw), &cfg); err != nil {
						return fmt.Errorf("Unable to unmarshal the configs: %v", err)
					}
				}

				confs = append(confs, cfg)
			}

			if topic.Partitions != 0 {
				partitions = topic.Partitions
			}
			if err := config.Client.UpdateTopic(topic.TopicName, []api.KV{topic.Configs}, partitions); err != nil {
				return fmt.Errorf("failed to update topic [%s]. [%s]", topic.TopicName, err.Error())
			}

			return bite.PrintInfo(cmd, "Config updated for topic [%s]", topic.TopicName)
		},
	}

	cmd.Flags().StringVar(&topic.TopicName, "name", "", "Topic to update")
	cmd.Flags().StringVar(&configsRaw, "configs", "", `Topic configs .e.g. "{\"key\": \"max.message.bytes\", \"value\": \"1000020\"}"`)
	cmd.Flags().IntVar(&partitions, "partitions", 0, "Number of partitions (can only be increased)")
	bite.CanBeSilent(cmd)
	bite.Prepend(cmd, bite.FileBind(&topic))

	return cmd
}
