package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "strconv"
    // "time"
    "encoding/json"
    "sync"
)

const VIEWER int = 0
const EDITOR int = 1
const MASTER int = 2

// Peer Map
var myPeerInfo = &PeerInfo{m: make(map[string]int)} // Map of Peer Addresses to Permission Levels

type PeerInfo struct {
    m  map[string]int
    mu sync.RWMutex
}

type Init struct {
    Port int
    PI map[string]int
}

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
// func IssuePort_JS(ws *websocket.Conn) {
//     // Receive Message
//     var msg = make([]byte, 512)
//     var n int
//     var err error
//     if n, err = ws.Read(msg); err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("JS Client: %s\n", msg[:n])
//     // Issue New Port # to JS Client
//     if JSClientCount < len(GoPortList) {
//         if _, err := ws.Write([]byte(strconv.Itoa(GoPortList[JSClientCount]))); err != nil {
//             log.Fatal(err)
//         } else {
//             fmt.Printf("Use ports %s and %s\n",strconv.Itoa(GoPortList[JSClientCount]),strconv.Itoa(GoPortList[JSClientCount]+1))
//         }
//         JSClientCount += 1
//     }
// }

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
        init := Init{
            Port: GoPortList[JSClientCount],
            PI: myPeerInfo.m,
        }
        e := json.NewEncoder(ws)
        err := e.Encode(init)
        if err != nil {
            log.Fatal(err)
        } else {
            fmt.Println("Sent Init To JS Client")
            fmt.Printf("\tinit.Port: %s\n",init.Port)
            fmt.Printf("\tinit.PI: %s\n",init.PI)
        }
        JSClientCount += 1
    }
}

// Provide set of peer addresses
func IssuePeerSet_Go(ws *websocket.Conn) {
    // Receive Message - Peer's Socket Address
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Go Client: %s\n", msg[:n])
    // Add Peer's Address to Peer Set (if len(map) == 0, master, else, viewer by default)
    if len(myPeerInfo.m) == 0 {
        fmt.Printf("Issue as Master\n")
        myPeerInfo.m[string(msg[:n])] = MASTER
    } else {
        myPeerInfo.m[string(msg[:n])] = VIEWER
    }
    // Issue Peer Set to Go Client
    e := json.NewEncoder(ws)
    err = e.Encode(myPeerInfo.m)
    if err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Peer Set Sent\n")
        for key, value := range myPeerInfo.m {
            fmt.Println("Key:", key, "Value:", value)
        }
    }
}

func main() {
    fmt.Printf("Starting server at http://localhost:8080/\n")
    fmt.Printf("Issuing ports at ws://localhost:8080/ws\n")

    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go", websocket.Handler(IssuePort_Go))
    http.Handle("/ws/js", websocket.Handler(IssuePort_JS))
    http.Handle("/ws/go/peer", websocket.Handler(IssuePeerSet_Go))

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}