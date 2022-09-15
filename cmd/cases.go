package cmd

import (
	"fmt"
	"io/ioutil"
	"mnezerka/geonet/store"
	"os"
	"path/filepath"
	"text/template"

	"github.com/apex/log"
	"github.com/ptrv/go-gpx"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

var cfgCasesDir string

type Case struct {
	Seq         int
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Dir         string
	Gpx         []string `yaml:"gpx"`
	Js          []string
}

var casesCmd = &cobra.Command{
	Use:   "cases",
	Short: "Generate test cases",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		log.Infof("generating cases from dir %s", cfgCasesDir)

		// ensure public directory exists
		publicDir := filepath.Join(".", "public")
		err := os.MkdirAll(publicDir, os.ModePerm)
		if err != nil {
			return err
		}

		cases, err := generateCases(cfgCasesDir)
		if err != nil {
			return err
		}

		log.Infof("Cases: %v", cases[0].Js)

		err = cases2Html(cases, publicDir)
		if err != nil {
			return err
		}

		return nil
	},
}

func generateCases(dir string) ([]*Case, error) {

	var result []*Case

	if _, err := os.Stat(cfgCasesDir); os.IsNotExist(err) {
		return result, fmt.Errorf("directory %s does not exist", cfgCasesDir)
	}

	dirs, err := ioutil.ReadDir(dir)
	if err != nil {
		return result, err
	}
	for caseIx, caseDir := range dirs {
		if caseDir.IsDir() {
			log.Infof("case dir: %s", caseDir.Name())
			items, err := ioutil.ReadDir(dir + "/" + caseDir.Name())
			if err != nil {
				return result, err
			}
			log.Infof("case dir with %d items", len(items))

			// loop through case directory content
			for _, item := range items {
				if item.Name() == "case.yaml" || item.Name() == "case.yml" {

					// filepath to absolute
					filename, err := filepath.Abs(dir + "/" + caseDir.Name() + "/" + item.Name())
					if err != nil {
						return result, err
					}

					// read file content
					yamlFile, err := ioutil.ReadFile(filename)
					if err != nil {
						return result, err
					}

					// parse case metadata from yaml file
					var c Case
					c.Seq = caseIx
					c.Dir = dir + "/" + caseDir.Name()
					err = yaml.Unmarshal(yamlFile, &c)
					if err != nil {
						return result, err
					}

					err = generateCase(&c)
					if err != nil {
						return result, err
					}

					result = append(result, &c)
				}
			}
		}
	}
	return result, nil
}

func generateCase(c *Case) error {

	log.Infof("Case %s", c.Title)

	store := store.NewStore()

	for file_ix := 0; file_ix < len(c.Gpx); file_ix++ {

		gpx, err := gpx.ParseFile(c.Dir + "/" + c.Gpx[file_ix])
		if err != nil {
			log.WithError(err)
			return err
		}

		store.AddGpx(gpx, 0.0002)

		// visualize add polyline
		c.Js = append(c.Js, GpxToPolyline(gpx, file_ix+1, c.Seq))
	}

	// visualize hulls
	c.Js = append(c.Js, StoreHullsToPolygons(store, c.Seq))

	// visualize lines
	c.Js = append(c.Js, StoreLinesToPolylines(store, c.Seq))

	log.Infof("js items: %d", len(c.Js))

	return nil
}

func cases2Html(cases []*Case, publicDir string) error {

	// read all case related templates
	t, err := template.ParseGlob("cases/tpl*.html")
	if err != nil {
		log.WithError(err)
		return err
	}

	// create output file
	f, err := os.Create(publicDir + "/index.html")
	if err != nil {
		log.WithError(err)
		return err
	}

	// process templates
	err = t.ExecuteTemplate(f, "tpl_cases.html", cases)
	if err != nil {
		log.WithError(err)
		return err
	}

	return nil
}

func init() {
	casesCmd.PersistentFlags().StringVar(&cfgCasesDir, "cases-dir", "./cases", "cases directory)")
	rootCmd.AddCommand(casesCmd)
}
