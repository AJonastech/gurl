package main

import (
	"fmt"
	"net/url"
	"os"
)

//lets have a very basic function that validates the url input

func isValidUrl(rawURL string) bool{
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false 
	}

	return true
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr,"Error: Missing Required input")
		fmt.Fprintf(os.Stderr,"Usage: %s <url>\n", os.Args[0])
		os.Exit(1)
	}
    urlString := os.Args[1]
	if !isValidUrl(urlString){
      fmt.Fprint(os.Stderr, "please enter a valid url\n")
	  os.Exit(1)
	}

	
	fmt.Fprintf(os.Stdout,"Url %s\n", urlString)

}