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
    "crypto/rsa"
    "crypto/rand"
    "crypto/md5"
    "encoding/binary"
    "bytes"
    "os"
    "encoding/gob"
)

// Types and Globals

type PeerRank struct {
    m  map[string]int
    mu sync.RWMutex
}

type PeerKeys struct {
    m  map[string]rsa.PublicKey
    mu sync.RWMutex
}

type Init struct {
    CWS_addr string
    PSOC_addr string
    PR map[string]int
    PK map[string]rsa.PublicKey
}

var (
    myPeerRank = &PeerRank{m: make(map[string]int)} // Map of Peer Addresses to Permission Levels
    myPeerKeys = &PeerKeys{m: make(map[string]rsa.PublicKey)} // Map of Peer Addresses to Public Keys

    key_len int
    pvkey *rsa.PrivateKey
    pbkey *rsa.PublicKey

    Port int = 12345
    JSPortList []int
    JSClientCount = 0
)

const (
    VIEWER int = 0
    EDITOR int = 1
    DIRECTOR int = 2
)

func DecryptMessage(msg []byte) []byte {
    md5hash := md5.New()
    var decrypted_packets [][]byte
    var msg_len int = len(msg)
    var packet_size int = key_len/8 // key_len measured in bits and each byte is 8 bits
    var full_packets int = msg_len/packet_size
    for packet_num := 0; packet_num < full_packets; packet_num++ {
        // fmt.Printf("OAEP decrypting [%x]...\n", string(msg[packet_size*packet_num:packet_size*(packet_num+1)]))
        packet, err := rsa.DecryptOAEP(md5hash, rand.Reader, pvkey, msg[packet_size*packet_num:packet_size*(packet_num+1)], nil)
        if err != nil {
            log.Fatal("(DecryptMessage) ",err)
        } else {
            // fmt.Printf("...to [%s]\n", packet)
            decrypted_packets = append(decrypted_packets,packet)
        }
    }
    return bytes.Join(decrypted_packets,nil)
}

func ReadWebSocket(ws *websocket.Conn) []byte {
    total_bytes_bytes_encrypted := make([]byte,key_len/8)
    if _, err := ws.Read(total_bytes_bytes_encrypted); err != nil {
        log.Println("(ReadWebSocket) ",err)
    }

    md5hash := md5.New()
    total_bytes_bytes, err := rsa.DecryptOAEP(md5hash, rand.Reader, pvkey, total_bytes_bytes_encrypted, nil)
    if err != nil {
        log.Println("(ReadWebSocket) ",err)
    }

    var total_bytes uint32
    total_bytes = binary.LittleEndian.Uint32(total_bytes_bytes)

    // fmt.Printf("OAEP decrypted [%x] to \n[%d]\n",total_bytes_bytes_encrypted, total_bytes)

    msg := make([]byte, total_bytes)
    if _, err := ws.Read(msg); err != nil {
        log.Println("(ReadWebSocket) ",err)
    }
    return msg
}

func EncryptWriteWebSocket(ws *websocket.Conn, msg []byte, pbkey *rsa.PublicKey) error {
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
    // fmt.Printf("OAEP encrypted [%d] to \n[%x]\n", total_bytes, packet)

    for packet_num := 0; packet_num < full_packets; packet_num++ {
        // fmt.Printf("OAEP encrypting [%s]...\n", string(msg[packet_size*packet_num:packet_size*(packet_num+1)]))
        packet, err := rsa.EncryptOAEP(md5hash, rand.Reader, pbkey, msg[packet_size*packet_num:packet_size*(packet_num+1)], nil)
        if err != nil {
            log.Println("(EncryptWriteWebSocket) Encrypt Packet: ",err)
            return err
        } else {
            // fmt.Printf("...to [%x]\n", packet)
            encrypted_packets = append(encrypted_packets,packet)
        }
    }
    packet, err = rsa.EncryptOAEP(md5hash, rand.Reader, pbkey, msg[packet_size*full_packets:], nil)
    if err != nil {
        log.Println("(EncryptWriteWebSocket) Encrypt Remainder: ",err)
        return err
    } else {
        // fmt.Printf("OAEP encrypted [%s] to \n[%x]\n", string(msg[packet_size*full_packets:]), packet)
        encrypted_packets = append(encrypted_packets,packet)
    }
    encrypted_msg := bytes.Join(encrypted_packets,nil)
    // fmt.Printf("Encrypted Message\n%x\n",encrypted_msg)

    encrypted_msg_length := len(encrypted_msg)
    packet_size = 1024
    full_packets = encrypted_msg_length/packet_size
    remainder = encrypted_msg_length - full_packets*packet_size

    for packet_num := 0; packet_num < full_packets; packet_num++ {
        packet := encrypted_msg[packet_size*packet_num:packet_size*(packet_num+1)]
        if _, err := ws.Write(packet); err != nil {
            log.Println("(EncryptWriteWebSocket) Write Packet: ",err)
            return err
        }
        // fmt.Printf("(EncryptWriteWebSocket) Write Packet\n%x\n",packet)
    }
    if remainder > 0 {
        packet := encrypted_msg[packet_size*full_packets:]
        if _, err := ws.Write(packet); err != nil {
            log.Println("(EncryptWriteWebSocket) Write Remainder: ",err)
            return err
        }
        // fmt.Printf("(EncryptWriteWebSocket) Write Remainder\n%x\n",packet)
    }
    return nil
}

