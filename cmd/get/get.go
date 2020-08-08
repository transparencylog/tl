package get

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/cavaliercoder/grab"
	"github.com/spf13/cobra"
	"go.transparencylog.net/tl/config"
	"go.transparencylog.net/tl/sumdb"
)

var GetCmd = &cobra.Command{
	Use:   "get [URL]",
	Short: "Download a URL to a local file and verify the contents with the asset tranparency log",

	Args: cobra.ExactArgs(1),

	Run: get,
}

func get(cmd *cobra.Command, args []string) {
	durl := args[0]

	u, err := url.Parse(durl)
	if err != nil {
		panic(err)
	}
	key := u.Host + u.Path

	cache := config.ClientCache()

	// Step 1: Download the tlog entry for the URL
	client := sumdb.NewClient(cache)
	_, data, err := client.Lookup(key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fetched note: https://%s/lookup/%s\n", config.ServerAddr, key)

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
