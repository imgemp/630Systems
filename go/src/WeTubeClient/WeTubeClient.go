package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "golang.org/x/net/websocket"
    "strings"
    "strconv"
    "encoding/json"
    "crypto/rand"
    "sync"
)

// Types and Globals

type PeerInfo struct {
    m  map[string]int
    mu sync.RWMutex
}

type PeerChans struct {
    m  map[string]chan PeerMessage
    mu sync.RWMutex
}

type Command struct {
    Action string
    Argument string
    Target string
}

type PeerMessage struct {
    ID   string
    Addr string
    Body Command
    PI map[string]int
}

type ClientMessage struct {
    Body Command
    PI map[string]int
}

var (
    cws_addr string
    psoc_addr string

    myPeerInfo = &PeerInfo{m: make(map[string]int)} // Map of Peer Addresses to Permission Levels
    myPeerChans = &PeerChans{m: make(map[string]chan PeerMessage)} // Map of Peer Addresses to Channels

    out chan PeerMessage
    in chan ClientMessage
    blockExit chan string
)

const (
    VIEWER int = 0
    EDITOR int = 1
    MASTER int = 2
)

// ONLY FOR LOCAL TESTING
func RetrieveSockets() {
    url := "ws://localhost:8080/ws/go"
    tempWS, err := websocket.Dial(url, "", "http://localhost/")
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve Client Port from Websocket
    var msg = make([]byte, 512)
    var n int
    if n, err = tempWS.Read(msg); err != nil {
        log.Fatal(err)
    }

    // Construct Client and P2P Websocket Addresses
    clientPort, err := strconv.Atoi(string(msg[:n]))
    if err != nil {
        log.Fatal(err)
    }
    cws_addr = strings.Join([]string{":",strconv.Itoa(clientPort)},"");
    psoc_addr = strings.Join([]string{":",strconv.Itoa(clientPort+1)},"");
}

// Serve JS Client WebSocket Connection
func ServeClient(ws *websocket.Conn) {
    var reason string = "None"
    go ReceiveFromClient(ws)
    go SendToClient(ws)
    reason = <-blockExit
    fmt.Printf("(ServeClient) Closing Client WebSocket Connection: %s\n",reason)
}

func ReceiveFromClient(ws *websocket.Conn) {
    d := json.NewDecoder(ws)
    for {
        var cmsg ClientMessage
        err := d.Decode(&cmsg)
        if err != nil {
            log.Fatal(err)
        } else {
            UpdatePeers(cmsg.PI)
            cmd := cmsg.Body
            pmsg := PeerMessage{
                ID: RandomID(),
                Addr: psoc_addr,
                Body: cmd,
                PI: myPeerInfo.m,
            }
            out <- pmsg
            fmt.Println("(ReceiveFromClient) Success")
            fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
            fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
            fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
            fmt.Printf("\tPI: %s\n",cmsg.PI)
        }
    }
}

func SendToClient(ws *websocket.Conn) {
    for {
        cmsg := <-in
        e := json.NewEncoder(ws)
        err := e.Encode(cmsg)
        if err != nil {
            log.Println("(SendToClient) JSON Error: ",err)
        } else {
            cmd := cmsg.Body
            fmt.Println("(SendToClient) Success")
            fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
            fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
            fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
            fmt.Printf("\tPI: %s\n",cmsg.PI)
        }
    }
}

// Union Input Peer Map with Local Peer Info & Channels
func UpdatePeers(PI map[string]int) {
    for addr, rank := range PI {
        if _, ok := myPeerChans.m[addr]; !ok {
            myPeerChans.mu.Lock()
            myPeerChans.m[addr] = make(chan PeerMessage)
            myPeerChans.mu.Unlock()
        }
        if _, ok := myPeerInfo.m[addr]; !ok {
            myPeerInfo.mu.Lock()
            myPeerInfo.m[addr] = rank
            myPeerInfo.mu.Unlock()
        }
    }
}

// Serve P2P Socket Connection
func ServePeers(l net.Listener) {
    go ReceiveFromPeers(l)
    go SendToPeers()
}

