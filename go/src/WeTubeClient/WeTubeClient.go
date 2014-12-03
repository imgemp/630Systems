package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "golang.org/x/net/websocket"
    "strings"
    "strconv"
    // "errors"
    // "encoding/binary"
    // "bytes"
    "encoding/json"
    "crypto/rand"
    "sync"
    // "time"
)

// Global Variables
var (
    origin string = "http://localhost/"
    localPort int
    p2pPort int
    myLocalWebsocketAddr string
    myP2PSocketAddr string
    myPeerInfo = &PeerInfo{m: make(map[string]int)} // Map of Peer Addresses to Permission Levels
    myPeerChans = &PeerChans{m: make(map[string]chan PeerMessage)} // Map of Peer Addresses to Channels
)

const VIEWER int = 0
const EDITOR int = 1
const MASTER int = 2

// Ping Server Websocket for Ports AND PEER LIST WITH PERMISSIONS FOR EACH PEER (PASS AS JSON OBJECT)
func RetrieveSockets() {
    url := "ws://localhost:8080/ws/go"
    tempWS, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    if _, err = tempWS.Write([]byte("Which ports should I use?")); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Which ports should I use?\n")
    }

    // Retrieve Ports from Websocket
    var msg = make([]byte, 512)
    var n int
    if n, err = tempWS.Read(msg); err != nil {
        log.Fatal(err)
    }

    // Construct Local and P2P Websocket Addresses
    localPort, err = strconv.Atoi(string(msg[:n]))
    p2pPort = localPort + 1
    var empty []string
    var temp []string
    temp = append(empty,":")
    temp = append(temp,strconv.Itoa(localPort))
    myLocalWebsocketAddr = strings.Join(temp,"")
    temp = append(empty,":")
    temp = append(temp,strconv.Itoa(p2pPort))
    myP2PSocketAddr = strings.Join(temp,"")
    fmt.Printf("WeTubeServer: Use port %s and %s\n",strconv.Itoa(localPort),strconv.Itoa(p2pPort))
}

func RetrievePeerInfo() {
    url := "ws://localhost:8080/ws/go/peer"
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    if _, err = ws.Write([]byte(myP2PSocketAddr)); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Who are the peers? (myP2PSocketAddr: %s)\n",myP2PSocketAddr)
    }

    // Retrieve Peer Info from Websocket
    d := json.NewDecoder(ws)
    err = d.Decode(&myPeerInfo.m)
    if err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Received Peer Info\n")
        for addr, rank := range myPeerInfo.m {
            fmt.Println("Addr:", addr, "Rank:", rank)
        }
    }
    // Transfer this info to peerchans
    fmt.Println("Making Peer Channels")
    for addr, rank := range myPeerInfo.m {
        fmt.Println("Addr:", addr, "Rank:", rank)
        myPeerChans.m[addr] = make(chan PeerMessage)
    }
}

var out chan PeerMessage
var in chan ClientMessage
var blockExit chan string

type Command struct {
    Action string
    Argument string
    Target string
}

// Serve JS Client WebSocket Connection
func ServeClient(ws *websocket.Conn) {
    var reason string = "None"
    go ReceiveFromClient(ws)
    go SendToClient(ws)
    reason = <-blockExit
    fmt.Println("Closing Client WebSocket Connection: %s\n",reason)
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
                Addr: myP2PSocketAddr,
                Body: cmd,
                PI: myPeerInfo.m,
            }
            out <- pmsg
            fmt.Println("Received Message From Client")
            fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
            fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
            fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
            fmt.Printf("\tPI: %s\n",cmsg.PI)
        }
    }
}

