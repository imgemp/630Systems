package main

import (
    "fmt"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "strings"
)

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

func main() {

    // Retrieve Port # from Server
    origin := "http://localhost/"
    url := "ws://localhost:8080/ws/go"
    tempWS, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    if _, err := tempWS.Write([]byte("Which port should I use?")); err != nil {
        log.Fatal(err)
    } else {
        fmt.Printf("Which port should I use?\n")
    }
    var Port = make([]byte, 512)
    var n int
    if n, err = tempWS.Read(Port); err != nil {
        log.Fatal(err)
    }
    var wsAddress string
    var temp []string
    temp = append(temp,":")
    temp = append(temp,string(Port[:n]))
    wsAddress = strings.Join(temp,"")
    fmt.Printf("WeTubeServer: Use port %s\n",Port[:n])
    // tempWS.Close() // How do I hang up the websocket connection?

    // Create Client Websocket
    fmt.Printf("Starting local server with websocket at ws://localhost%s/ws\n", wsAddress)
    http.Handle("/ws", websocket.Handler(RelayAndRespond))
    err = http.ListenAndServe(wsAddress, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}