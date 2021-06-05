package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const projects = "/api/v4/projects/"

type importResponse struct {
	ID                int       `json:"id"`
	Description       string    `json:"description"`
	Name              string    `json:"name"`
	NameWithNamespace string    `json:"name_with_namespace"`
	Path              string    `json:"path"`
	PathWithNamespace string    `json:"path_with_namespace"`
	CreatedAt         time.Time `json:"created_at"`
	ExportStatus      string    `json:"export_status"`
	Links             struct {
		APIURL string `json:"api_url"`
		WebURL string `json:"web_url"`
	} `json:"_links"`
}

type importStatusResponse struct {
	ID                int              `json:"id"`
	Description       string           `json:"description"`
	Name              string           `json:"name"`
	NameWithNamespace string           `json:"name_with_namespace"`
	Path              string           `json:"path"`
	PathWithNamespace string           `json:"path_with_namespace"`
	CreatedAt         time.Time        `json:"created_at"`
	ImportStatus      string           `json:"import_status"`
	CorrelationID     string           `json:"correlation_id"`
	FailedRelations   []FailedRelation `json:"failed_relations"`
}

type FailedRelation struct {
	ID               int       `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	ExceptionClass   string    `json:"exception_class"`
	ExceptionMessage string    `json:"exception_message"`
	Source           string    `json:"source"`
	RelationName     string    `json:"relation_name"`
}

func importFile(uri, token, namespace, path, filename string) importResponse {
	client := http.Client{}

	URL := uri + projects + "/import"

	// Open a file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Close the file on exit
	defer file.Close()

	body := &bytes.Buffer{}

	// Create the multipart writer
	writer := multipart.NewWriter(body)

	// Create a new form-data header with the provided field name and file name
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		log.Fatal(err)
	}

	// Copy the contents of the file directly in to part
	_, err = io.Copy(part, file)

	// Add additional fields to our writer
	if err = writer.WriteField("path", path); err != nil {
		log.Fatal(err)
	}

	// Only add the namespace field if it has been provided.
	if namespace != "" {
		if err = writer.WriteField("namespace", namespace); err != nil {
			log.Fatal(err)
		}
	}

	// Add the file
	if err = writer.WriteField("file", filename); err != nil {
		log.Fatal(err)
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new request
	req, err := http.NewRequest("POST", URL, body)
	// You must set the content type
	req.Header.Add("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var r importResponse
	json.Unmarshal(bs, &r)

	if debug {
		fmt.Println("Import Response:", string(bs))
		fmt.Println("Import Unmarshal:", r)
	}

	return r
}

func getImportStatus(uri, token, id string) importStatusResponse {
	client := http.Client{}

	URL := uri + projects + id + "/import"
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("PRIVATE-TOKEN", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var r importStatusResponse
	json.Unmarshal(bs, &r)

	if debug {
		fmt.Println("Status Response:", string(bs))
		fmt.Println("Status Unmarshal:", r)
	}

	return r
}
