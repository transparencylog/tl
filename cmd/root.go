// Copyright © 2019 The Transparency Log Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.transparencylog.net/tl/cmd/cat"
	"go.transparencylog.net/tl/cmd/get"
	"go.transparencylog.net/tl/cmd/verify"
	"go.transparencylog.net/tl/cmd/version"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tl",
	Short: "Validate asset downloads against a public tranparency log ",
	Long: `tl validates assets (software downloads, documents, etc) downloaded from
https:// URLs with a cryptographic integrity validation. The public
transparency log tl uses provides users an assurance that the cryptographic
hash digest of the asset you are downloading does not differ from the value in
a public immutable log.`,
}

func init() {
	rootCmd.AddCommand(get.GetCmd)
	rootCmd.AddCommand(verify.VerifyCmd)
	rootCmd.AddCommand(cat.CatCmd)
	rootCmd.AddCommand(version.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
