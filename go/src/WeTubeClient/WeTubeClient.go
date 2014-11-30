package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "strings"
    "strconv"
    "encoding/binary"
    "bytes"
)

// Global Variables
var origin string = "http://localhost/"
var localPort int
var p2pPort int
var myLocalWebsocket string
var myP2PWebsocket string
// var localChannel chan int
var p2pChannel chan int32
var count int32

// Relay the data received on the WebSocket.
func RelayAndRespond(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("JS Client: %s.\n", msg[:n])
    // Relay Message to Peers
    // Receive Message from Peers
    // Respond with Message
    if _, err := ws.Write([]byte("Hello to you too!")); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Hello to you too!\n")
    }
}

// Relay the data received on the WebSocket.
func P2P() {
    // Receive Message
    // var msg = make([]byte, 512)
    // var n int
    // var err error
    // if n, err = ws.Read(msg); err != nil {
    //     log.Fatal(err)
    // }
    // if n != 0 {
    //     fmt.Printf("Other Peer: %s.\n", msg[:n])
    // }
    // buf := bytes.NewBuffer(msg[:n])
    // binary.Read(buf, binary.LittleEndian, &count)
    for true {
        // Increment Count and Write to Buffer
        buf := new(bytes.Buffer)
        count = <-p2pChannel
        count = count + 1
        var err error
        err = binary.Write(buf, binary.LittleEndian, count)
        if err != nil {
            fmt.Println("binary.Write failed:", err)
        }
        // Dial Server Websocket for Peer List
        var serverWebsocket_URL = "ws://localhost:8080/ws/go/peer"
        var serverWebsocket *websocket.Conn
        serverWebsocket, err = websocket.Dial(serverWebsocket_URL, "", origin)
        if err != nil {
            log.Fatal(err)
        }
        if _, err := serverWebsocket.Write([]byte(myP2PWebsocket)); err != nil {
            log.Fatal(err)
        } else {
            fmt.Printf("What is my peer's websocket port?\n")
        }
        // Retrieve Peer List from Server - just retrieving other peer websocket for now
        var msg = make([]byte, 512)
        var n int
        if n, err = serverWebsocket.Read(msg); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("WebServer: Your peer's websocket port is %s\n", msg[:n])
        var temp []string
        var myPeerWebsocket_URL string
        var myPeerWebsocket *websocket.Conn
        temp = append(temp,"ws://localhost:")
        temp = append(temp,string(msg[:n]))
        temp = append(temp,"/ws/peer")
        myPeerWebsocket_URL = strings.Join(temp,"")
        // Write Message to Peers
        myPeerWebsocket, err = websocket.Dial(myPeerWebsocket_URL, "", origin)
        if err != nil {
            log.Fatal(err)
        }
        if _, err := myPeerWebsocket.Write(buf.Bytes()); err != nil {
            log.Fatal(err)
        } else {
            fmt.Printf("I've updated count to %d\n",count)
        }
    }
}

// func Listen4JS() {
// }

func Add2PeerChannel(ws *websocket.Conn) {
    // Receive Message
    var msg = make([]byte, 512)
    var n int
    var err error
    if n, err = ws.Read(msg); err != nil {
        log.Fatal(err)
    }
    if n != 0 {
        fmt.Printf("Other Peer: %s.\n", msg[:n])
    }
    buf := bytes.NewBuffer(msg[:n])
    binary.Read(buf, binary.LittleEndian, &count)
    p2pChannel <- count
}

func main() {

    // ONLY FOR TESTING

    // Ping Server Websocket for Ports
    url := "ws://localhost:8080/ws/go/init"
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
    myLocalWebsocket = strings.Join(temp,"")
    temp = append(empty,":")
    temp = append(temp,strconv.Itoa(p2pPort))
    myP2PWebsocket = strings.Join(temp,"")
    fmt.Printf("WeTubeServer: Use port %s and %s\n",strconv.Itoa(localPort),strconv.Itoa(p2pPort))
    // tempWS.Close() // How do I hang up the websocket connection?

    // Create Client Websocket Handler
    http.Handle("/ws", websocket.Handler(RelayAndRespond))

    // Create Peer to Peer Websocket Handler
    http.Handle("/ws/peer", websocket.Handler(Add2PeerChannel))

    // Listen and Serve at Websockets
    fmt.Printf("Listening for [internal,p2p] websocket at ws://localhost[%s/ws,%s/ws/peer]\n", myLocalWebsocket, myP2PWebsocket)
    err = http.ListenAndServe(myLocalWebsocket, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
    err = http.ListenAndServe(myP2PWebsocket, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }

    // Test by Incrementing Count back and forth
    count = 0
    if (p2pPort == 12348) {
        // Convert count to binary
        buf := new(bytes.Buffer)
        err = binary.Write(buf, binary.LittleEndian, count)
        if err != nil {
            fmt.Println("binary.Write failed:", err)
        }
        // Dial websocket
        url = "ws://localhost:12346/ws/peer"
        fmt.Printf("Dialing peer's websocket at %s\n",url)
        tempWS, err = websocket.Dial(url, "", origin)
        if err != nil {
            fmt.Printf("Peer's websocket not ready!\n")
            log.Fatal(err)
        } else {
            fmt.Printf("Initial P2P message sent!\n")
        }
        // Write to websocket
        if _, err := tempWS.Write(buf.Bytes()); err != nil {
            log.Fatal(err)
        } else {
            fmt.Printf("Count is %d\n",count)
        }
    }

    go P2P()
}