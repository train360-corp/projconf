/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/train360-corp/projconf/go/cmd"
	"os"
)

func main() {
	err := cmd.ProjConf().Execute()
	if err != nil {
		if _, err := os.Stderr.WriteString(fmt.Sprintf("%v\n", color.RedString(err.Error()))); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
