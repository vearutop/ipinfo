// Package main is an example CLI app.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/vearutop/ipinfo/cloud"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/mmdb"
)

func main() {
	var idxs []netrie.SafeIPLookuper

	mmDB := flag.String("mmdb", "", "input MMDB (MaxMind GeoIP) file")
	mmDBType := flag.String("mmdb-type", "", "mmdb type: city, country, asn, anon, conn")
	dispDir := flag.String("disp-cloud-dir", "", "path to github.com/disposable/cloud-ip-ranges directory")
	refresh := flag.Bool("refresh", false, "refresh local netrie index from github")
	flag.Func("netrie", "path to netrie index file, multiple DBs can be provided with multiple flags\n"+
		"by default found in IPINFO_DEFAULT_DB env var glob\n"+
		"or downloaded from https://github.com/vearutop/ipinfo/releases/tag/index",
		func(s string) error {
			idx, err := netrie.OpenFile(s)
			if err != nil {
				return err
			}

			idxs = append(idxs, idx)

			return nil
		})

	output := flag.String("output", "", "output file for netrie built from mmdb/cloud index")

	flag.Parse()

	ips := flag.Args()

	if len(ips) == 0 && *mmDB == "" && *dispDir == "" && *output == "" {
		fmt.Println("Usage: ipinfo [-mmdb <mmdb>] [-disp-cloud-dir <dir>] [-output <file>] [...ip|host]")
		flag.PrintDefaults()

		return
	}

	if *mmDB != "" && *output != "" {
		var nameOpt func(o *mmdb.Options)

		switch *mmDBType {
		case "city":
			nameOpt = mmdb.CityCountryISOCode
		case "city-loc":
			nameOpt = mmdb.CityCountryISOCodeLoc
		case "country":
			nameOpt = mmdb.CountryISOCode
		case "asn":
			nameOpt = mmdb.ASNOrg
		case "anon":
			nameOpt = mmdb.AnonymousIP
		case "conn":
			nameOpt = mmdb.ConnectionType
		}

		tr := netrie.NewCIDRLargeIndex()

		if strings.HasPrefix(*mmDB, "https://") {
			println("downloading MMDB:", *mmDB)

			resp, err := http.Get(*mmDB)
			if err != nil {
				log.Fatal("downloading MMDB:", err)
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					log.Println("closing response body:", err)
				}
			}()

			f, err := os.CreateTemp("", "mmdb")
			if err != nil {
				log.Fatal("creating temp file:", err)
			}

			if _, err := io.Copy(f, resp.Body); err != nil {
				log.Fatal("copying response body:", err)
			}

			if err := f.Close(); err != nil {
				log.Fatal("closing temp file:", err)
			}

			println("downloaded MMDB:", f.Name())
			*mmDB = f.Name()

			defer os.Remove(f.Name())
		}

		log.Println("building index...")

		err := mmdb.Load(tr, *mmDB, func(o *mmdb.Options) {
			if nameOpt != nil {
				nameOpt(o)
			}
		})
		if err != nil {
			log.Fatal("build index:", err)
		}

		log.Println("nets: ", tr.Len(), "names: ", tr.LenNames(), "nodes: ", tr.LenNodes())
		log.Println("minimizing...")
		tr.Minimize()
		log.Println("nodes: ", tr.LenNodes())

		err = tr.SaveToFile(*output)
		if err != nil {
			log.Fatal("save file:", err)
		}

		return
	}

	if *dispDir != "" {
		if *output == "" {
			log.Fatal("output file is required when using -disp-cloud-dir")
		}

		tr := netrie.NewCIDRIndex()

		log.Println("loading cloud networks...")

		if err := cloud.LoadCloudLocal(tr, *dispDir); err != nil {
			log.Fatal(err)
		}

		tr.Minimize()
		log.Println("nets:", tr.Len())
		log.Println("names:", tr.LenNames())

		err := tr.SaveToFile(*output)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	if len(idxs) == 0 {
		dflt, err := defaultIdxs(*refresh)
		if err != nil {
			log.Fatal(err)
		}

		idxs = dflt
	}

	newLine := false
	seen := make(map[string]bool)

	for _, ip := range ips {
		ips, err := net.LookupIP(ip)
		if err != nil {
			log.Fatal(err)
		}

		for _, ip := range ips {
			if seen[ip.String()] {
				continue
			}

			seen[ip.String()] = true

			if newLine {
				fmt.Println()
			}

			fmt.Println(ip.String())

			for _, idx := range idxs {
				fmt.Printf("%s: %s\n", idx.Metadata().Name, idx.LookupIP(ip))
			}

			newLine = true
		}
	}
}

func defaultIdxs(refresh bool) ([]netrie.SafeIPLookuper, error) {
	var res []netrie.SafeIPLookuper

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

		idx, err := netrie.OpenFile(tmpName)
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
