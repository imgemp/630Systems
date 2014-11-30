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
var IDMap map[string]string // Map of Peer IDs to P2P Websockets

// ONLY FOR TESTING
// Issue a new ports (internal and external websockets) to the Go Client
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
        JSClientCount += 2
    }
}

func ReportPeers_Go(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Go Client: %s\n", msg[:n])
    // Retrieve Peers
    var Peer string
    if (IDList[0] == string(msg[:n])) {
        Peer = IDList[1]
    } else {
        Peer = IDList[0]
    }
    // Report Peer List to Go Client
    if _, err := ws.Write([]byte(Peer)); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Your peer's websocket is at port %s\n",Peer)
    }
}

func main() {
    fmt.Printf("Starting server at http://localhost:8080/\n")
    fmt.Printf("Issuing ports at ws://localhost:8080/ws\n")

    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go/init", websocket.Handler(IssuePort_Go))
    http.Handle("/ws/js", websocket.Handler(IssuePort_JS))
    http.Handle("/ws/go/peer", websocket.Handler(ReportPeers_Go))

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}