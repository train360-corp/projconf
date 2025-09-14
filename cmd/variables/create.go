/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package variables

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/internal/flags"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"strings"
)

var (
	createVariableTypeStatic       bool // static generator
	createVariableStaticValue      string
	createVariableStaticValueEmpty bool

	createVariableTypeRandom            bool // random generator
	createVariableRandomValueLength     int
	createVariableRandomValueUseNumbers bool
	createVariableRandomValueUseLetters bool
	createVariableRandomValueUseSymbols bool
)

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
	RunE: func(c *cobra.Command, args []string) error {

		client, _ := api.FromFlags(authFlags)

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

		resp, err := client.CreateVariableV1WithResponse(c.Context(), projectId, req)
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

	createVariableCmd.Flags().StringVar(&projectIdStr, flags.ProjectIdFlag, "", "the id of the project to create the variable with")
	err := createVariableCmd.MarkFlagRequired(flags.ProjectIdFlag)
	if err != nil {
		panic(err)
	}

	flags.SetupAuthFlags(createVariableCmd, authFlags)
	viper.BindPFlags(createVariableCmd.Flags())
}
