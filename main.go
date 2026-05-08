package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)



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

	res, err := makeRequest(urlString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request to %s: %v\n", urlString, err)
		os.Exit(1)
	}
defer res.Body.Close()

	fmt.Fprintf(os.Stdout,"Url %s\n", urlString)
	fmt.Fprintf(os.Stdout,"Response status code: %d\n", res.StatusCode)
}


//This is a very basic function that validates the url input

func isValidUrl(rawURL string) bool{
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false 
	}

	return true
}

//now that i have the input and validation layer working, i will work on the function to make the request to the url

func makeRequest(urlString string)(response *http.Response, err error){
	
	res, err := http.Get(urlString)

	if err != nil{
		return nil, err
	}

	return res, nil
}