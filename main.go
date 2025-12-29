// Package main is an example CLI app.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vearutop/ipinfo/cloud"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/mmdb"
)

func main() {
	mmDB := flag.String("mmdb", "", "input MMDB (MaxMind GeoIP) file")
	mmDBType := flag.String("mmdb-type", "", "mmdb type: city, country, asn, anon")
	dispDir := flag.String("disp-cloud-dir", "", "path to github.com/disposable/cloud-ip-ranges directory")
	netIdx := flag.String("netrie", "", "path to netrie index file")
	output := flag.String("output", "", "output file for netrie built from mmdb/cloud index")

	flag.Parse()

	ip := flag.Arg(0)

	if ip == "" && *mmDB == "" && *dispDir == "" && *output == "" {
		fmt.Println("Usage: ipinfo [-mmdb <mmdb>] [-disp-cloud-dir <dir>] [-output <file>] [<ip>]")
		flag.PrintDefaults()

		return
	}

	if *mmDB != "" && *output != "" {
		var nameOpt func(o *mmdb.Options)

		switch *mmDBType {
		case "city":
			nameOpt = mmdb.CityCountryISOCode
		case "country":
			nameOpt = mmdb.CountryISOCode
		case "asn":
			nameOpt = mmdb.ASNOrg
		case "anon":
			nameOpt = mmdb.AnonymousIP
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

		err := mmdb.Load(tr, *mmDB, func(o *mmdb.Options) {
			//o.PrintProgress = true

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

		if err := cloud.LoadCloudLocal(tr, *dispDir); err != nil {
			log.Fatal(err)
		}

		tr.Minimize()
		println("nets:", tr.Len())
		println("names:", tr.LenNames())

		err := tr.SaveToFile(*output)
		if err != nil {
			log.Fatal(err)
		}

		return
	}

	var tr netrie.SafeIPLookuper

	if *netIdx != "" {
		idx, err := netrie.OpenFile(*netIdx)
		if err != nil {
			log.Fatal(err)
		}
		defer idx.Close()

		tr = idx
	} else {
		log.Fatal("netrie index file is required")
	}

	fmt.Println(tr.Lookup(ip))
}