func UpdatePeers(PI map[string]int) {
    for addr, rank := range PI {
        fmt.Println("Addr:", addr, "Rank:", rank)
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

func SendToClient(ws *websocket.Conn) {
    for {
        cmsg := <-in
        e := json.NewEncoder(ws)
        err := e.Encode(cmsg)
        if err != nil {
            log.Fatal(err)
        } else {
            cmd := cmsg.Body
            fmt.Println("Sent Message To Client")
            fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
            fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
            fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
            fmt.Printf("\tPI: %s\n",cmsg.PI)
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
                log.Fatal(err)
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
            fmt.Println("Received PeerMessage From Peer")
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
        log.Println("<", c.RemoteAddr(), "error:", err)
        return pmsg, false
    }
    c.Close()
    return pmsg, true
}

func SendToPeers() {
    for {
        pmsg := <-out
        fmt.Println("Sending PeerMessage To Peers")
        fmt.Printf("\tm.ID: %s\n",pmsg.ID)
        fmt.Printf("\tm.Addr: %s\n",pmsg.Addr)
        fmt.Printf("\tm.Body: %s\n",pmsg.Body)
        fmt.Printf("\tm.PI: %s\n",pmsg.PI)
        myPeerChans_copy := myPeerChans.Copy()
        go AddToChannels(pmsg,myPeerChans_copy)
        go DistributeToPeers(myPeerChans_copy)
    }
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
        log.Println(">", addr, "dial error:", err)
    }
    defer func() {
        c.Close()
    }()
    e := json.NewEncoder(c)
    err = e.Encode(pmsg)
    if err != nil {
        log.Println(">", addr, "error:", err)
    }
    fmt.Printf("Sent PeerMessage To Peer: %s\n",addr)
}

// Relay the data received on the WebSocket.
// func AnswerClient(ws *websocket.Conn) {
//     // Receive PeerMessage
//     var msg = make([]byte, 512)
//     var n int
//     var err error
//     if n, err = ws.Read(msg); err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("JS Client: %s\n", msg[:n])
//     var s string = string(msg[:n])
//     if (s == "Who are my peers?") {
//         // Respond with peers
//         duration := time.Duration(2)*time.Second
//         time.Sleep(duration)
//         e := json.NewEncoder(ws)
//         fmt.Println("Creating encoder")
//         time.Sleep(duration)
//         duration = time.Duration(2)*time.Second
//         for addr, rank := range myPeerInfo.m {
//             fmt.Println("Addr:", addr, "Rank:", rank)
//         }
//         err = e.Encode(myPeerInfo.m)
//         fmt.Println("Actually encoding")
//         time.Sleep(duration)
//         duration = time.Duration(4)*time.Second
//         if err != nil {
//             log.Fatal(err)
//         } else {
//             fmt.Printf("Peer Info Sent\n")
//             time.Sleep(duration)
//             duration = time.Duration(2)*time.Second
//         }
//         // Alert peers to client's existence
//         // create PeerMessage with command: Mote, arg_str: client's address, arg_int: PeerInfo.m
//     } else {
//         m := PeerMessage{
//             ID:   RandomID(),
//             Addr: myP2PSocketAddr,
//             Body: s,
//             PI: myPeerInfo.m,
//         }
//         Seen(m.ID)
//         fmt.Printf("About to broadcast: %s\n",m)
//         go broadcast(m)
//         fmt.Println("PeerMessage broadcasted")
//         for addr, _ := range myPeerChans.m {
//             fmt.Printf("Attempting to dial peer %s\n",addr)
//             go DialPeer(addr)
//         }
//     }
// }

// func AnswerPeer(c net.Conn) {
//     log.Println("<", c.RemoteAddr(), "accepted connection")
//     d := json.NewDecoder(c)
//     for {
//         fmt.Println("got here1")
//         var m PeerMessage
//         err := d.Decode(&m)
//         fmt.Println("done decoding")
//         if err != nil {
//             log.Println("<", c.RemoteAddr(), "error:", err)
//             break
//         }
//         fmt.Println("got here2")
//         if Seen(m.ID) {
//             continue
//         }
//         log.Printf("< %v received: %v", c.RemoteAddr(), m)
//         fmt.Println(m.Body)
//         // Send PeerMessage to Client
//         DialClient(m)
//     }
//     c.Close()
//     log.Println("<", c.RemoteAddr(), "close")
// }

// func DialClient(msg PeerMessage) {
//     fmt.Println("dialing client")
//     // "ws://localhost:8080/ws/go"
//     var temp []string
//     temp = append(temp,"ws://localhost")
//     temp = append(temp,myLocalWebsocketAddr)
//     temp = append(temp,"/ws")
//     var wsaddr string = strings.Join(temp,"")
//     ws, err := websocket.Dial(wsaddr, "", origin)
//     if err != nil {
//         fmt.Println("Error dialing client")
//         log.Fatal(err)
//     } else {
//         fmt.Printf("Dial succesful: %s\n",wsaddr)
//     }
//     e := json.NewEncoder(ws)
//     err = e.Encode(msg)
//     if err != nil {
//         fmt.Println("Error encoding PeerMessage to js client")
//         log.Fatal(err)
//     } else {
//         fmt.Printf("PeerMessage Sent\n")
//     }
// }

// RandomID returns an 8 byte random string in hexadecimal.
func RandomID() string {
    b := make([]byte, 8)
    n, _ := rand.Read(b)
    return fmt.Sprintf("%x", b[:n])
}

type PeerInfo struct {
    m  map[string]int
    mu sync.RWMutex
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

type PeerChans struct {
    m  map[string]chan PeerMessage
    mu sync.RWMutex
}

// func (p *PeerChans) Add(addr string) <-chan PeerMessage {
//     p.mu.Lock()
//     defer p.mu.Unlock()
//     if _, ok := p.m[addr]; ok {
//         return nil
//     }
//     ch := make(chan PeerMessage)
//     var peer = Peer{ch: ch, Rank: -1}
//     p.m[addr] = peer
//     return ch
// }

// func (p *PeerChans) Remove(addr string) {
//     p.mu.Lock()
//     defer p.mu.Unlock()
//     delete(p.m, addr)
// }

// func (p *PeerChans) List() []chan PeerMessage {
//     p.mu.RLock()
//     defer p.mu.RUnlock()
//     l := make([]chan PeerMessage, 0, len(p.m)-1)
//     for addr, ch := range p.m {
//         if addr != myP2PSocketAddr {
//             l = append(l, ch)
//         }
//     }
//     return l
// }

func (p *PeerChans) Copy() (map[string]chan PeerMessage) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    copy := make(map[string]chan PeerMessage)
    for addr, ch := range p.m {
        if addr != myP2PSocketAddr {
            copy[addr] = ch
        }
    }
    return copy
}

// func broadcast(m PeerMessage) {
//     // channel := make(chan int)
//     // var two int = 2
//     // channel <- two
//     fmt.Println("got here")
//     // fmt.Printf("Length of channel list = %d\n",len(test))
//     for _, ch := range myPeerChans.List() {
//         fmt.Printf("PeerMessage is %s\n",m)
//         ch <- m // block on channels for now
//         // select {
//         // case ch <- m:
//         // default:
//         //     // Okay to drop PeerMessages sometimes.
//         // }
//     }
// }

// DIAL FUNCTION NEEDS WORK
// func DialPeer(addr string) {
//     if addr == myP2PSocketAddr {
//         return // Don't try to dial self.
//     }

//     // ch := myPeerChans.Add(addr)
//     // if ch == nil {
//     //     return // Peer already connected.
//     // }
//     // defer myPeerChans.Remove(addr)
//     // ch := peers.m[addr]
//     ch := myPeerChans.m[addr]

//     log.Println(">", addr, "dialing")
//     c, err := net.Dial("tcp", addr)
//     if err != nil {
//         log.Println(">", addr, "dial error:", err)
//         return
//     }
//     log.Println(">", addr, "connected")
//     defer func() {
//         c.Close()
//         log.Println(">", addr, "closed")
//     }()

//     fmt.Println("setting up encoder")
//     e := json.NewEncoder(c)
//     fmt.Println("blocking on channel")
//     m := <-ch
//     fmt.Println("printing PeerMessage")
//     fmt.Println(m)
//     fmt.Println("encoding PeerMessage")
//     err = e.Encode(m)
//     if err != nil {
//         log.Println(">", addr, "error:", err)
//         return
//     }
//     // for m := range ch { // why range?
//     //     err := e.Encode(m)
//     //     if err != nil {
//     //         log.Println(">", addr, "error:", err)
//     //         return
//     //     }
//     // }
// }

// var seenIDs = struct {
//     m map[string]bool
//     sync.Mutex
// }{m: make(map[string]bool)}

// func Seen(id string) bool {
//     if id == "" {
//         return false
//     }
//     seenIDs.Lock()
//     ok := seenIDs.m[id]
//     seenIDs.m[id] = true
//     seenIDs.Unlock()
//     return ok
// }

func main() {

    // ONLY FOR TESTING
    RetrieveSockets()
    // RetrievePeerInfo()

    out = make(chan PeerMessage)
    in = make(chan ClientMessage)
    blockExit = make(chan string)

    // Listen at Peer Socket
    fmt.Printf("Listening for Peers at http://localhost%s\n", myP2PSocketAddr)
    l, err := net.Listen("tcp",myP2PSocketAddr)
    if err != nil {
        log.Fatal(err)
    }

    // Serve Peers on Request
    // go func() {
    //     for {
    //         fmt.Printf("Waiting for peer\n")
    //         c, err := l.Accept()
    //         if err != nil {
    //                 log.Fatal(err)
    //         }
    //         fmt.Printf("Accepting peer\n")
    //         go AnswerPeer(c)
    //     }
    // }()
    ServePeers(l)

    // Create Client Websocket Handler
    http.Handle("/ws", websocket.Handler(ServeClient))

    // Listen at Client Websocket
    fmt.Printf("Listening for JS Client at ws://localhost%s/ws\n", myLocalWebsocketAddr)
    err = http.ListenAndServe(myLocalWebsocketAddr, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }

}