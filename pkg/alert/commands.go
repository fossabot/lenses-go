package alert

import (
	"github.com/kataras/golog"
	"github.com/landoop/bite"
	"github.com/landoop/lenses-go/pkg/api"
	config "github.com/landoop/lenses-go/pkg/configs"
	"github.com/spf13/cobra"
)

//NewAlertGroupCommand creates the `alert` command
func NewAlertGroupCommand() *cobra.Command {
	root := &cobra.Command{
		Use:              "alert",
		Short:            "Manage alerts",
		Example:          "alert",
		TraverseChildren: true,
		SilenceErrors:    true,
	}

	root.AddCommand(
		NewGetAlertSettingsCommand(),
		NewAlertSettingGroupCommand(),
	)

	return root
}

//NewGetAlertsCommand creates the `alerts` command
func NewGetAlertsCommand() *cobra.Command {
	var (
		sse      bool
		pageSize int
	)

	cmd := &cobra.Command{
		Use:              "alerts",
		Short:            "Print the registered alerts",
		Example:          "alerts",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sse {
				handler := func(alert api.Alert) error {
					return bite.PrintObject(cmd, alert) // keep json here?
				}
				return config.Client.GetAlertsLive(handler)
			}
			alerts, err := config.Client.GetAlerts(pageSize)
			if err != nil {
				golog.Errorf("Failed to retrieve alerts. [%s]", err.Error())
				return err
			}
			return bite.PrintObject(cmd, alerts)
		},
	}

	cmd.Flags().BoolVar(&sse, "live", false, "Enables real-time push alert notifications")
	cmd.Flags().IntVar(&pageSize, "page-size", 25, "Size of items to be included in the list")

	bite.CanPrintJSON(cmd)

	return cmd
}

//NewGetAlertSettingsCommand creates the `alert settings` command
func NewGetAlertSettingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "settings",
		Short:            "Print all alert settings",
		Example:          "alert settings",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := config.Client.GetAlertSettings()
			if err != nil {
				return err
			}

			// force json, may contains conditions that are easier to be seen in json format.
			return bite.PrintJSON(cmd, settings)
		},
	}

	bite.CanPrintJSON(cmd)

	return cmd
}

//NewAlertSettingGroupCommand creates the `alert setting` command
func NewAlertSettingGroupCommand() *cobra.Command {
	var (
		id     int
		enable bool
	)

	cmd := &cobra.Command{
		Use:              "setting",
		Short:            "Print or enable a specific alert setting based on ID",
		Example:          "alert setting --id=1001 [--enable]",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("enable") {
				if err := config.Client.EnableAlertSetting(id, enable); err != nil {
					return err
				}

				if enable {
					return bite.PrintInfo(cmd, "Alert setting [%d] enabled", id)
				}
				return bite.PrintInfo(cmd, "Alert setting [%d] disabled", id)
			}

			settings, err := config.Client.GetAlertSetting(id)
			if err != nil {
				golog.Errorf("Failed to retrieve alert [%d]. [%s]", id, err.Error())
				return err
			}

			// force json, may contains conditions that are easier to be seen in json format.
			return bite.PrintObject(cmd, settings)
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "--id=1001")
	cmd.MarkFlagRequired("id")

	cmd.Flags().BoolVar(&enable, "enable", false, "--enable")

	bite.CanPrintJSON(cmd)
	bite.CanBeSilent(cmd)

	cmd.AddCommand(NewGetAlertSettingConditionsCommand())
	cmd.AddCommand(NewAlertSettingConditionGroupCommand())

	return cmd
}

//NewGetAlertSettingConditionsCommand creates `alert setting conditions`
func NewGetAlertSettingConditionsCommand() *cobra.Command {
	var alertID int

	cmd := &cobra.Command{
		Use:              "conditions",
		Short:            "Print alert setting's conditions",
		Example:          "alert setting conditions --alert=1001",
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conds, err := config.Client.GetAlertSettingConditions(alertID)
			if err != nil {
				golog.Errorf("Failed to retrieve alert setting conditions for [%d]. [%s]", alertID, err.Error())
				return err
			}

			// force-json
			return bite.PrintObject(cmd, conds)
		},
	}

	cmd.Flags().IntVar(&alertID, "alert", 0, "--alert=1001")
	cmd.MarkFlagRequired("alert")

	bite.CanBeSilent(cmd)
	bite.CanPrintJSON(cmd)

	return cmd
}