func ReceiveFromPeers(l net.Listener) {
    for {
        c, err := l.Accept()
        if err != nil {
            log.Println("(ReceiveFromPeers) Listener Error: <", c.RemoteAddr(), "> ",err)
        }
        if pmsg, ok := ReceivePeerMessage(c); ok {
            cmd := Command{
                Action: pmsg.Body.Action,
                Argument: pmsg.Body.Argument,
                Target: pmsg.Body.Target,
            }
            cmsg := ClientMessage{
                Body: cmd,
                PI: pmsg.PI,
            }
            fmt.Println("(ReceiveFromPeers) Success")
            fmt.Printf("\tm.ID: %s\n",pmsg.ID)
            fmt.Printf("\tm.Addr: %s\n",pmsg.Addr)
            fmt.Printf("\tm.Body: %s\n",pmsg.Body)
            fmt.Printf("\tm.PI: %s\n",pmsg.PI)
            in <- cmsg
        }
    }
}

func ReceivePeerMessage(c net.Conn) (PeerMessage,bool) {
    d := json.NewDecoder(c)
    var pmsg PeerMessage
    err := d.Decode(&pmsg)
    if err != nil {
        log.Println("(ReceivePeerMessage) JSON Error <", c.RemoteAddr(), "> ", err)
        return pmsg, false
    }
    c.Close()
    return pmsg, true
}

func SendToPeers() {
    for {
        pmsg := <-out
        fmt.Println("(SendToPeers) Sending...")
        fmt.Printf("\tm.ID: %s\n",pmsg.ID)
        fmt.Printf("\tm.Addr: %s\n",pmsg.Addr)
        fmt.Printf("\tm.Body: %s\n",pmsg.Body)
        fmt.Printf("\tm.PI: %s\n",pmsg.PI)
        myPeerChans_copy := myPeerChans.Copy()
        go AddToChannels(pmsg,myPeerChans_copy)
        go DistributeToPeers(myPeerChans_copy)
    }
}

func (p *PeerChans) Copy() (map[string]chan PeerMessage) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    copy := make(map[string]chan PeerMessage)
    for addr, ch := range p.m {
        if addr != psoc_addr {
            copy[addr] = ch
        }
    }
    return copy
}

func AddToChannels(pmsg PeerMessage, p map[string]chan PeerMessage) {
    for _, ch := range p {
        ch <- pmsg
    }
}

func DistributeToPeers(p map[string]chan PeerMessage) {
    for addr, ch := range p {
        go DialPeer(addr,ch)
    }
}

func DialPeer(addr string, ch chan PeerMessage) {
    pmsg := <-ch
    c, err := net.Dial("tcp", addr)
    if err != nil {
        log.Println("(DialPeer) Dial Error: <", addr, "> ", err)
        return
    }
    defer func() {
        c.Close()
    }()
    e := json.NewEncoder(c)
    err = e.Encode(pmsg)
    if err != nil {
        log.Println("(DialPeer) JSON Error: <", addr, "> ", err)
        return
    }
    fmt.Printf("(DialPeer) Success: %s\n",addr)
}

// RandomID returns an 8 byte random string in hexadecimal.
func RandomID() string {
    b := make([]byte, 8)
    n, _ := rand.Read(b)
    return fmt.Sprintf("%x", b[:n])
}

func main() {

    // ONLY FOR LOCAL TESTING
    RetrieveSockets()

    out = make(chan PeerMessage)
    in = make(chan ClientMessage)
    blockExit = make(chan string)

    // Listen at Peer Socket
    fmt.Printf("(Main) Listening for Peers at http://localhost%s\n", psoc_addr)
    l, err := net.Listen("tcp",psoc_addr)
    if err != nil {
        log.Fatal(err)
    }

    ServePeers(l)

    // Create Client Websocket Handler
    http.Handle("/ws", websocket.Handler(ServeClient))

    // Listen at Client Websocket
    fmt.Printf("(Main) Listening for JS Client at ws://localhost%s/ws\n", cws_addr)
    err = http.ListenAndServe(cws_addr, nil)
    if err != nil {
        panic("(Main) ListenAndServe: " + err.Error())
    }

}