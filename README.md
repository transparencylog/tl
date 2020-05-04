# btget

[![Go Report Card](https://goreportcard.com/badge/github.com/transparencylog/btget)](https://goreportcard.com/report/github.com/transparencylog/btget)

**Alpha Warning**: btget works as described but has minimal testing, no peer review, and no load testing. Please test the tool and provide feedback.

`btget` downloads URLs and verifies the contents against a publicly recorded cryptographic log. The public log gives users of btget a number of useful properties:

- Verifiability of a downloaded URL's contents being identical to what the rest of the world sees
- Searchability of recorded content changes of a URL
- Notifications to any interested party about changes to the URLs contents

`btget` relies on a public webservice, hosted at binary.tranparencylog.net, which keeps an append only log of the cryptographic digests of all URLs it has seen. If a URL has not been seen before the service downloads the URL and stores a cryptographic digest of those contents in the log. The `btget` tool will then download the digest from the service and verify this digest matches the contents of the files the user retrieved.

## Installation

Download the appropriate release from https://github.com/transparencylog/btget/releases and extract the archive

## Example Usage

Use btget to download the v3.4.2 [etcd](https://etcd.io) release for macOS and verify that the contents are publicly recorded.

```
btget https://github.com/etcd-io/etcd/releases/download/v3.4.2/etcd-v3.4.2-darwin-amd64.zip
```

## FAQ

If you have a question that isn't answered here please [open an issue](https://github.com/transparencylog/btget/issues/new). 

- **Q**: Where did this idea come from?
- **A**: This project draws inspiration from a variety of sources. Technically it relies on a large amount of the ecosystem developed for the [Go modules ecosystem](https://proxy.golang.org). It also is inspired by a design doc for [Binary Transparency](https://wiki.mozilla.org/Security/Binary_Transparency) originally designed for the Mozilla Firefox project. 

- **Q**: Why not build this ontop of blockchain?
- **A**: Blockchain could be used to create a similar guarantee. However, using transparency log technology extends a number of advantages and was a pragmatic choice to get this project going: the industry momentum of transparency log technology [(1)](https://ct.cloudflare.com/about) and [Go modules](https://proxy.golang.org), leverage existing web technologies like DNS and TLS, and finally most practical applications that want to use blockchain with the web end up using a centralized gateway for speed and reliability [(3)](https://blog.cloudflare.com/cloudflare-ethereum-gateway/)[(4)](https://infura.io/docs/ethereum/json-rpc/eth_blockNumber). Perhaps as the bridge between the web and blockchain matures it will become a more practical option.

- **Q**: Why not use GPG keys or other public key signing?
- **A**: This is complimentary to public key signing! Public key signing asserts that someone with access to the private key signed the exact content. But, the private key can be used to generate an unlimited number of signatures for different content. If the URLs contents are both signed and logged in the URL content record then there is a guarantee that both the owner of the private key signed the content AND the content being fetched is cryptographically identical to the content other people are fetching using btget.

- **Q**: Where does the name btget come from?
- **A**: The "bt" stands for binary transparency.

- **Q**: What are examples of practical attacks this could mitigate?
- **A**: A well known example is the Feb. 2016 attack on the Linux Mint project where an attacker replaced a version of a Linux Mint release with a new version that included a backdoor vulnerability. With luck this was detected and mitigated within a day, however, there are likely many projects that have been attacked in a similar way without catching the attack. Further, the project could not make a strong assurance to the community on how long they were vulnerable, only stating "As far as we know, the only compromised edition was Linux Mint 17.3 Cinnamon edition.". By ensuring the cryptographic digests of all releases end up in a publicly audited log the project could have stated exactly when the content changed and potentially used a Certificate Transparency monitor to get notified quickly once it happened.

- **Q**: What happens if a URLs contents are modified?
- **A**: URLs whos contents will change aren't a good use case for fetching via btget. The first contents the service sees for a URL will be the contents that are stored forever.
