package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "strconv"
)

// Peer Map
var IDList []string
var IDMap = make(map[string]string) // Map of Peer IDs to P2P Websockets

// ONLY FOR TESTING
// Issue new ports (internal and external websockets) to the Go Client
var Port int = 12343
var GoPortList []int
var JSClientCount = 0
func IssuePort_Go(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Go Client: %s\n", msg[:n])
    // Issue New Port # to Go Client
    Port += 2
    GoPortList = append(GoPortList,Port)
    if _, err := ws.Write([]byte(strconv.Itoa(Port))); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Use ports %s and %s\n",strconv.Itoa(Port),strconv.Itoa(Port+1))
    }
}

// Issue new ports (internal and external websockets) to the JS Client
func IssuePort_JS(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("JS Client: %s\n", msg[:n])
    // Issue New Port # to JS Client
    if JSClientCount < len(GoPortList) {
        if _, err := ws.Write([]byte(strconv.Itoa(GoPortList[JSClientCount]))); err != nil {
            log.Fatal(err)
        } else {
            fmt.Printf("Use ports %s and %s\n",strconv.Itoa(GoPortList[JSClientCount]),strconv.Itoa(GoPortList[JSClientCount]+1))
            IDMap[strconv.Itoa(GoPortList[JSClientCount])] = strconv.Itoa(GoPortList[JSClientCount]+1)
        }
        JSClientCount += 1
    }
}

func main() {
    fmt.Printf("Starting server at http://localhost:8080/\n")
    fmt.Printf("Issuing ports at ws://localhost:8080/ws\n")

    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go", websocket.Handler(IssuePort_Go))
    http.Handle("/ws/js", websocket.Handler(IssuePort_JS))

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}