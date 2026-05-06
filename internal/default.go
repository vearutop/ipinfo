// Package internal keeps reusable code.
package internal

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/vearutop/netrie"
)

// DefaultIdxs returns default IP lookuper indexes.
func DefaultIdxs(refresh bool, inMem bool) ([]netrie.IPLookuper, error) {
	var res []netrie.IPLookuper

	if env := os.Getenv("IPINFO_DEFAULT_DB"); env != "" {
		fns, err := filepath.Glob(env)
		if err != nil {
			return nil, err
		}

		for _, fn := range fns {
			idx, err := netrie.OpenFile(fn)
			if err != nil {
				return nil, err
			}

			if idx.Metadata().Name == "" {
				idx.Metadata().Name = strings.TrimSuffix(path.Base(fn), ".bin")
			}

			res = append(res, idx)
		}

		return res, nil
	}

	for _, dbURL := range []string{
		"https://github.com/vearutop/ipinfo/releases/download/index/cloud.bin.zst",
		"https://github.com/vearutop/ipinfo/releases/download/index/asn-bot.bin.zst",
		"https://github.com/vearutop/ipinfo/releases/download/index/asn-lite.bin.zst",
		"https://github.com/vearutop/ipinfo/releases/download/index/city-lite.bin.zst",
	} {
		tmpName := path.Join(os.TempDir(), strings.TrimSuffix(path.Base(dbURL), ".zst"))

		_, err := os.Stat(tmpName)
		if errors.Is(err, os.ErrNotExist) || refresh {
			log.Println("downloading", dbURL)

			req, err := http.NewRequest(http.MethodGet, dbURL, nil)
			if err != nil {
				return nil, err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}

			defer resp.Body.Close()

			log.Println("saving to", tmpName)

			f, err := os.Create(tmpName)
			if err != nil {
				return nil, err
			}

			var r io.Reader = resp.Body

			if strings.HasSuffix(dbURL, ".zst") {
				zr, err := zstd.NewReader(r)
				if err != nil {
					return nil, err
				}

				r = zr
			}

			if _, err := io.Copy(f, r); err != nil {
				return nil, err
			}

			if err := f.Close(); err != nil {
				return nil, err
			}

			if err := resp.Body.Close(); err != nil {
				return nil, err
			}
		}

		var idx netrie.IPLookuper

		if inMem {
			idx, err = netrie.LoadFromFile(tmpName)
		} else {
			idx, err = netrie.OpenFile(tmpName)
		}

		if err != nil {
			return nil, err
		}

		if idx.Metadata().Name == "" {
			idx.Metadata().Name = strings.TrimSuffix(strings.TrimSuffix(path.Base(dbURL), ".zst"), ".bin")
		}

		res = append(res, idx)
	}

	return res, nil
}
