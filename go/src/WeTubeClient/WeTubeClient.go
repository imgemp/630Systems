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
    myPeerChans = &PeerChans{m: make(map[string]chan Message)} // Map of Peer Addresses to Channels
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
        myPeerChans.m[addr] = make(chan Message)
    }
}

var out chan Message
var in chan Command
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
    var cmd Command
    for {
        err := d.Decode(&cmd)
        if err != nil {
            log.Fatal(err)
        } else {
            // Should construct message for peers first here
            m := Message{
                ID: RandomID(),
                Addr: myP2PSocketAddr,
                Body: cmd,
                PI: myPeerInfo.m,
            }
            out <- m
            fmt.Println("Received Command From Client")
            fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
            fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
            fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
        }
    }
}

func SendToClient(ws *websocket.Conn) {
    for {
        select {
            case cmd := <-in:
                e := json.NewEncoder(ws)
                err := e.Encode(cmd)
                if err != nil {
                    log.Fatal(err)
                } else {
                    fmt.Println("Sent Command To Client")
                    fmt.Printf("\tcmd.Action: %s\n",cmd.Action)
                    fmt.Printf("\tcmd.Argument: %s\n",cmd.Argument)
                    fmt.Printf("\tcmd.Target: %s\n",cmd.Target)
                }
            default:
                // Continue on if channel is empty
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
        if m, ok := ReceiveMessage(c); ok {
            cmd := Command{
                Action: m.Body.Action,
                Argument: m.Body.Argument,
                Target: m.Body.Target,
            }
            fmt.Println("Received Message From Peer")
            fmt.Printf("\tm.ID: %s\n",m.ID)
            fmt.Printf("\tm.Addr: %s\n",m.Addr)
            fmt.Printf("\tm.Body: %s\n",m.Body)
            fmt.Printf("\tm.PI: %s\n",m.PI)
            in <- cmd
        }
    }
}

func ReceiveMessage(c net.Conn) (Message,bool) {
    d := json.NewDecoder(c)
    var m Message
    err := d.Decode(&m)
    if err != nil {
        log.Println("<", c.RemoteAddr(), "error:", err)
        return m, false
    }
    c.Close()
    return m, true
}

func SendToPeers() {
    for {
        select {
            case m := <-out:
                fmt.Println("Sending Message To Peers")
                fmt.Printf("\tm.ID: %s\n",m.ID)
                fmt.Printf("\tm.Addr: %s\n",m.Addr)
                fmt.Printf("\tm.Body: %s\n",m.Body)
                fmt.Printf("\tm.PI: %s\n",m.PI)
                myPeerChans_copy := myPeerChans.Copy()
                go AddToChannels(m,myPeerChans_copy)
                go DistributeToPeers(myPeerChans_copy)
            default:
                // Continue on if channel empty
        }
    }
}

func AddToChannels(m Message, p map[string]chan Message) {
    for _, ch := range p {
        ch <- m
    }
}

func DistributeToPeers(p map[string]chan Message) {
    for addr, ch := range p {
        go DialPeer(addr,ch)
    }
}

func DialPeer(addr string, ch chan Message) {
    m := <-ch
    c, err := net.Dial("tcp", addr)
    if err != nil {
        log.Println(">", addr, "dial error:", err)
    }
    defer func() {
        c.Close()
    }()
    e := json.NewEncoder(c)
    err = e.Encode(m)
    if err != nil {
        log.Println(">", addr, "error:", err)
    }
    fmt.Printf("Sent Message To Peer: %s\n",addr)
}

// Relay the data received on the WebSocket.
// func AnswerClient(ws *websocket.Conn) {
//     // Receive Message
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
//         // create message with command: Mote, arg_str: client's address, arg_int: PeerInfo.m
//     } else {
//         m := Message{
//             ID:   RandomID(),
//             Addr: myP2PSocketAddr,
//             Body: s,
//             PI: myPeerInfo.m,
//         }
//         Seen(m.ID)
//         fmt.Printf("About to broadcast: %s\n",m)
//         go broadcast(m)
//         fmt.Println("Message broadcasted")
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
//         var m Message
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
//         // Send Message to Client
//         DialClient(m)
//     }
//     c.Close()
//     log.Println("<", c.RemoteAddr(), "close")
// }

// func DialClient(msg Message) {
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
//         fmt.Println("Error encoding message to js client")
//         log.Fatal(err)
//     } else {
//         fmt.Printf("Message Sent\n")
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

type Message struct {
    ID   string
    Addr string
    Body Command
    PI map[string]int
}

type PeerChans struct {
    m  map[string]chan Message
    mu sync.RWMutex
}

// func (p *PeerChans) Add(addr string) <-chan Message {
//     p.mu.Lock()
//     defer p.mu.Unlock()
//     if _, ok := p.m[addr]; ok {
//         return nil
//     }
//     ch := make(chan Message)
//     var peer = Peer{ch: ch, Rank: -1}
//     p.m[addr] = peer
//     return ch
// }

// func (p *PeerChans) Remove(addr string) {
//     p.mu.Lock()
//     defer p.mu.Unlock()
//     delete(p.m, addr)
// }

// func (p *PeerChans) List() []chan Message {
//     p.mu.RLock()
//     defer p.mu.RUnlock()
//     l := make([]chan Message, 0, len(p.m)-1)
//     for addr, ch := range p.m {
//         if addr != myP2PSocketAddr {
//             l = append(l, ch)
//         }
//     }
//     return l
// }

func (p *PeerChans) Copy() (map[string]chan Message) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    copy := make(map[string]chan Message)
    for addr, ch := range p.m {
        if addr != myP2PSocketAddr {
            copy[addr] = ch
        }
    }
    return copy
}

// func broadcast(m Message) {
//     // channel := make(chan int)
//     // var two int = 2
//     // channel <- two
//     fmt.Println("got here")
//     // fmt.Printf("Length of channel list = %d\n",len(test))
//     for _, ch := range myPeerChans.List() {
//         fmt.Printf("Message is %s\n",m)
//         ch <- m // block on channels for now
//         // select {
//         // case ch <- m:
//         // default:
//         //     // Okay to drop messages sometimes.
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
//     fmt.Println("printing message")
//     fmt.Println(m)
//     fmt.Println("encoding message")
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
    RetrievePeerInfo()

    out = make(chan Message)
    in = make(chan Command)
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