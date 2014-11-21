package main

import (
    "fmt"
    "net/http"
)

func main() {
	fmt.Printf("Starting server at http://localhost:8080/\n")
    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.ListenAndServe(":8080", nil)
}