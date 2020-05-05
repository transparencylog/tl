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
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cavaliercoder/grab"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "rsc.io/sqlite"

	"go.transparencylog.net/btget/sumdb"
)

var cfgFile string
var cacheFile string
var serverAddr string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "btget [URL]",
	Short: "Get a URL and verify the contents with a binary tranparency log",
	Long: `btget is similar to other popular URL fetchers with an additional layer of security.
By using a transparency log, that enables third-party auditing, btget gives you
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
	serverAddr = "binary.transparencylog.net"
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/btget/config.yaml)")

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

	btgetDir := filepath.Join(home, ".config", "btget")
	err = os.MkdirAll(btgetDir, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cacheFile = filepath.Join(btgetDir, "btget.db")

	// Search config in home directory with name ".btget" (without extension).
	viper.AddConfigPath(btgetDir)
	viper.SetConfigName("config")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

type clientCache struct {
	sql *sql.DB
}

func NewClientCache() *clientCache {
	client := &clientCache{}

	sdb, err := sql.Open("sqlite3", cacheFile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := sdb.Exec(`create table if not exists kv (k primary key, v)`); err != nil {
		log.Fatal(err)
	}
	client.sql = sdb
	return client
}

func (c *clientCache) ReadRemote(path string) ([]byte, error) {
	resp, err := http.Get("https://" + serverAddr + path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http get: %v", resp.Status)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c *clientCache) ReadConfig(file string) (data []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("read config %s: %v", file, err)
		}
	}()

	data, err = sqlRead(c.sql, "config:"+file)
	if strings.HasSuffix(file, "/latest") && err == sql.ErrNoRows {
		return nil, nil
	}
	return data, err
}

func (c *clientCache) WriteConfig(file string, old, new []byte) error {
	if old == nil {
		return sqlWrite(c.sql, "config:"+file, new)
	}
	return sqlSwap(c.sql, "config:"+file, old, new)
}

func (c *clientCache) ReadCache(file string) ([]byte, error) {
	return sqlRead(c.sql, "file:"+file)
}

func (c *clientCache) WriteCache(file string, data []byte) {
	sqlWrite(c.sql, "file:"+file, data)
}

func (c *clientCache) Log(msg string) {
	log.Print(msg)
}

func (c *clientCache) SecurityError(msg string) {
	log.Fatal(msg)
}

func get(cmd *cobra.Command, args []string) {
	durl := args[0]

	u, err := url.Parse(durl)
	if err != nil {
		panic(err)
	}
	key := u.Host + u.Path

	// Step 0: Initialize cache if needed
	vkey := "log+998cdb6b+AUDa+aCu48rSILe2yaFwjrL5p3h5jUi4x4tTX0wSpeXU"
	cache := NewClientCache()
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
	logDigest := strings.Trim(string(data), "\n")

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

		logSum, err := hex.DecodeString(logDigest)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(fileSum, logSum) {
			log.Fatalf("file digest %x != log digest %x", fileSum, logSum)
		}

		fmt.Printf("validated file sum: %x\n", fileSum)

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

func sqlRead(db *sql.DB, key string) ([]byte, error) {
	var value []byte
	err := db.QueryRow(`select v from kv where k = ?`, key).Scan(&value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func sqlWrite(db *sql.DB, key string, value []byte) error {
	_, err := db.Exec(`insert or replace into kv (k, v) values (?, ?)`, key, value)
	return err
}

func sqlSwap(db *sql.DB, key string, old, value []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var txOld []byte
	if err := tx.QueryRow(`select v from kv where k = ?`, key).Scan(&txOld); err != nil {
		return err
	}
	if !bytes.Equal(txOld, old) {
		return sumdb.ErrWriteConflict
	}
	if _, err := tx.Exec(`insert or replace into kv (k, v) values (?, ?)`, key, value); err != nil {
		return err
	}
	return tx.Commit()
}
