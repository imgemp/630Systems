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
)

// Global Variables
var (
    origin string = "http://localhost/"
    localPort int
    p2pPort int
    // myLocalWebsocket websocket.Conn
    myLocalWebsocketAddr string
    myP2PSocketAddr string
    // count int // FOR TESTING
    // newInput bool // FOR TESTING
)

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

func ClientHandler(ws *websocket.Conn) {
    ClientRelay(ws)
}

// Relay the data received on the WebSocket.
func ClientRelay(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("JS Client: %s.\n", msg[:n])
    // ATTACH PEER LIST TO ALL RELAY MESSAGES
    var s string = string(msg[:n])
    m := Message{
        ID:   RandomID(),
        Addr: myP2PSocketAddr,
        Body: s,
    }
    Seen(m.ID)
    broadcast(m)

    // if _, err := ws.Write([]byte("Hello to you too!")); err != nil {
    //     log.Fatal(err)
    // } else {
    //     fmt.Printf("Hello to you too!\n")
    // }


    // for {
    //     if newInput {
    //         s := strconv.Itoa(count)
    //         // fmt.Printf("Count String is %s\n",s)
    //         m := Message{
    //             ID:   RandomID(),
    //             Addr: myP2PSocketAddr,
    //             Body: s,
    //         }
    //         // fmt.Printf("Count String is now %s\n",m.Body)
    //         Seen(m.ID)
    //         broadcast(m)
    //         // fmt.Println("done broadcasting original message")
    //     }
    // }
}

// Relay the data received on the WebSocket.
// func P2P() {
//     // Receive Message
//     // var msg = make([]byte, 512)
//     // var n int
//     // var err error
//     // if n, err = ws.Read(msg); err != nil {
//     //     log.Fatal(err)
//     // }
//     // if n != 0 {
//     //     fmt.Printf("Other Peer: %s.\n", msg[:n])
//     // }
//     // buf := bytes.NewBuffer(msg[:n])
//     // binary.Read(buf, binary.LittleEndian, &count)
//     for true {
//         // Increment Count and Write to Buffer
//         buf := new(bytes.Buffer)
//         count = <-p2pChannel
//         count = count + 1
//         var err error
//         err = binary.Write(buf, binary.LittleEndian, count)
//         if err != nil {
//             fmt.Println("binary.Write failed:", err)
//         }
//         // Dial Server Websocket for Peer List
//         var serverWebsocket_URL = "ws://localhost:8080/ws/go/peer"
//         var serverWebsocket *websocket.Conn
//         serverWebsocket, err = websocket.Dial(serverWebsocket_URL, "", origin)
//         if err != nil {
//             log.Fatal(err)
//         }
//         if _, err := serverWebsocket.Write([]byte(myP2PWebsocket)); err != nil {
//             log.Fatal(err)
//         } else {
//             fmt.Printf("What is my peer's websocket port?\n")
//         }
//         // Retrieve Peer List from Server - just retrieving other peer websocket for now
//         var msg = make([]byte, 512)
//         var n int
//         if n, err = serverWebsocket.Read(msg); err != nil {
//             log.Fatal(err)
//         }
//         fmt.Printf("WebServer: Your peer's websocket port is %s\n", msg[:n])
//         var temp []string
//         var myPeerWebsocket_URL string
//         var myPeerWebsocket *websocket.Conn
//         temp = append(temp,"ws://localhost:")
//         temp = append(temp,string(msg[:n]))
//         temp = append(temp,"/ws/peer")
//         myPeerWebsocket_URL = strings.Join(temp,"")
//         // Write Message to Peers
//         myPeerWebsocket, err = websocket.Dial(myPeerWebsocket_URL, "", origin)
//         if err != nil {
//             log.Fatal(err)
//         }
//         if _, err := myPeerWebsocket.Write(buf.Bytes()); err != nil {
//             log.Fatal(err)
//         } else {
//             fmt.Printf("I've updated count to %d\n",count)
//         }
//     }
// }

// func Add2PeerChannel(ws *websocket.Conn) {
//     // Receive Message
//     var msg = make([]byte, 512)
//     var n int
//     var err error
//     if n, err = ws.Read(msg); err != nil {
//         log.Fatal(err)
//     }
//     if n != 0 {
//         fmt.Printf("Other Peer: %s.\n", msg[:n])
//     }
//     buf := bytes.NewBuffer(msg[:n])
//     binary.Read(buf, binary.LittleEndian, &count)
//     p2pChannel <- count
// }

