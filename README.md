Commando, a CSV library for Go
==============================

The Commando package aims to provide easy marshalling and unmarshalling of CSV in Go.  It’s a fork of [gocarina/gocsv](https://github.com/gocarina/gocsv), with a simplified API.

Installation
=====

```go get -u github.com/evenco/commando```

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

	"github.com/evenco/commando"
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

    // Create an Unmarshaller which deserializes into *Client{}
    unmarshaller, err := commando.NewUnmarshaller(&Client{}, csv.NewReader(clientsFile))
    if err != nil {
        panic(err)
    }

    // Read everything, accumulating in clients
    err = commando.ReadAllCallback(um, func(record interface{}) error {
        clients = append(clients, record.(*Client))
        return nil
    })

	if err != nil {
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

    // Initialize Marshaller for *Client{} structs
    marshaller, err := commando.NewMarshaller(&Client{}, csv.NewWriter(clientsFile))
    if err != nil {
		panic(err)
	}

    for _, client := range clients {
        if err := marshaller.Write(client); err != nil {
            panic(err)
        }
    }
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
