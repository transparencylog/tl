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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cavaliercoder/grab"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.transparencylog.net/tl/clientcache/badger"
	"go.transparencylog.net/tl/sumdb"
)

var cfgFile string
var cacheFile string
var serverAddr string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tl [URL]",
	Short: "Get a URL and verify the contents with a binary tranparency log",
	Long: `tl is similar to other popular URL fetchers with an additional layer of security.
By using a transparency log, that enables third-party auditing, tl gives you
strong guarantees that the cryptographic hash digest of the binary you are
downloading appears in a public log.
`,

	Args: cobra.ExactArgs(1),

	Run: get,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	serverAddr = "beta-asset.transparencylog.net"
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/tl/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tlDir := filepath.Join(home, ".config", "tl")
	err = os.MkdirAll(tlDir, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cacheFile = filepath.Join(tlDir, "tl.badger.db")

	// Search config in home directory with name ".tl" (without extension).
	viper.AddConfigPath(tlDir)
	viper.SetConfigName("config")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func get(cmd *cobra.Command, args []string) {
	durl := args[0]

	u, err := url.Parse(durl)
	if err != nil {
		panic(err)
	}
	key := u.Host + u.Path

	// Step 0: Initialize cache if needed
	vkey := "log+3809a75e+ARmkoBH4C+/rbs9QomTtpLJQCkzfY171BfHZLEnmA/+e"
	cache := badger.NewClientCache(cacheFile, serverAddr)
	_, err = cache.ReadConfig("key")
	if err != nil {
		if err := cache.WriteConfig("key", nil, []byte(vkey)); err != nil {
			log.Fatal(err)
		}
	}

	// Step 1: Download the tlog entry for the URL
	client := sumdb.NewClient(cache)
	_, data, err := client.Lookup(key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fetched note: https://%s/lookup/%s\n", serverAddr, key)

	// create download request
	req, err := grab.NewRequest("", durl)
	if err != nil {
		fmt.Printf("failed to create grab request: %v\n", err)
		os.Exit(1)
	}
	req.NoCreateDirectories = true

	req.AfterCopy = func(resp *grab.Response) (err error) {
		var f *os.File
		f, err = os.Open(resp.Filename)
		if err != nil {
			return
		}
		defer func() {
			f.Close()
		}()

		h := sha256.New()
		_, err = io.Copy(h, f)
		if err != nil {
			return err
		}

		fileSum := h.Sum(nil)

		want := "h1:" + base64.StdEncoding.EncodeToString(fileSum)
		for _, line := range strings.Split(string(data), "\n") {
			if line == want {
				break
			}
			if strings.HasPrefix(line, "h1:") {
				log.Fatalf("file digest %x != log digest %x", fileSum, line)
			}
		}

		fmt.Printf("validated file sha256sum: %x\n", fileSum)

		req.SetChecksum(sha256.New(), fileSum, true)

		return
	}

	// download and validate file
	resp := grab.DefaultClient.Do(req)
	if err := resp.Err(); err != nil {
		fmt.Printf("Failed to grab: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Download validated and saved to", resp.Filename)
}
