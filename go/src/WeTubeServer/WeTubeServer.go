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
    "math/big"
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

    pvkey rsa.PrivateKey
    pbkey rsa.PublicKey

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

    // Hard-Code Key Pair
    pbkey.N = big.NewInt(3135329303)
    pbkey.E = 65537
    pvkey.D = big.NewInt(2810295473)
    pvkey.Primes = []*big.Int{big.NewInt(58601),big.NewInt(53503)}
    pvkey.PublicKey = pbkey

    // Setup Handlers
    fmt.Printf("(Main) Starting server at http://localhost:8080/\n")
    fmt.Printf("(Main) Issuing ports at ws://localhost:8080/ws\n")
    http.Handle("/", http.FileServer(http.Dir("./go/src/WeTubeClient/")))
    http.Handle("/ws/go", websocket.Handler(ServeGo))
    http.Handle("/ws/js", websocket.Handler(ServeJS))

    // Start Server
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("(Main) ListenAndServe: ", err)
    }
}

// -----BEGIN RSA PRIVATE KEY-----
// MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
// wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
// 1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
// 3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
// pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
// GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
// AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
// L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
// X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
// U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
// 37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
// -----END RSA PRIVATE KEY-----

// -----BEGIN PUBLIC KEY-----
// MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0
// FPqri0cb2JZfXJ/DgYSF6vUpwmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/
// 3j+skZ6UtW+5u09lHNsj6tQ51s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQAB
// -----END PUBLIC KEY-----