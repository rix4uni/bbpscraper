## bbpscraper
Scan Bug Bounty Programs using bbpscraper.

## Installation
```
go install github.com/rix4uni/bbpscraper@latest
```

## Download prebuilt binaries
```
wget https://github.com/rix4uni/bbpscraper/releases/download/v0.0.1/bbpscraper-linux-amd64-0.0.1.tgz
tar -xvzf bbpscraper-linux-amd64-0.0.1.tgz
rm -rf bbpscraper-linux-amd64-0.0.1.tgz
mv bbpscraper ~/go/bin/bbpscraper
```
Or download [binary release](https://github.com/rix4uni/bbpscraper/releases) for your platform.

## Compile from source
```
git clone --depth 1 github.com/rix4uni/bbpscraper.git
cd bbpscraper; go install
```

## Usage
```
Usage of bbpscraper:
  -mmc int
        Minimum number of regex matches to consider a valid result (default 2)
  -parallel int
        Number of concurrent domain scans (default 10)
  -path string
        File with list of paths to append
  -silent
        silent mode.
  -stop int
        Stop after N successful matches per domain (use 0 to check all paths) (default 1)
  -summary string
        File to write path match summary
  -timeout int
        Timeout per request in seconds (default 15)
  -verbose
        Enable verbose output for debugging
  -version
        Print the version of the tool and exit.
```

## Output Examples

Single URL:
```
echo "https://google.com" | bbpscraper -path bbp-paths.txt
```

Multiple URLs:
```
cat subs.txt | bbpscraper -path bbp-paths.txt
```

Large Scale Scan:
```
▶ Download domain list
wget -q https://github.com/rix4uni/top-1m-domains/raw/refs/heads/master/alldomains.txt.gz && gunzip alldomains.txt.gz

▶ Scan using httpx+bbpscraper
cat alldomains.txt | httpx -duc -nc -silent | bbpscraper -path bbp-paths.txt

▶ Scan using sed+bbpscraper, not accurate but you can save lot of time
cat alldomains.txt | sed 's/^/https:\/\//' | bbpscraper -path bbp-paths.txt
```

Output
```
https://github.com/bounty/bugbounty/reward.html ["bounty", "reward", "scope"]
https://apple.com/support/security ["bounty", "reward", "security@apple.com"]
```

Generate `bbp-paths.txt`
```
cat bbp-paths.txt | egrep -v "\." | sed -E 's|[^/]$|&/|' | unew -q t.txt
cat bbp-paths.txt | grep "\." | unew -q t.txt

rm -rf bbp-paths.txt && mv t.txt bbp-paths.txt
```