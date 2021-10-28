**Project Archived**

This project is archived at the transparency log service at https://beta.transparencylog.net will be shutoff on November 9th.

If you are looking for an assest signing and verification system that is backed by a transparency log checkout the [sigstore/cosign](https://github.com/sigstore/cosign) project. In particular they have a subcommand called [sget](https://github.com/sigstore/cosign#blobs) that offers a sort of similar guarantee although not across any arbitrary URL.

# tl

[![Go Report Card](https://goreportcard.com/badge/github.com/transparencylog/tl)](https://goreportcard.com/report/github.com/transparencylog/tl)

**Beta Warning**: tl works as described but hasn't had a security review, and has minimal load testing. Please test the tool and provide feedback.

`tl` verifies the contents of URLs against a publicly recorded cryptographic log. `tl` is flexible and can download, print, or verify existing files.

The asset transparency log gives users of `tl` a number of useful properties:

- Verifiability of a downloaded URL's contents being identical to what the rest of the world sees
- Searchability of recorded content changes of a URL
- Notifications to any interested party about changes to the URLs contents

`tl` relies on a public webservice, hosted at beta-asset.transparencylog.net, which keeps an append only log of the cryptographic digests of all URLs it has seen. If a URL has not been seen before the service downloads the URL and stores a cryptographic digest of those contents in the log. The `tl` tool will then download the digest from the service and verify this digest matches the contents of the files the user retrieved.

## Installation

Download the appropriate release for macOS, Windows, and Linux from https://github.com/transparencylog/tl/releases and extract the archive.

Linux users can optionally use [.deb and .rpm](https://dl.equinox.io/transparencylog/tl/stable) packages.

macOS and Linux homebrew users can optionally use brew.

```
brew install transparencylog/tap/tl
```

## Example Usage

Use `tl` to download the v5.8 Linux source code and verify that the contents are publicly recorded.

```
./tl get https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.8.tar.xz
```

Or if you prefer to download using a familiar tool, say curl:

```
URL=https://cdn.kernel.org/pub/linux/kernel/v5.x/linux-5.8.tar.xz
FILE=linux-5.8.tar.xz

curl -L $URL -o $FILE
./tl verify $URL $FILE
```

`tl` also implements a cat subcommand for doing things like installing from a shell script:

```
tl cat https://raw.githubusercontent.com/Homebrew/install/fea1e80d/install.sh | bash
```

## Frequently Asked Questions (FAQ)

The [FAQ](https://www.transparencylog.com/frequently-asked-questions/)
covers topics like: Why not use blockchain? Does this replace GPG? What
attacks could this mitigate?