// func servePeer(c net.Conn) {
//     fmt.Printf("<%saccepted connection",c.RemoteAddr())
//     // Read count
//     var msg = make([]byte, 512)
//     var n int
//     var err error
//     if n, err = c.Read(msg); err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("Peer: %s\n", msg[:n])
//     buf := bytes.NewBuffer(msg[:n])
//     binary.Read(buf, binary.LittleEndian, &count)
//     fmt.Printf("Peer: Count is %d\n",count)
//     // Increment Count
//     count += 1
//     err = binary.Write(buf, binary.LittleEndian, count)
//     if err != nil {
//         fmt.Println("binary.Write failed:", err)
//     }
//     // Redial Peer
//     var addr string = "localhost:12348"
//     if (p2pPort == 12348) {
//         addr = "localhost:12346"
//     }
//     fmt.Printf("Dialing peer's socket at %s\n",addr)
//     c, err = net.Dial("tcp",addr)
//     if err != nil {
//         fmt.Printf("Peer's socket not ready!\n")
//         log.Fatal(err)
//     } else {
//         fmt.Printf("Dial successful!\n")
//     }
//     // Write to socket
//     if _, err := c.Write(buf.Bytes()); err != nil {
//         log.Fatal(err)
//     } else {
//         fmt.Printf("Count is %d\n",count)
//     }
//     c.Close()
//     fmt.Printf("<%sclosed\n",c.RemoteAddr())
// }

// func startComm() {
//     // Test by Incrementing Count back and forth
//     count = 0
//     if (p2pPort == 12348) {
//         // Convert count to binary
//         buf := new(bytes.Buffer)
//         err := binary.Write(buf, binary.LittleEndian, count)
//         if err != nil {
//             fmt.Println("binary.Write failed:", err)
//         }
//         // Dial Peer
//         var addr string = "localhost:12346"
//         fmt.Printf("Dialing peer's socket at %s\n",addr)
//         c, err := net.Dial("tcp",addr)
//         if err != nil {
//             fmt.Printf("Peer's socket not ready!\n")
//             log.Fatal(err)
//         } else {
//             fmt.Printf("Dial successful!\n")
//         }
//         // Write to socket
//         if _, err := c.Write(buf.Bytes()); err != nil {
//             log.Fatal(err)
//         } else {
//             fmt.Printf("Count is %d\n",count)
//         }
//     }
// }

// func startComm() {
//     // Test by Incrementing Count back and forth
//     count = 0
//     newInput = true
//     if (p2pPort == 12348) {
//         var addr string = "localhost:12346"
//         // ch := make(chan<- Message)
//         // peers.m[addr] = ch
//         // Add count to channels
//         go readInput()
//         // Dial Peer
//         fmt.Printf("Dialing peer's socket at %s\n",addr)
//         go dial(addr)
//     }
// }

func servePeer(c net.Conn) {
    log.Println("<", c.RemoteAddr(), "accepted connection")
    d := json.NewDecoder(c)
    for {
        var m Message
        // fmt.Println("decoding message")
        err := d.Decode(&m)
        if err != nil {
            log.Println("<", c.RemoteAddr(), "error:", err)
            break
        }
        if Seen(m.ID) {
            continue
        }
        log.Printf("< %v received: %v", c.RemoteAddr(), m)
        fmt.Println(m.Body)

        // Write body to websocket
        // fmt.Println("writing to websocket")
        // myLocalWebsocket, err := websocket.Dial(myLocalWebsocketAddr, "", origin)
        // if err != nil {
        //     log.Fatal(err)
        // }
        // if _, err = myLocalWebsocket.Write([]byte(string(m.Body))); err != nil {
        //     log.Fatal(err)
        // } else {
        //     fmt.Printf("Wrote to websocket: %s\n",m.Body)
        // }
        // Intercept message and increment count - FOR TESTING
        // var body string = m.Body
        // count, err = strconv.Atoi(body)
        // if err != nil {
        //     log.Println("string to int error line 266")
        // }
        // count = count + 1
        // newInput = true
        // Broadcast
        broadcast(m)
        go dial(m.Addr)
    }
    c.Close()
    log.Println("<", c.RemoteAddr(), "close")
}

// RandomID returns an 8 byte random string in hexadecimal.
func RandomID() string {
    b := make([]byte, 8)
    n, _ := rand.Read(b)
    return fmt.Sprintf("%x", b[:n])
}

