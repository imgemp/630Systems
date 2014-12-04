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
    console.log("(onYouTubeIframeAPIReady) Player Ready");
}
// 4. The API will call this function when the video player is ready.
function onPlayerReady(event) {
}
// 5. The API calls this function when the player's state changes.
//    The function indicates that when playing a video (state=1),
//    the player should play for six seconds and then stop.
var done = false;
function onPlayerStateChange(event) {
}
// OnClick Video Commands
function playVideo() {
    player.playVideo();
    var cmd = {
        Action: "Play",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    cws.send(JSON.stringify(msg));
    console.log("(playVideo) Play");
}
function pauseVideo() {
    player.pauseVideo();
    var cmd = {
        Action: "Pause",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    cws.send(JSON.stringify(msg));
    console.log("(pauseVideo) Pause");
}
function stopVideo() {
    player.stopVideo();
    var cmd = {
        Action: "Stop",
        Argument: null,
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    cws.send(JSON.stringify(msg));
    console.log("(stopVideo) Stop");
}
function seekTo(seconds) {
    player.seekTo(seconds, true);
    var cmd = {
        Action: "SeekTo",
        Argument: seconds.toString(),
        Target: null
    };
    var msg = { Body: cmd, PI: myPeerInfo };
    cws.send(JSON.stringify(msg));
    console.log("(seekTo) SeekTo " + seconds.toString() + " Seconds");
}
function ChangeRank(fromRank, toRank) {
    var index = document.getElementById(fromRank).selectedIndex;
    var option = document.getElementById(fromRank).options[index];
    document.getElementById(fromRank).remove(index);
    document.getElementById(toRank).add(option);
    console.log("(ChangeRank) " + fromRank + " to " + toRank + ": " + option.text);
}
function PromoteEditor() {
    ChangeRank('Editor', 'Master');
}
function DemoteMaster() {
    ChangeRank('Master', 'Editor');
}
function PromoteViewer() {
    ChangeRank('Viewer', 'Editor');
}
function DemoteEditor() {
    ChangeRank('Editor', 'Viewer');
}
function KingViewer() {
    ChangeRank('Viewer', 'Master');
}
function CrushMaster() {
    ChangeRank('Master', 'Viewer');
}
// Connect to Client WebSocket
function ClientWebSocket() {
    cws = new WebSocket(cws_addr, "protocolOne");
    cws.onopen = function (event) {
        var cmd = { Action: "NewPeer", Argument: null, Target: null };
        var msg = { Body: cmd, PI: myPeerInfo };
        cws.send(JSON.stringify(msg));
        console.log("(ClientWebSocket/onopen)");
        console.log(msg);
    };
    cws.onmessage = function (event) {
        var msg = JSON.parse(event.data);
        console.log("(ClientWebSocket/onmessage) " + event.data.trim());
        HandleMessage(msg);
    };
    cws.onclose = function (event) {
        console.log("(ClientWebSocket) WebSocket Closing...", event.code, event.reason);
    };
}
// Update myPeerInfo & HTML Ranks
function UpdatePeers(PI) {
    for (var addr in PI) {
        if (!myPeerInfo[addr]) {
            myPeerInfo[addr] = PI[addr];
            AddHTMLRank(addr, PI[addr]);
        }
    }
}
// Handle Peer Messages
function HandleMessage(msg) {
    switch (msg.Body.Action) {
        case "NewPeer":
            console.log("(HandleMessage) NewPeer");
            UpdatePeers(msg.PI);
            break;
        case "Play":
            console.log("(HandleMessage) Play");
            player.playVideo();
            break;
        case "Pause":
            console.log("(HandleMessage) Pause");
            player.pauseVideo();
            break;
        case "Stop":
            console.log("(HandleMessage) Stop");
            player.stopVideo();
            break;
        case "SeekTo":
            console.log("(HandleMessage) SeekTo");
            player.seekTo(msg.Body.Argument, true);
            break;
        default:
            console.log("(HandleMessage) Command Not Recognized");
    }
}
// Populate HTML Ranks on Startup
function PopulateHTMLRanks() {
    for (var addr in myPeerInfo) {
        AddHTMLRank(addr, myPeerInfo[addr]);
    }
}
// Update Single HTML Rank - might want to make this check to see if addr is already in rank list or somewhere in ranks
function AddHTMLRank(addr, rank) {
    var option = document.createElement("option");
    option.text = addr;
    switch (rank) {
        case 2 /* Master */:
            console.log("(UpdateHTMLRank) Adding Master");
            document.getElementById('Master').add(option);
            break;
        case 1 /* Editor */:
            console.log("(UpdateHTMLRank) Adding Editor");
            document.getElementById('Editor').add(option);
            break;
        case 0 /* Viewer */:
            console.log("(UpdateHTMLRank) Adding Viewer");
            document.getElementById('Viewer').add(option);
            break;
        default:
            console.log("(UpdateHTMLRank) Rank Not Recognized");
    }
}
var Rank;
(function (Rank) {
    Rank[Rank["Viewer"] = 0] = "Viewer";
    Rank[Rank["Editor"] = 1] = "Editor";
    Rank[Rank["Master"] = 2] = "Master";
})(Rank || (Rank = {}));
// Establish WebSocket Connection with WeTube (Go) Client
var cws_addr;
var cws;
var myPeerInfo;
var sws = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
sws.onmessage = function (event) {
    var init = JSON.parse(event.data);
    cws_addr = "ws://localhost:" + init.Port + "/ws";
    myPeerInfo = init.PI;
    PopulateHTMLRanks();
    ClientWebSocket();
    sws.close();
};
