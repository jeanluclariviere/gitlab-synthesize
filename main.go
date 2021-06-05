package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

var debug bool

func main() {

	uri := flag.String("uri", "", "Gitlab URI")
	token := flag.String("token", "", "Gitlab API token")
	group := flag.String("group", "sample-data", "Destination group")
	name := flag.String("name", "sample-data", "Project name")
	filename := flag.String("filename", "sample.tar.gz", "Filename of exported project")
	count := flag.Int("count", 1, "Number of times to import")
	flag.BoolVar(&debug, "debug", false, "Debug logging")
	flag.Parse()

	for i := 0; i < *count; i++ {
		importAndWait(*uri, *token, *group, *name+"-"+strconv.Itoa(i), *filename)
	}
}

func importAndWait(uri, token, namespace, path, filename string) {
	r1 := importFile(uri, token, namespace, path, filename)

	var imported bool = false
	var count int

	for !imported || count == 100 {
		r2 := getImportStatus(uri, token, strconv.Itoa(r1.ID))
		switch r2.ImportStatus {
		case "failed", "finished":
			fmt.Println("Import status:", r2.ImportStatus)
			imported = true
		case "nil":
			fmt.Println("Import failed, empty response")
			imported = true
		default:
			fmt.Println("Import status:", r2.ImportStatus)
			time.Sleep(time.Duration(5) * time.Second)
			count++
		}
	}
}
