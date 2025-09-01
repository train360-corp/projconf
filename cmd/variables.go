/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/internal/utils/tables"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"strings"
)

var createVariableTypeStatic bool // static generator
var createVariableStaticValue string
var createVariableStaticValueEmpty bool

var createVariableTypeRandom bool // random generator
var createVariableRandomValueLength int
var createVariableRandomValueUseNumbers, createVariableRandomValueUseLetters, createVariableRandomValueUseSymbols bool

var variablesCmd = &cobra.Command{
	Use:           "variables",
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage variables in a ProjConf server instance",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return isReady()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var listVariablesCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List variables in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(projectIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid project id (%v)", projectIdStr, err)
		}
		projectId = id
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.GetVariablesV1WithResponse(cmd.Context(), projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no variables found")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tables.Build[api.Variable](
					*resp.JSON200,
					tables.ColumnsByFieldNames[api.Variable]("Id", "Key", "GeneratorType", "GeneratorData"),
					tables.WithTitle("Variables"),
					tables.WithStyle(table.StyleLight),
				))
			}
		} else {
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

var createVariableCmd = &cobra.Command{
	Use:           "create",
	Aliases:       []string{"new"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.ExactArgs(1),
	Short:         "Create a variable in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		id, err := uuid.Parse(projectIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid project id (%v)", projectIdStr, err)
		}
		projectId = id

		if !validators.IsValidVariable(args[0]) {
			return fmt.Errorf("\"%v\" is not a valid variable name", args[0])
		}

		if createVariableTypeRandom {
			if createVariableRandomValueLength < 1 {
				return fmt.Errorf("\"%v\" is not a valid variable length (min: 1)", createVariableTypeRandom)
			}
			if !createVariableRandomValueUseNumbers && !createVariableRandomValueUseLetters && !createVariableRandomValueUseSymbols {
				return fmt.Errorf("at least one of --numbers, --letters, or --symbols must be provided")
			}
		}

		if createVariableTypeStatic && createVariableStaticValue == "" && !createVariableStaticValueEmpty {
			return fmt.Errorf("\"%v\" is required when using static variable type (if you want the value to be empty, use --empty)", args[0])
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)

		req := api.CreateVariableV1JSONRequestBody{
			Key: args[0],
		}

		found := false
		if createVariableTypeRandom {
			found = true
			req.Generator.FromSecretGeneratorRandom(api.SecretGeneratorRandom{
				Type: api.SecretGeneratorRandomType(api.GeneratorTypeRANDOM),
				Data: api.RandomGeneratorData{
					Length:  float32(createVariableRandomValueLength),
					Letters: createVariableRandomValueUseLetters,
					Symbols: createVariableRandomValueUseSymbols,
					Numbers: createVariableRandomValueUseNumbers,
				},
			})
		}
		if createVariableTypeStatic {
			found = true
			value := ""
			if !createVariableStaticValueEmpty {
				value = createVariableStaticValue
			}
			req.Generator.FromSecretGeneratorStatic(api.SecretGeneratorStatic{
				Type: api.SecretGeneratorStaticType(api.GeneratorTypeSTATIC),
				Data: value,
			})
		}
		if !found {
			return errors.New("an unexpected error occurred when choosing the variable type to generate")
		}

		resp, err := client.CreateVariableV1WithResponse(cmd.Context(), projectId, req)
		if err != nil {
			return fmt.Errorf("request failed: %v", err.Error())
		}

		if resp.JSON201 != nil {
			fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
		} else {
			if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
				return fmt.Errorf("variable \"%s\" already exists", args[0])
			}
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

func init() {

	// random value generator
	createVariableCmd.Flags().BoolVar(&createVariableTypeRandom, "random", false, "generate a random value")
	createVariableCmd.Flags().BoolVar(&createVariableRandomValueUseLetters, "letters", false, "include letters in a random value")
	createVariableCmd.Flags().BoolVar(&createVariableRandomValueUseNumbers, "numbers", false, "include numbers in a random value")
	createVariableCmd.Flags().BoolVar(&createVariableRandomValueUseSymbols, "symbols", false, "include symbols in a random value")
	createVariableCmd.Flags().IntVar(&createVariableRandomValueLength, "length", 1, "length of the random value (min: 1)")

	// static value generator
	createVariableCmd.Flags().BoolVar(&createVariableTypeStatic, "static", false, "generate a static value")
	createVariableCmd.Flags().StringVar(&createVariableStaticValue, "value", "", "the value to use for a static generator")
	createVariableCmd.Flags().BoolVar(&createVariableStaticValueEmpty, "empty", false, "generate an empty static value")
	createVariableCmd.MarkFlagsMutuallyExclusive("value", "empty")

	// XOR
	createVariableCmd.MarkFlagsOneRequired("random", "static")
	createVariableCmd.MarkFlagsMutuallyExclusive("random", "static")

	projectIdFlag := "project-id"
	listVariablesCmd.Flags().StringVar(&projectIdStr, projectIdFlag, "", "the id of the project to list variables for")
	listVariablesCmd.MarkFlagRequired(projectIdFlag)

	createVariableCmd.Flags().StringVar(&projectIdStr, projectIdFlag, "", "the id of the project to create the variable with")
	createVariableCmd.MarkFlagRequired(projectIdFlag)

	addAuthFlags(listVariablesCmd)
	addAuthFlags(createVariableCmd)
	variablesCmd.AddCommand(listVariablesCmd)
	variablesCmd.AddCommand(createVariableCmd)
	viper.BindPFlags(listVariablesCmd.Flags())
	viper.BindPFlags(createVariableCmd.Flags())
	rootCmd.AddCommand(variablesCmd)
}
