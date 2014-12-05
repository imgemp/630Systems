package main

import (
    "fmt"
    "log"
    "net"
    "bytes"
    "net/http"
    "golang.org/x/net/websocket"
    "encoding/json"
    "crypto/rand"
    "crypto/rsa"
    "sync"
    "crypto/md5"
    "encoding/binary"
    "encoding/gob"
    "os"
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

type ServerInit struct {
    CWS_addr string
    PSOC_addr string
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

    key_len int
    pvkey *rsa.PrivateKey
    pbkey *rsa.PublicKey
    pbkey_Server *rsa.PublicKey
)

const (
    VIEWER int = 0
    EDITOR int = 1
    MASTER int = 2
)

func EncryptMessage(msg []byte, pbkey *rsa.PublicKey) []byte {
    md5hash := md5.New()
    var encrypted_packets [][]byte
    var msg_len int = len(msg)
    var packet_size int = 32
    var full_packets int = msg_len/packet_size
    var remainder int = msg_len - (msg_len/packet_size)*packet_size
    
    var total_bytes uint32
    if remainder > 0 {
        total_bytes = uint32((full_packets+1)*key_len/8)
    } else {
        total_bytes = uint32(full_packets*key_len/8)
    }

    total_bytes_bytes := make([]byte,4)
    binary.LittleEndian.PutUint32(total_bytes_bytes,total_bytes)
    packet, err := rsa.EncryptOAEP(md5hash, rand.Reader, pbkey, total_bytes_bytes, nil)
    encrypted_packets = append(encrypted_packets,packet)
    fmt.Printf("OAEP encrypted [%d] to \n[%x]\n", total_bytes, packet)

    for packet_num := 0; packet_num < full_packets; packet_num++ {
        fmt.Printf("OAEP encrypting [%s]...\n", string(msg[packet_size*packet_num:packet_size*(packet_num+1)]))
        packet, err := rsa.EncryptOAEP(md5hash, rand.Reader, pbkey, msg[packet_size*packet_num:packet_size*(packet_num+1)], nil)
        if err != nil {
            log.Fatal("(EncryptMessage) ",err)
        } else {
            fmt.Printf("...to [%x]\n", packet)
            encrypted_packets = append(encrypted_packets,packet)
        }
    }
    packet, err = rsa.EncryptOAEP(md5hash, rand.Reader, pbkey, msg[packet_size*full_packets:], nil)
    if err != nil {
        log.Fatal("(EncryptMessage) ",err)
    } else {
        fmt.Printf("OAEP encrypted [%s] to \n[%x]\n", string(msg[packet_size*full_packets:]), packet)
        encrypted_packets = append(encrypted_packets,packet)
    }
    return bytes.Join(encrypted_packets,nil)
}

func DecryptMessage(msg []byte) []byte {
    md5hash := md5.New()
    var decrypted_packets [][]byte
    var msg_len int = len(msg)
    var packet_size int = key_len/8 // key_len measured in bits and each byte is 8 bits
    var full_packets int = msg_len/packet_size
    for packet_num := 0; packet_num < full_packets; packet_num++ {
        fmt.Printf("OAEP decrypting [%x]...\n", string(msg[packet_size*packet_num:packet_size*(packet_num+1)]))
        packet, err := rsa.DecryptOAEP(md5hash, rand.Reader, pvkey, msg[packet_size*packet_num:packet_size*(packet_num+1)], nil)
        if err != nil {
            log.Fatal("(DecryptMessage) ",err)
        } else {
            fmt.Printf("...to [%s]\n", packet)
            decrypted_packets = append(decrypted_packets,packet)
        }
    }
    return bytes.Join(decrypted_packets,nil)
}

func ReadWebSocket(ws *websocket.Conn) []byte {
    total_bytes_bytes_encrypted := make([]byte,key_len/8)
    if _, err := ws.Read(total_bytes_bytes_encrypted); err != nil {
        log.Println("(ReadWebSocket) Read Total Bytes: ",err)
    }

    md5hash := md5.New()
    fmt.Println("ReadWebSocket: ",pvkey)
    total_bytes_bytes, err := rsa.DecryptOAEP(md5hash, rand.Reader, pvkey, total_bytes_bytes_encrypted, nil)
    if err != nil {
        log.Println("(ReadWebSocket) Decryption: ",err)
    }

    var total_bytes uint32
    total_bytes = binary.LittleEndian.Uint32(total_bytes_bytes)

    fmt.Printf("OAEP decrypted [%x] to \n[%d]\n",total_bytes_bytes_encrypted, total_bytes)

    msg := make([]byte, total_bytes)
    if _, err := ws.Read(msg); err != nil {
        log.Println("(ReadWebSocket): Read Msg ",err)
    }
    return msg
}

func ReadTCPSocket(c net.Conn) []byte {
    total_bytes_bytes_encrypted := make([]byte,key_len/8)
    if _, err := c.Read(total_bytes_bytes_encrypted); err != nil {
        log.Println("(ReadTCPSocket) ",err)
    }

    md5hash := md5.New()
    total_bytes_bytes, err := rsa.DecryptOAEP(md5hash, rand.Reader, pvkey, total_bytes_bytes_encrypted, nil)
    if err != nil {
        log.Println("(ReadTCPSocket) ",err)
    }

    var total_bytes uint32
    total_bytes = binary.LittleEndian.Uint32(total_bytes_bytes)

    fmt.Printf("OAEP decrypted [%x] to \n[%d]\n",total_bytes_bytes_encrypted, total_bytes)

    msg := make([]byte, total_bytes)
    if _, err := c.Read(msg); err != nil {
        log.Println("(ReadTCPSocket) ",err)
    }
    return msg
}

// ONLY FOR LOCAL TESTING
func RetrieveSockets(pbkey_Server *rsa.PublicKey) {
    url := "ws://localhost:8080/ws/go"
    ws, err := websocket.Dial(url, "", "http://localhost/")
    if err != nil {
        log.Fatal("(RetrieveSockets) ",err)
    }

    // Send Public Key to Server
    pbkey_bytes, err := json.Marshal(pbkey)
    if err != nil {
        log.Fatal("(RetrieveSockets) ",err)
    }
    // sign message
    pbkey_encrypted := EncryptMessage(pbkey_bytes,pbkey_Server)
    // fmt.Printf("OAEP encrypted [%s] to \n[%x]\n", string(pbkey_bytes), pbkey_encrypted)
    if _, err := ws.Write(pbkey_encrypted); err != nil {
        log.Fatal("(RetrieveSockets) ",err)
    }
    fmt.Println("Sent public key to server")

    // Retrieve Client Websocket and P2P Socket Addresses
    msg_encrypted := ReadWebSocket(ws)
    msg_decrypted := DecryptMessage(msg_encrypted)
    // VerifyPKCS1v15(pbkey, hash crypto.Hash, hashed []byte, sig []byte) (err error)
    msg_verified := msg_decrypted
    var smsg ServerInit
    err = json.Unmarshal(msg_verified,&smsg)
    if err != nil {
        log.Fatal("(RetrieveSockets) ",err)
    } else {
        fmt.Println("Received info from server")
        fmt.Println(smsg)
        UpdatePeers(smsg.PI)
        cws_addr = smsg.CWS_addr
        psoc_addr = smsg.PSOC_addr
    }
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
            log.Fatal("(ReceiveFromClient) ",err)
        } else {
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

    msg_encrypted := ReadTCPSocket(c)
    msg_decrypted := DecryptMessage(msg_encrypted)
    // VerifyPKCS1v15(pbkey, hash crypto.Hash, hashed []byte, sig []byte) (err error)
    msg_verified := msg_decrypted
    var pmsg PeerMessage
    err := json.Unmarshal(msg_verified,&pmsg)
    if err != nil {
        log.Println("(ReceivePeerMessage) JSON Error <", c.RemoteAddr(), "> ", err)
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
    pmsg_bytes, err := json.Marshal(pmsg)
    if err != nil {
        log.Println("(DialPeer) JSON Error: <", addr, "> ", err)
        return
    }
    // sign message
    pmsg_encrypted := EncryptMessage(pmsg_bytes,pbkey) // need peer's public key
    if _, err := c.Write(pmsg_encrypted); err != nil {
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

    // Generate Private and Public Key
    key_len = 1024
    var err error
    pvkey, err = rsa.GenerateKey(rand.Reader, key_len)
    pbkey = &pvkey.PublicKey

    // HTTP Server's Public Key
    file, err := os.Open("go/src/WeTubeClient/publicgob.key")
    if err != nil {
        log.Fatal(err)
    }
    d := gob.NewDecoder(file)
    err = d.Decode(&pbkey_Server)

    // ONLY FOR LOCAL TESTING
    RetrieveSockets(pbkey_Server)

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