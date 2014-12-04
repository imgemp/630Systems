/// <reference path="youtube.d.ts" />
// / <reference path="PlayerControls.ts" />
// This just adds a youtube iframe to the div in the html
// 2. This code loads the IFrame Player API code asynchronously.
var tag = document.createElement('script');

tag.src = "https://www.youtube.com/iframe_api";
var firstScriptTag = document.getElementsByTagName('script')[0];
firstScriptTag.parentNode.insertBefore(tag, firstScriptTag);

// 3. This function creates an <iframe> (and YouTube player)
//    after the API code downloads.
var player;
function onYouTubeIframeAPIReady() {
    player = new YT.Player('player', {
        height: 390,
        width: 640,
        videoId: 'M7lc1UVf-VE',
        playerVars: {
            'controls': 0
        },
        events: {
            'onReady': onPlayerReady,
            'onStateChange': onPlayerStateChange
        }
    });
    console.log("Player Ready");
}

// 4. The API will call this function when the video player is ready.
function onPlayerReady(event) {
    // event.target.playVideo();
}

// 5. The API calls this function when the player's state changes.
//    The function indicates that when playing a video (state=1),
//    the player should play for six seconds and then stop.
var done = false;
function onPlayerStateChange(event) {
    // if (event.data == YT.PlayerState.PLAYING && !done) {
    //   setTimeout(stopVideo, 6000);
    //   done = true;
    // }
}

function playVideo() {
    player.playVideo();
    var cmd = {
        Action: "Play",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    ws.send(JSON.stringify(msg));
    console.log(cmd);
}

function pauseVideo() {
    player.pauseVideo();
    var cmd = {
        Action: "Pause",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    ws.send(JSON.stringify(msg));
    console.log(cmd);
}

function stopVideo() {
    player.stopVideo();
    var cmd = {
        Action: "Stop",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    ws.send(JSON.stringify(msg));
    console.log(cmd);
}

function seekTo(seconds) {
    player.seekTo(seconds, true);
    var cmd = {
        Action: "SeekTo",
        Argument: seconds.toString(),
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    ws.send(JSON.stringify(msg));
    console.log(cmd);
}

console.log("Starting WeTubeClient (JS)");

// Dial Client Websocket
function DialWebSocket(addr) {
    ws = new WebSocket(addr, "protocolOne");
    ws.onopen = function (event) {
        var cmd = { Action: "NewPeer", Argument: null, Target: null };
        var msg = { Body: cmd, PI: myPeerInfo };
        ws.send(JSON.stringify(msg));
        console.log(msg);
    };
    ws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log("Go Client: " + event.data.trim()); // this will turn into a command to be parsed and executed, should also update peer set
        HandleMessage(msg);
    };
    ws.onclose = function (event) {
        console.log("WebSocket closing...", event.code, event.reason);
    };
}

function UpdatePeers(PI) {
    for (var addr in PI) {
        if (!myPeerInfo[addr]) {
            myPeerInfo[addr] = PI[addr];
            UpdateMEVList(addr, PI[addr]);
        }
    }
}

function HandleMessage(msg) {
    switch (msg.Body.Action) {
        case "NewPeer":
            UpdatePeers(msg.PI);
            console.log("NewPeer");
            break;
        case "Play":
            player.playVideo();
            console.log("Play");
            break;
        case "Pause":
            player.pauseVideo();
            console.log("Pause");
            break;
        case "Stop":
            player.stopVideo();
            console.log("Stop");
            break;
        case "SeekTo":
            player.seekTo(msg.Body.Argument, true);
            console.log("SeekTo");
            break;
        default:
            console.log("Command Not Recognized");
    }
}

function PopulateMEVLists(PI) {
    for (var addr in PI) {
        UpdateMEVList(addr, PI[addr]);
    }
}

function UpdateMEVList(addr, rank) {
    var option = document.createElement("option");
    option.text = addr;
    console.log(addr);
    switch (rank) {
        case 2 /* Master */:
            var MList = document.getElementById('Master');
            MList.add(option);
            console.log("Adding Master");
            break;
        case 1 /* Editor */:
            var EList = document.getElementById('Editor');
            EList.add(option);
            console.log("Adding Editor");
            break;
        case 0 /* Viewer */:
            var VList = document.getElementById('Viewer');
            VList.add(option);
            console.log("Adding Viewer");
            break;
        default:
            console.log("Rank Not Recognized");
    }
}

var Rank;
(function (Rank) {
    Rank[Rank["Viewer"] = 0] = "Viewer";
    Rank[Rank["Editor"] = 1] = "Editor";
    Rank[Rank["Master"] = 2] = "Master";
})(Rank || (Rank = {}));

// Establish WebSocket Connection with WeTube (Go) Client
var myLocalWebSocketAddr;
var ws;
var myPeerInfo;
var tempWebSocket = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
tempWebSocket.onopen = function (event) {
    // tempWebSocket.send("Which port should I use?");
    console.log("Which port should I use?");
};
tempWebSocket.onmessage = function (event) {
    var init = JSON.parse(event.data);
    console.log("WeTubeServer: Use port " + init.Port);
    console.log("Connecting to websocket at ws://localhost:" + init.Port + "/ws");
    myLocalWebSocketAddr = "ws://localhost:" + init.Port + "/ws";
    console.log(init.PI);
    myPeerInfo = init.PI;
    PopulateMEVLists(myPeerInfo);
    DialWebSocket(myLocalWebSocketAddr);
    tempWebSocket.close();
};
