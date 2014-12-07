WeTube
==========

Ian Gemp
----------

##Description
WeTube is a peer to peer system implemented in Go that allows multiple users to watch YouTube videos synchronously (see wetubedescription.pdf).  This implementation is meant for academic purposes only; it is not production ready.  The controls available to the users consist of only play, pause, stop, seek, and role promotion/demotion.  The player only plays one video.  The peers are simulated as different ports.
  
##System Requirements
This system requires the user to install Go.  All testing was done in the Google Chrome browser using Go version 1.3.3 darwin/amd64 for Mac OSX.  I expect the browser interface to only be limited by the YouTube API since the rest of the html and javascript is fairly simple.

##Dependencies
**Go**:

You must have Go installed.  See [Go](https://golang.org/doc/install).

##Usage

Steps:

*Setup*

- export PATH=$PATH:$GOROOT/bin (where $GOROOT is the path to your Go installation e.g. /usr/local/go)
- export GOPATH=$HOME/go (where $HOME is the path to the local clone of this repo)
- export PATH=$PATH:$GOPATH/bin

*Run*

- Open up four bash shells (Terminals)
- In each shell, cd $GOPATH/..
- In one shell, WeTubeServer  `enter`
- In the other 3 shells, WeTubeClient `enter`
- Open up a browser (preferably Chrome)
- Open up 3 tabs and go to [http://localhost:8080/](http://localhost:8080/) in each one

You should now be ready to try out any of the features described below.  Useful comments are printed to the browser javascript console and shells respectively (e.g. message contents w/ keys, user requests/commands, election results, etc.).

###User Priveleges
Privilege levels separate the users into different levels of control.  Viewers are at the bottom of the hierarchy and Directors are at the top.  Each successive level adds new abilities.  Use the arrows next to the select fields to change the privilege levels of peers.
- Viewers - View Only
- Editors - Video Controls
- Directors - Privilege Manipulation

###Encryption/Decryption + Signatures
All messages are encrypted with [RSA](http://golang.org/pkg/crypto/rsa/)-OAEP.  The server public key is saved to file in both the WeTubeClient and WeTubeServer (although unnecessary) directories and is loaded at startup.  All messages between peers are signed with RSASSA-PSS.  The server is assumed to be trusted; no certificate or signatures exist in the initial client-server handoff.  Due to message length limits set by Go's RSA package, messages are broken up and encrypted individually and then spliced back together.  The encrypted message length itself is encrypted and prepended to the encrypted message to enable decryption by peers.

###Elections
If all the directors of the video are dropped, a new director is elected.  The peer with the largest port address is selected in the case of a tie.  Although a single peer will be upgraded to Director status if all other peers drop, the last man standing will be unable to control the video due to the message passing paradigm.  To test this feature, simply abort (`ctr-c`) a director's Go client and perform an operation with any one of the remaining WeTube peers (play, promote viewer, etc.).

##Known Bugs
Encryption/Decryption using Go presented a number of complications, especially when combined with added bugs in the Go [websocket](https://godoc.org/golang.org/x/net/websocket) package.  For some reasons I have not uncovered yet, the RSA-OAEP encryption/decryption functions complain of exceeding message length even after numerous efforts to present the functions with small constant size packets.  This error causes the system to reject anything more than 3 peers (the 4th peer is never able to download the set of public keys from the server due to the issue described above).

There are probably other bugs lurking elsewhere, but it's hard to test for them when you can only run 3 peers at once.

##WebSocket & TCP Communication
The browser javascript client and native Go client communicate through a bare (not encrypted) WebSocket connection.  Since both these parties run natively, it's assumed that they do not present any security vulnerability.  The Go client conducts a single handshake with the HTTP server through a WebSocket simply to obtain the set of peer public keys (and port addresses in this case).  All communication between peers is conducted through the Go clients using tcp.  Since the WebSockets present some issues when reading/writing large byte arrays, messages are broken up into more manageable packets.

##Message Passing
Users are required to be allowed to attempt any action they wish in their browser irrespective of their privilege level.  For this reason, I could not simply disable buttons based on privelege levels.  Instead, I used a message passing paradigm where all user actions (button clicks, etc.) are interpreted as outgoing requests from the user to the peer pool.  The requests are then validated individually by each peer and returned to the user if the action was legal.  In this sense, all incoming actions are commands.  The exceptions to this paradigm are elections/votes, new peer introductions, and dropped connection alerts.

##Channels & Locks
Within Go, channels are used extensively to take advantage of Go's asynchronous subroutines (`go foo()`).  One channel is dedicated to distributing incoming messages to the client, another is used for storing outgoing messages from the client, and set of channels (for each peer) is dedicated to distributing messages concurrently to all peers.

Locks are used to maintain the integrity of various peer information maps, the peer channels, and the seen map (to track duplicate messages).  This is necessary in the presence of the concurrent Go processes and endless for loops.

##Acknowledgements
This project was delegated by Emery Berger as Project #2 of the 630 Systems course at UMass Amherst (Fall 2014).
