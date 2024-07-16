// credits to benpate for the source code: https://gist.github.com/benpate/f92b77ea9b3a8503541eb4b9eb515d8a
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

/***********************
This is a simple demonstration of how to use the built-in template package in Go to implement
"template fragments" as described here: https://htmx.org/essays/template-fragments/
Go accomplishes this with the {{block}} action (described here: https://pkg.go.dev/text/template)
which defines and executes a template fragment inline inside of another template.  You only have
to wire up your application to use the correct template name and the fragment will be executed.
************************/

var page *template.Template

// init function sets up the template+fragment.  Most of the work is actually done here.
// In a larger program, this would likely be stored in a separate file, but this makes for a
// simple example.
func init() {

	page = template.New("main")

	page = template.Must(page.Parse(`<!DOCTYPE html>
	
	<html>
	<head>
		<script src="https://unpkg.com/htmx.org@1.8.0"></script>
		<link rel="stylesheet" href="https://unpkg.com/missing.css@1.1.1"/>
		<title>Template Fragment Example</title>
	</head>
	<body>
		<h1>Template Fragment Example</h1>
		<div>{{.announcement}}</div>
		
		<p>This page demonstrates how to create and serve 
		<a href="https://htmx.org/essays/template-fragments/">template fragments</a> 
		using the <a href="https://pkg.go.dev/text/template">built-in template package</a> in Go.</p>
		
		<p>This is accomplished by using the "block" action in the template, which lets you
		define and execute a sub-template in a single step.</p>
		<!-- Here's the fragment.  We can target it by executing the "buttonOnly" template. -->
		{{block "buttonOnly" .}}
			<button hx-get="/?counter={{.next}}&template=buttonOnly" hx-swap="outerHTML">
				This Button Has Been Clicked {{.counter}} Times
			</button>
		{{end}}
	</body>
	</html>`))
}

// handleRequest does the work to execute the template (or fragment) and serve the result.
// It's mostly boilerplate, so don't get hung up on it.
func handleRequest(w http.ResponseWriter, r *http.Request) {

	// Collect state info to pass to the template
	templateName := r.URL.Query().Get("template")
	if templateName == "" {
		templateName = "main" // default value in case the query parameter is missing
	}

	// Pack state info into a map to pass to the template
	// data := make(map[string]int)
	// data["counter"] = counter
	// data["next"] = counter + 1
	data := make(map[string]string)
	if err, announcement := getLatestAnnouncement(); err == nil {
		data["announcement"] = announcement.Text
	}

	// Execute the template and handle errors
	if err := page.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// main is the entry point for the program. It sets up and executes the HTTP server.
func main() {
	getLatestAnnouncement()
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}

type CollectionList struct {
	Page int `json:"page"`
	PerPage int `json:"perPage"`
	TotalPages int `json:"totalPages"`
	TotalItems int `json:"totalItems"`
	Items []announcementRecord `json:"items"`
}

type announcementRecord struct {
	ID string `json:"id"`
	CollectionId string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
	Created string `json:"created"`
	Updated string `json:"updated"`
	Text string `json:"text"`
}

func getLatestAnnouncement() (error, announcementRecord) {
	var announcements CollectionList

    // make GET request to API to get user by ID
    apiUrl := "http://192.168.4.144:8080/api/collections/announcements/records?perPage=1&skipTotal=true&sort=-created"
    request, err := http.NewRequest("GET", apiUrl, nil)

    if err != nil {
        fmt.Println(err)
    }

    request.Header.Set("Content-Type", "application/json; charset=utf-8")

    client := &http.Client{}
    response, err := client.Do(request)

    if err != nil {
        fmt.Println(err)
    }

    responseBody, err := io.ReadAll(response.Body)

    if err != nil {
        fmt.Println(err)
    }

		er := json.Unmarshal(responseBody, &announcements)
		if er != nil {
			fmt.Println(er)
		}
    fmt.Println("Status: ", response.Status)

		if len(announcements.Items) == 0 {
			return errors.New("No announcements found"), announcements.Items[0]
		}

    // clean up memory after execution
   defer response.Body.Close()
	 return nil, announcements.Items[0]
}