// needs to take input from JS Client, just testing now
// func readInput() {
//     for {
//         if newInput {
//             s := strconv.Itoa(count)
//             // fmt.Printf("Count String is %s\n",s)
//             m := Message{
//                 ID:   RandomID(),
//                 Addr: myP2PSocketAddr,
//                 Body: s,
//             }
//             // fmt.Printf("Count String is now %s\n",m.Body)
//             Seen(m.ID)
//             broadcast(m)
//             // fmt.Println("done broadcasting original message")
//         }
//     }
// }

type Message struct {
    ID   string
    Addr string
    Body string
}

var peers = &Peers{m: make(map[string]chan<- Message)}

type Peers struct {
    m  map[string]chan<- Message // receive-only channels
    mu sync.RWMutex
}

func (p *Peers) Add(addr string) <-chan Message {
    p.mu.Lock()
    defer p.mu.Unlock()
    if _, ok := p.m[addr]; ok {
        return nil
    }
    ch := make(chan Message)
    p.m[addr] = ch
    return ch
}

// chan<- is receive only
// <-chan is send only

// func (p *Peers) Add(addr string) <-chan Message {
//     p.mu.Lock()
//     defer p.mu.Unlock()
//     if _, ok := p.m[addr]; ok {
//         ch := make(chan Message)
//         var m Message
//         m = <- p.m[addr]
//         ch <- m
//         return ch
//     }
//     ch := make(chan Message)
//     p.m[addr] = ch
//     return ch
// }

func (p *Peers) Remove(addr string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    delete(p.m, addr)
}

func (p *Peers) List() []chan<- Message {
    p.mu.RLock()
    defer p.mu.RUnlock()
    l := make([]chan<- Message, 0, len(p.m))
    for _, ch := range p.m {
        l = append(l, ch)
    }
    return l
}

func broadcast(m Message) {
    for _, ch := range peers.List() {
        // fmt.Println("adding message to peer channels")
        select {
        case ch <- m:
        default:
            // Okay to drop messages sometimes.
        }
    }
}

// DIAL FUNCTION NEEDS WORK
func dial(addr string) {
    if addr == myP2PSocketAddr {
        return // Don't try to dial self.
    }

    ch := peers.Add(addr)
    if ch == nil {
        return // Peer already connected.
    }
    defer peers.Remove(addr)
    // ch := peers.m[addr]

    log.Println(">", addr, "dialing")
    c, err := net.Dial("tcp", addr)
    if err != nil {
        log.Println(">", addr, "dial error:", err)
        return
    }
    log.Println(">", addr, "connected")
    defer func() {
        c.Close()
        log.Println(">", addr, "closed")
    }()

    e := json.NewEncoder(c)
    // fmt.Println("Encoding")
    for m := range ch {
        // fmt.Println("got here")
        // fmt.Printf("message is:%s",m)
        err := e.Encode(m)
        if err != nil {
            log.Println(">", addr, "error:", err)
            return
        }
    }
    // newInput = false
    // fmt.Println("Done Encoding")
}

var seenIDs = struct {
    m map[string]bool
    sync.Mutex
}{m: make(map[string]bool)}

func Seen(id string) bool {
    if id == "" {
        return false
    }
    seenIDs.Lock()
    ok := seenIDs.m[id]
    seenIDs.m[id] = true
    seenIDs.Unlock()
    return ok
}

func main() {

    // ONLY FOR TESTING
    RetrieveSockets()

    // Listen at Peer Socket
    fmt.Printf("Listening for Peers at http://localhost%s\n", myP2PSocketAddr)
    l, err := net.Listen("tcp",myP2PSocketAddr)
    if err != nil {
        log.Fatal(err)
    }

    // Serve Peers on Request
    go func() {
        for {
            fmt.Printf("Waiting for peer\n")
            c, err := l.Accept()
            if err != nil {
                    log.Fatal(err)
            }
            fmt.Printf("Accepting peer\n")
            go servePeer(c)
        }
    }()

    // Create Client Websocket Handler
    http.Handle("/ws", websocket.Handler(ClientHandler))

    // Listen at Client Websocket
    fmt.Printf("Listening for JS Client at ws://localhost%s/ws\n", myLocalWebsocketAddr)
    err = http.ListenAndServe(myLocalWebsocketAddr, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }

}