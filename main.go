package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage:\n\n\t%s <ldap URL>\n", os.Args[0])
		os.Exit(1)
	}
	output := deobfuscateLog4J(os.Args[1])
	grabLDAP(output)
}

func grabLDAP(ldapURL string) {
	l, err := ldap.DialURL(ldapURL)

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	baseDN := "dc=example,dc=com"
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=*))", // The filter to apply
		[]string{},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	var javacodebase, javafactory string

	for _, entry := range sr.Entries {
		fmt.Printf("Attributes:\n")
		for _, attr := range entry.Attributes {
			fmt.Printf("%s\t=\t%s\n", attr.Name, attr.Values)

			if len(attr.Values) > 0 {
				if attr.Name == "javaCodeBase" {
					javacodebase = attr.Values[0]
				} else if attr.Name == "javaFactory" {
					javafactory = attr.Values[0]
				}
			}
		}

		if len(javacodebase) > 0 && len(javafactory) > 0 {

			if !strings.HasSuffix(javacodebase, "/") {
				javacodebase = javacodebase + "/"
			}
			DownloadFile(javacodebase + javafactory + ".class")
		}

	}
}

func DownloadFile(origurl string) {

	fileURL, err := url.Parse(origurl)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(origurl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d\n", fileName, size)

}

func deobfuscateLog4J(input string) string {

	fmt.Println("Before: ", input)
	input = strings.TrimSpace(input)

	for {
		aidx := strings.LastIndex(input, "$")

		if aidx > 0 {
			bidx := strings.Index(input[aidx:], "}")
			if bidx > 0 {
				input = input[:aidx] + input[aidx+bidx-1:aidx+bidx] + input[aidx+bidx+1:]
			} else {
				break
			}
		} else if aidx == 0 && len(input) > 3 {
			input = input[2 : len(input)-1]
			break
		} else {
			break
		}
	}

	input = strings.TrimPrefix(input, "jndi:")

	fmt.Println("After:  ", input)
	return input
}
