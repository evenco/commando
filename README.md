Go CSV
=====

The Commando package aims to provide easy serialization and deserialization functions to use CSV in Golang

API and techniques inspired from https://godoc.org/gopkg.in/mgo.v2

[![GoDoc](https://godoc.org/github.com/gocarina/commando?status.png)](https://godoc.org/github.com/gocarina/commando)
[![Build Status](https://travis-ci.org/gocarina/commando.svg?branch=master)](https://travis-ci.org/gocarina/commando)

Installation
=====

```go get -u github.com/gocarina/commando```

Full example
=====

Consider the following CSV file

```csv

client_id,client_name,client_age
1,Jose,42
2,Daniel,26
3,Vincent,32

```

Easy binding in Go!
---

```go

package main

import (
	"fmt"
	"os"

	"github.com/gocarina/commando"
)

type Client struct { // Our example struct, you can use "-" to ignore a field
	Id      string `csv:"client_id"`
	Name    string `csv:"client_name"`
	Age     string `csv:"client_age"`
	NotUsed string `csv:"-"`
}

func main() {
	clientsFile, err := os.OpenFile("clients.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer clientsFile.Close()

	clients := []*Client{}

	if err := commando.UnmarshalFile(clientsFile, &clients); err != nil { // Load clients from file
		panic(err)
	}
	for _, client := range clients {
		fmt.Println("Hello", client.Name)
	}

	if _, err := clientsFile.Seek(0, 0); err != nil { // Go to the start of the file
		panic(err)
	}

	clients = append(clients, &Client{Id: "12", Name: "John", Age: "21"}) // Add clients
	clients = append(clients, &Client{Id: "13", Name: "Fred"})
	clients = append(clients, &Client{Id: "14", Name: "James", Age: "32"})
	clients = append(clients, &Client{Id: "15", Name: "Danny"})
	csvContent, err := commando.MarshalString(&clients) // Get all clients as CSV string
	//err = commando.MarshalFile(&clients, clientsFile) // Use this to save the CSV back to the file
	if err != nil {
		panic(err)
	}
	fmt.Println(csvContent) // Display all clients as CSV string

}

```

Customizable Converters
---

```go

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format("20060201"), nil
}

// You could also use the standard Stringer interface 
func (date *DateTime) String() (string) {
	return date.String() // Redundant, just for example
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("20060201", csv)
	return err
}

type Client struct { // Our example struct with a custom type (DateTime)
	Id       string   `csv:"id"`
	Name     string   `csv:"name"`
	Employed DateTime `csv:"employed"`
}

```

Customizable CSV Reader / Writer
---

```go

func main() {
        ...
	
        commando.SetCSVReader(func(in io.Reader) commando.CSVReader {
            r := csv.NewReader(in)
            r.Comma = '|'
            return r // Allows use pipe as delimiter
        })	
	
        ...
	
        commando.SetCSVReader(func(in io.Reader) commando.CSVReader {
            r := csv.NewReader(in)
            r.LazyQuotes = true
            r.Comma = '.'
            return r // Allows use dot as delimiter and use quotes in CSV
        })
	
        ...
	
        commando.SetCSVReader(func(in io.Reader) commando.CSVReader {
            //return csv.NewReader(in)
            return commando.LazyCSVReader(in) // Allows use of quotes in CSV
        })

        ...

        commando.UnmarshalFile(file, &clients)

        ...

        commando.SetCSVWriter(func(out io.Writer) *SafeCSVWriter {
            writer := csv.NewWriter(out)
            writer.Comma = '|'
            return commando.NewSafeCSVWriter(writer)
        })

        ...

        commando.MarshalFile(&clients, file)

        ...
}

```
