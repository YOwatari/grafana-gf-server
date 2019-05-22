package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ziutek/rrd"
)

var config Config
var graphs []Graph

type Config struct {
	ListenAddr string
	ListenPort int
	DBFilePath string
}

type Graph struct {
	RRDFileName string
	ServiceName string
	SectionName string
}

func respondJSON(w http.ResponseWriter, result interface{}) {
	j, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,HEAD,OPTIONS")

	if _, err := w.Write([]byte(j)); err != nil {
		log.Println(err)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	var result []string
	for _, g := range graphs {
		f := fmt.Sprintf("%s/%s", path.Dir(config.DBFilePath), g.RRDFileName)
		if _, err := os.Stat(f); err != nil {
			log.Println(err)
		}

		res, err := rrd.Info(f)
		if err != nil {
			log.Println(err)
		}

		for ds, _ := range res["ds.index"].(map[string]interface{}) {
			result = append(result, f+":"+ds)
		}
	}

	respondJSON(w, result)
}

func getGraphs(filename string) ([]Graph, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println(err)
		}
	}()

	rows, err := db.Query(`SELECT * FROM graphs`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println(err)
		}
	}()

	got := make([]Graph, 0)
	for rows.Next() {
		var id string
		var r Graph

		err := rows.Scan(id, &r.ServiceName, &r.SectionName)
		if err != nil {
			return nil, err
		}
		r.RRDFileName = fmt.Sprintf("%x.rrd", md5.Sum([]byte(id)))
		got = append(got, r)
	}

	return got, nil
}

func main() {
	var err error

	flag.StringVar(&config.ListenAddr, "i", "0.0.0.0", "")
	flag.IntVar(&config.ListenPort, "p", 9000, "")
	flag.StringVar(&config.DBFilePath, "db", "./gforecast.db", "")

	graphs, err = getGraphs(config.DBFilePath)

	http.HandleFunc("/search", search)

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.ListenAddr, config.ListenPort), nil)
	if err != nil {
		log.Fatalln(err)
	}
}
