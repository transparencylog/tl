package verify

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.transparencylog.net/tl/config"
	"go.transparencylog.net/tl/sumdb"
)

var VerifyCmd = &cobra.Command{
	Use:   "verify [URL] [file]",
	Short: "Verify the contents of a locally downloaded file with the asset tranparency log",

	Args: cobra.ExactArgs(2),

	Run: verify,
}

func verify(cmd *cobra.Command, args []string) {
	durl := args[0]
	file := args[1]

	_ = file

	u, err := url.Parse(durl)
	if err != nil {
		panic(err)
	}
	key := u.Host + u.Path

	cache := config.ClientCache()
	client := sumdb.NewClient(cache)

	// Step 1: Generate sha256sum of the file
	f, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		log.Fatal(err)
	}

	fileSum := h.Sum(nil)

	want := "h1:" + base64.StdEncoding.EncodeToString(fileSum)

	// Step 1: Download the tlog entry for the URL
	_, data, err := client.LookupOpts(key, sumdb.LookupOpts{Digest: want})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fetched note: %s/lookup/%s\n", config.ServerURL, key)

	for _, line := range strings.Split(string(data), "\n") {
		if line == want {
			break
		}
		if strings.HasPrefix(line, "h1:") {
			log.Fatalf("file digest %x != log digest %x", fileSum, line)
		}
	}

	fmt.Printf("validated file sha256sum: %x\n", fileSum)
}
