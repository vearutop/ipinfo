// Package main is an example CLI app.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vearutop/ipinfo/cloud"
	"github.com/vearutop/ipinfo/internal"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/mmdb"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout io.Writer) error {
	var idxs []netrie.IPLookuper

	fs := flag.NewFlagSet("ipinfo", flag.ContinueOnError)
	fs.SetOutput(stdout)

	mmDB := fs.String("mmdb", "", "input MMDB (MaxMind GeoIP) file")
	mmDBType := fs.String("mmdb-type", "", "mmdb type: city, country, asn, anon, conn")
	dispDir := fs.String("disp-cloud-dir", "", "path to github.com/disposable/cloud-ip-ranges directory")
	refresh := fs.Bool("refresh", false, "refresh local netrie index from github")
	fs.Func("netrie", "path to netrie index file, multiple DBs can be provided with multiple flags\n"+
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

	output := fs.String("output", "", "output file for netrie built from mmdb/cloud index")

	if err := fs.Parse(args); err != nil {
		return err
	}

	ips := fs.Args()

	if len(ips) == 0 && *mmDB == "" && *dispDir == "" && *output == "" {
		fmt.Fprintln(stdout, "Usage: ipinfo [-mmdb <mmdb>] [-disp-cloud-dir <dir>] [-output <file>] [...ip|host]")
		fs.PrintDefaults()

		return nil
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
			return fmt.Errorf("build index: %w", err)
		}

		log.Println("nets: ", tr.Len(), "names: ", tr.LenNames(), "nodes: ", tr.LenNodes())
		log.Println("minimizing...")
		tr.Minimize()
		log.Println("nodes: ", tr.LenNodes())

		err = tr.SaveToFile(*output)
		if err != nil {
			return fmt.Errorf("save file: %w", err)
		}

		return nil
	}

	if *mmDB != "" {
		if len(ips) == 0 {
			return errors.New("at least one ip or host is required when using -mmdb without -output")
		}

		return resolveMMDB(ips, *mmDB, stdout)
	}

	if *dispDir != "" {
		if *output == "" {
			return errors.New("output file is required when using -disp-cloud-dir")
		}

		tr := netrie.NewCIDRIndex()

		log.Println("loading cloud networks...")

		if err := cloud.LoadCloudLocal(tr, *dispDir); err != nil {
			return err
		}

		tr.Minimize()
		log.Println("nets:", tr.Len())
		log.Println("names:", tr.LenNames())

		err := tr.SaveToFile(*output)
		if err != nil {
			return err
		}

		return nil
	}

	if len(idxs) == 0 {
		dflt, err := internal.DefaultIdxs(*refresh, false)
		if err != nil {
			return err
		}

		idxs = dflt
	}

	return resolveIPs(ips, idxs, stdout)
}

func resolveIPs(ips []string, idxs []netrie.IPLookuper, stdout io.Writer) error {
	newLine := false
	seen := make(map[string]bool)

	for _, ip := range ips {
		ips, err := net.LookupIP(ip)
		if err != nil {
			return err
		}

		for _, ip := range ips {
			if seen[ip.String()] {
				continue
			}

			seen[ip.String()] = true

			if newLine {
				fmt.Fprintln(stdout)
			}

			fmt.Fprintln(stdout, ip.String())

			for _, idx := range idxs {
				fmt.Fprintf(stdout, "%s: %s\n", idx.Metadata().Name, idx.LookupIP(ip))
			}

			newLine = true
		}
	}

	return nil
}

func resolveMMDB(inputs []string, mmdbPath string, stdout io.Writer) error {
	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = db.Close()
	}()

	enc := json.NewEncoder(stdout)

	for _, input := range inputs {
		ips, err := net.LookupIP(input)
		if err != nil {
			return err
		}

		for _, ip := range ips {
			var rec map[string]any
			if err := db.Lookup(ip, &rec); err != nil {
				return err
			}

			out := make(map[string]any, len(rec)+1)
			out["ip"] = ip.String()

			for k, v := range rec {
				out[k] = v
			}

			if err := enc.Encode(out); err != nil {
				return err
			}
		}
	}

	return nil
}