//NewAlertSettingConditionGroupCommand creates `alert setting condition`
func NewAlertSettingConditionGroupCommand() *cobra.Command {
	rootSub := &cobra.Command{
		Use:              "condition",
		Short:            "Manage alert setting's condition",
		Example:          `alert setting condition delete --alert=1001 --condition="28bbad2b-69bb-4c01-8e37-28e2e7083aa9"`,
		TraverseChildren: true,
		SilenceErrors:    true,
	}

	rootSub.AddCommand(NewCreateOrUpdateAlertSettingConditionCommand())
	rootSub.AddCommand(NewDeleteAlertSettingConditionCommand())

	return rootSub
}

//NewCreateOrUpdateAlertSettingConditionCommand creates `alert condition set` command
func NewCreateOrUpdateAlertSettingConditionCommand() *cobra.Command {
	var conds SettingConditionPayloads
	var cond SettingConditionPayload

	cmd := &cobra.Command{
		Use:              "set",
		Aliases:          []string{"create", "update"},
		Short:            "Create or Update an alert setting's condition or load from file",
		Example:          `alert setting condition set --alert=1001 --condition="lag >= 100000 or alert setting condition set ./alert_cond.yml`,
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(conds.Conditions) > 0 {
				alertID := conds.AlertID
				for _, condition := range conds.Conditions {
					err := config.Client.CreateOrUpdateAlertSettingCondition(alertID, condition)
					if err != nil {
						golog.Errorf("Failed to creating/updating alert setting condition [%s]. [%s]", condition, err.Error())
						return err
					}
					bite.PrintInfo(cmd, "Condition [id=%d] added", alertID)
				}
				return nil
			}
			if err := bite.CheckRequiredFlags(cmd, bite.FlagPair{"alert": cond.AlertID, "condition": cond.Condition}); err != nil {
				return err
			}

			err := config.Client.CreateOrUpdateAlertSettingCondition(cond.AlertID, cond.Condition)
			if err != nil {
				golog.Errorf("Failed to creating/updating alert setting condition [%s]. [%s]", cond.Condition, err.Error())
				return err
			}

			return bite.PrintInfo(cmd, "Condition [id=%d] added", cond.AlertID)
		},
	}

	cmd.Flags().IntVar(&cond.AlertID, "alert", 0, "Alert ID")
	cmd.Flags().StringVar(&cond.Condition, "condition", "", `Alert condition .e.g. "lag >= 100000 on group group and topic topicA"`)

	bite.CanBeSilent(cmd)

	bite.Prepend(cmd, bite.FileBind(&cond))
	bite.Prepend(cmd, bite.FileBind(&conds))

	return cmd
}

//NewDeleteAlertSettingConditionCommand creates `alert condition delete` command
func NewDeleteAlertSettingConditionCommand() *cobra.Command {
	var (
		alertID       int
		conditionUUID string
	)

	cmd := &cobra.Command{
		Use:              "delete",
		Short:            "Delete an alert setting's condition",
		Example:          `alert setting condition delete --alert=1001 --condition="28bbad2b-69bb-4c01-8e37-28e2e7083aa9"`,
		TraverseChildren: true,
		SilenceErrors:    true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.Client.DeleteAlertSettingCondition(alertID, conditionUUID)
			if err != nil {
				golog.Errorf("Failed to deleting alert setting condition [%s]. [%s]", conditionUUID, err.Error())
				return err
			}

			return bite.PrintInfo(cmd, "Condition [%s] for alert setting [%d] deleted", conditionUUID, alertID)
		},
	}

	cmd.Flags().IntVar(&alertID, "alert", 0, "Alert ID")
	cmd.MarkFlagRequired("alert")
	cmd.Flags().StringVar(&conditionUUID, "condition", "", `Alert condition uuid .e.g. "28bbad2b-69bb-4c01-8e37-28e2e7083aa9"`)
	cmd.MarkFlagRequired("condition")
	bite.CanBeSilent(cmd)

	return cmd
}
