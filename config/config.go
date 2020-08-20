package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"go.transparencylog.com/tl/clientcache/badger"
)

var Version string
var Commit string
var Date string

var ServerURL string = "https://beta-asset.transparencylog.net"
var ServerKey string = "log+3809a75e+ARmkoBH4C+/rbs9QomTtpLJQCkzfY171BfHZLEnmA/+e"

func init() {
	s := os.Getenv("TL_DEBUG_SERVERURL")
	if s != "" {
		ServerURL = s
	}
	s = os.Getenv("TL_DEBUG_SERVERKEY")
	if s != "" {
		ServerKey = s
	}
}

// ClientCache returns an initialized ClientCache using ServerURL and ServerKey
func ClientCache() *badger.ClientCache {
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

	cacheFile := filepath.Join(tlDir, "tl.badger.db")

	// Initialize cache DB, if necessary
	cache := badger.NewClientCache(cacheFile, ServerURL)
	_, err = cache.ReadConfig("key")
	if err != nil {
		if err := cache.WriteConfig("key", nil, []byte(ServerKey)); err != nil {
			log.Fatal(err)
		}
	}

	return cache
}
