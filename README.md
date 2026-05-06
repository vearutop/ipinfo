# ipinfo

[![Build Status](https://github.com/vearutop/ipinfo/workflows/test-unit/badge.svg)](https://github.com/vearutop/ipinfo/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/vearutop/ipinfo/branch/master/graph/badge.svg)](https://codecov.io/gh/vearutop/ipinfo)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/vearutop/ipinfo)
[![Time Tracker](https://wakatime.com/badge/github/vearutop/ipinfo.svg)](https://wakatime.com/badge/github/vearutop/ipinfo)
![Code lines](https://sloc.xyz/github/vearutop/ipinfo/?category=code)
![Comments](https://sloc.xyz/github/vearutop/ipinfo/?category=comments)


## Install

```
go install github.com/vearutop/ipinfo@latest
$(go env GOPATH)/bin/ipinfo --help
```

Or download binary from [releases](https://github.com/vearutop/ipinfo/releases).

### Linux AMD64

```
wget https://github.com/vearutop/ipinfo/releases/latest/download/linux_amd64.tar.gz && tar xf linux_amd64.tar.gz && rm linux_amd64.tar.gz
./ipinfo -version
```

### Macos Intel

```
wget https://github.com/vearutop/ipinfo/releases/latest/download/darwin_amd64.tar.gz && tar xf darwin_amd64.tar.gz && rm darwin_amd64.tar.gz
codesign -s - ./ipinfo
./ipinfo -version
```

### Macos Apple Silicon (M1, etc...)

```
wget https://github.com/vearutop/ipinfo/releases/latest/download/darwin_arm64.tar.gz && tar xf darwin_arm64.tar.gz && rm darwin_arm64.tar.gz
codesign -s - ./ipinfo
./ipinfo -version
```


## Usage

```
Usage: ipinfo [-mmdb <mmdb>] [-disp-cloud-dir <dir>] [-output <file>] [...ip|host]
  -disp-cloud-dir string
        path to github.com/disposable/cloud-ip-ranges directory
  -mmdb string
        input MMDB (MaxMind GeoIP) file
  -mmdb-type string
        mmdb type: city, country, asn, anon, conn
  -netrie value
        path to netrie index file, multiple DBs can be provided with multiple flags
        by default found in IPINFO_DEFAULT_DB env var glob
        or downloaded from https://github.com/vearutop/ipinfo/releases/tag/index
  -output string
        output file for netrie built from mmdb/cloud index
  -refresh
        refresh local netrie index from github
```

You can query IP addresses.
```
ipinfo 43.173.173.55 45.89.66.149
2025/12/29 18:39:25 downloading https://github.com/vearutop/ipinfo/releases/download/index/cloud.bin.zst
2025/12/29 18:39:26 downloading https://github.com/vearutop/ipinfo/releases/download/index/asn-lite.bin.zst
2025/12/29 18:39:27 downloading https://github.com/vearutop/ipinfo/releases/download/index/city-lite.bin.zst
43.173.173.55
cloud: 
asn-lite: AS132203 Tencent Building, Kejizhongyi Avenue
city-lite: SG:Singapore

45.89.66.149
cloud: 
asn-lite: AS209641 I-servers Ltd
city-lite: RU:Moscow
```

Or hosts, or both.
```
ipinfo reddit.com
151.101.129.140
cloud: fastly
asn-lite: AS54113 FASTLY
city-lite: US:Unknown

151.101.65.140
cloud: fastly
asn-lite: AS54113 FASTLY
city-lite: US:Unknown
...
```

If you have GeoIP2 bases, you can index them 
```
ipinfo -mmdb GeoIP2-Connection-Type.mmdb -mmdb-type conn -output conn.bin
ipinfo -mmdb GeoIP2-Anonymous-IP.mmdb -mmdb-type anon -output anon.bin
ipinfo -mmdb GeoIP2-Country.mmdb -mmdb-type country -output country.bin
ipinfo -mmdb GeoIP2-ISP.mmdb -mmdb-type asn -output asn.bin
ipinfo -mmdb GeoIP2-City.mmdb -mmdb-type city -output cities.bin
```

Or resolve directly from MMDB without building a netrie index:

```
ipinfo -mmdb GeoIP2-Anonymous-IP.mmdb 194.36.25.11
{"ip":"194.36.25.11","is_anonymous":true,"is_anonymous_vpn":true,"is_hosting_provider":true}
```

and use by default with env var:

```
export IPINFO_DEFAULT_DB=/path/to/*.bin
```

```
ipinfo 43.173.173.55 45.89.66.149
43.173.173.55
GeoIP2 Anonymous IP: is_anonymous;is_hosting_provider
GeoIP2 ISP: AS132203 Tencent Building, Kejizhongyi Avenue
GeoIP2 City: SG:Singapore
GeoIP2 Connection Type: Corporate

45.89.66.149
GeoIP2 Anonymous IP: is_anonymous;is_hosting_provider
GeoIP2 ISP: AS209641 I-servers Ltd
GeoIP2 City: RU:Moscow
GeoIP2 Connection Type: Corporate
```
