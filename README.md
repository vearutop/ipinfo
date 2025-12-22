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