// ONLY FOR TESTING
func ServeGo(ws *websocket.Conn) {

    // Retrieve Go Client Public Key
    pbkey_encrypted := ReadWebSocket(ws)
    pbkey_decrypted := DecryptMessage(pbkey_encrypted)
    // VerifyPKCS1v15(pbkey, hash crypto.Hash, hashed []byte, sig []byte) (err error)
    pbkey_verified := pbkey_decrypted
    var go_pbkey *rsa.PublicKey
    err := json.Unmarshal(pbkey_verified,&go_pbkey)
    if err != nil {
        log.Fatal("(ServeGo) Unmarshal: ",err)
    }

    // Send Init Package to Go Client
    cws_addr := strings.Join([]string{":",strconv.Itoa(Port)},"")
    psoc_addr := strings.Join([]string{":",strconv.Itoa(Port+1)},"")
    if len(myPeerRank.m) == 0 {
        myPeerRank.m[psoc_addr] = DIRECTOR
    } else {
        myPeerRank.m[psoc_addr] = VIEWER
    }
    myPeerKeys.m[psoc_addr] = *go_pbkey
    init := Init{
        CWS_addr: cws_addr,
        PSOC_addr: psoc_addr,
        PR: myPeerRank.m,
        PK: myPeerKeys.m,
    }
    init_bytes, err := json.Marshal(init)
    if err != nil {
        log.Fatal("(ServeGo) Marshal: ",err)
    }

    if err = EncryptWriteWebSocket(ws,init_bytes,go_pbkey); err != nil {
        delete(myPeerRank.m,psoc_addr)
        delete(myPeerKeys.m,psoc_addr)
        log.Println("(ServeGo) EncryptWriteWebSocket Error : ",err)
    } else {
        JSPortList = append(JSPortList,Port)
        Port += 2
        fmt.Println("(ServeGo) Success")
        fmt.Printf("\tinit.cws_addr: %s\n",init.CWS_addr)
        fmt.Printf("\tinit.psoc_addr: %s\n",init.PSOC_addr)
        fmt.Printf("\tinit.PR: %s\n",init.PR)
        fmt.Printf("\tinit.PK: %s\n",init.PK)
    }
}

// Serve Peer List
func ServeJS(ws *websocket.Conn) {
    if JSClientCount < len(JSPortList) {
        cws_addr := strings.Join([]string{":",strconv.Itoa(JSPortList[JSClientCount])},"")
        e := json.NewEncoder(ws)
        err := e.Encode(cws_addr)
        if err != nil {
            log.Println("(ServeJS) JSON Error: ",err)
        } else {
            JSClientCount += 1
            fmt.Println("(ServeJS) Success")
            fmt.Printf("\tWS: %s\n",cws_addr)
        }
    }
}

func main() {

    // Key Length in Bits
    key_len = 1024

    // Original Code Used to Save Private Key and Public Key to File
    // pvkey, err := rsa.GenerateKey(rand.Reader, key_len)
    // pvkey_file, _ := os.Create("privategob.key")
    // e := gob.NewEncoder(pvkey_file)
    // e.Encode(pvkey)
    // pvkey_file.Close()

    // pbkey_file, _ := os.Create("publicgob.key")
    // e = gob.NewEncoder(pbkey_file)
    // e.Encode(pvkey.PublicKey)
    // pbkey_file.Close()

    // Read in Server Key Pair
    file, err := os.Open("go/src/WeTubeServer/privategob.key")
    if err != nil {
        log.Fatal(err)
    }
    d := gob.NewDecoder(file)
    err = d.Decode(&pvkey)
    pbkey = &pvkey.PublicKey

    // Setup Handlers
    fmt.Printf("(Main) Starting server at http://localhost:8080/\n")
    fmt.Printf("(Main) Issuing ports at ws://localhost:8080/ws\n")
    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go", websocket.Handler(ServeGo))
    http.Handle("/ws/js", websocket.Handler(ServeJS))

    // Start Server
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("(Main) ListenAndServe: ", err)
    }
}