package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "strconv"
    "encoding/json"
    "sync"
    "strings"
)

// Types and Globals

type PeerInfo struct {
    m  map[string]int
    mu sync.RWMutex
}

type Init struct {
    Port int
    PI map[string]int
}

var (
    myPeerInfo = &PeerInfo{m: make(map[string]int)} // Map of Peer Addresses to Permission Levels
    Port int = 12343
    GoPortList []int
    JSClientCount = 0
)

const (
    VIEWER int = 0
    EDITOR int = 1
    MASTER int = 2
)

// ONLY FOR TESTING
func ServeGo(ws *websocket.Conn) {
    Port += 2
    if _, err := ws.Write([]byte(strconv.Itoa(Port))); err != nil {
        Port -= 2
        log.Println("(ServeGo) WebSocket Write Error: ",err)
    } else {
        GoPortList = append(GoPortList,Port)
    }
}

// Serve Peer List
func ServeJS(ws *websocket.Conn) {
    if JSClientCount < len(GoPortList) {
        var port = GoPortList[JSClientCount]
        addr := strings.Join([]string{":",strconv.Itoa(port+1)},"")
        if len(myPeerInfo.m) == 0 {
            myPeerInfo.m[addr] = MASTER
        } else {
            myPeerInfo.m[addr] = VIEWER
        }
        init := Init{
            Port: port,
            PI: myPeerInfo.m,
        }
        e := json.NewEncoder(ws)
        err := e.Encode(init)
        if err != nil {
            delete(myPeerInfo.m,addr)
            log.Println("(ServeJS) JSON Error: ",err)
        } else {
            JSClientCount += 1
            fmt.Println("(ServeJS) Success")
            fmt.Printf("\tinit.Port: %s\n",init.Port)
            fmt.Printf("\tinit.PI: %s\n",init.PI)
        }
    }
}

func main() {
    fmt.Printf("(Main) Starting server at http://localhost:8080/\n")
    fmt.Printf("(Main) Issuing ports at ws://localhost:8080/ws\n")

    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go", websocket.Handler(ServeGo))
    http.Handle("/ws/js", websocket.Handler(ServeJS))

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("(Main) ListenAndServe: ", err)
    }
}