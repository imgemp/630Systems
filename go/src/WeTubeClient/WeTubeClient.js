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
    var msg = { Body: cmd, PR: myPeerRank, Addr: null };
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
    var msg = { Body: cmd, PR: myPeerRank, Addr: null };
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
    var msg = { Body: cmd, PR: myPeerRank, Addr: null };
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
    var msg = { Body: cmd, PR: myPeerRank, Addr: null };
    cws.send(JSON.stringify(msg));
    console.log("(seekTo) SeekTo " + seconds.toString() + " Seconds");
}

function ChangeRankHTML(fromRank, toRank, index) {
    if (index > 0) {
        var option = document.getElementById(fromRank).options[index];
        document.getElementById(fromRank).remove(index);
        document.getElementById(toRank).add(option);
        myPeerRank[option.text] = Str2Rank(toRank);
        myPeerIndex[option.text] = document.getElementById(toRank).length - 1;
        console.log("(ChangeRankHTML) " + fromRank + " to " + toRank + ": " + option.text);
    }
}

function PromoteEditor() {
    var index = document.getElementById("Editor").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Editor").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Director",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(PromoteEditor) Editor->Director");
        ChangeRankHTML('Editor', 'Director', index);
    }
}

function DemoteDirector() {
    var index = document.getElementById("Director").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Director").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Editor",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(DemoteDirector) Director->Editor");
        ChangeRankHTML('Director', 'Editor', index);
    }
}

function PromoteViewer() {
    var index = document.getElementById("Viewer").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Viewer").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Editor",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(PromoteViewer) Viewer->Editor");
        ChangeRankHTML('Viewer', 'Editor', index);
    }
}

function DemoteEditor() {
    var index = document.getElementById("Editor").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Editor").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Viewer",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(DemoteEditor) Editor->Viewer");
        ChangeRankHTML('Editor', 'Viewer', index);
    }
}

function KingViewer() {
    var index = document.getElementById("Viewer").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Viewer").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Director",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(KingViewer) Viewer->Director");
        ChangeRankHTML('Viewer', 'Director', index);
    }
}

function CrushDirector() {
    var index = document.getElementById("Director").selectedIndex;
    if (index > 0) {
        var addr = document.getElementById("Director").options[index].text;
        var cmd = {
            Action: "ChangeRank",
            Argument: "Viewer",
            Target: addr
        };
        var msg = { Body: cmd, PR: myPeerRank, Addr: null };
        cws.send(JSON.stringify(msg));
        console.log("(CrushDirector) Director->Viewer");
        ChangeRankHTML('Director', 'Viewer', index);
    }
}

// Connect to Client WebSocket
function ClientWebSocket() {
    cws = new WebSocket(cws_addr, "protocolOne");
    cws.onopen = function (event) {
        var cmd = { Action: "NewPeer", Argument: null, Target: null };
        var msg = { Body: cmd, PR: myPeerRank };
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

// Update myPeerRank & HTML Ranks
function UpdatePeers(PR) {
    for (var addr in PR) {
        if (!myPeerRank[addr]) {
            myPeerRank[addr] = PR[addr];
            var index = AddHTMLRank(addr, PR[addr]);
            if (index >= 0) {
                myPeerIndex[addr] = index;
            }
        }
    }
}

function Rank2Str(rank) {
    switch (rank) {
        case 2 /* Director */:
            return "Director";
        case 1 /* Editor */:
            return "Editor";
        case 0 /* Viewer */:
            return "Viewer";
        default:
            return null;
    }
}

function Str2Rank(rank) {
    switch (rank) {
        case "Director":
            return 2 /* Director */;
        case "Editor":
            return 1 /* Editor */;
        case "Viewer":
            return 0 /* Viewer */;
        default:
            return null;
    }
}

// Handle Peer Messages
function HandleMessage(msg) {
    switch (msg.Body.Action) {
        case "NewPeer":
            console.log("(HandleMessage) NewPeer");
            UpdatePeers(msg.PR);
            break;
        case "Play":
            if (myPeerRank[msg.Addr] > 0) {
                console.log("(HandleMessage) Play");
                player.playVideo();
            }
            break;
        case "Pause":
            if (myPeerRank[msg.Addr] > 0) {
                console.log("(HandleMessage) Pause");
                player.pauseVideo();
            }
            break;
        case "Stop":
            if (myPeerRank[msg.Addr] > 0) {
                console.log("(HandleMessage) Stop");
                player.stopVideo();
            }
            break;
        case "SeekTo":
            if (myPeerRank[msg.Addr] > 0) {
                console.log("(HandleMessage) SeekTo");
                player.seekTo(msg.Body.Argument, true);
            }
            break;
        case "ChangeRank":
            if (myPeerRank[msg.Addr] > 1) {
                console.log("(HandleMessage) ChangeRank");
                ChangeRankHTML(Rank2Str(myPeerRank[msg.Body.Target]), msg.Body.Argument, myPeerIndex[msg.Body.Target]);
            }
            break;
        default:
            console.log("(HandleMessage) Command Not Recognized");
    }
}

// Populate HTML Ranks on Startup
function PopulateHTMLRanks() {
    for (var addr in myPeerRank) {
        var index = AddHTMLRank(addr, myPeerRank[addr]);
        if (index >= 0) {
            myPeerIndex[addr] = index;
        }
    }
}

// Update Single HTML Rank - might want to make this check to see if addr is already in rank list or somewhere in ranks
function AddHTMLRank(addr, rank) {
    var option = document.createElement("option");
    option.text = addr;
    switch (rank) {
        case 2 /* Director */:
            console.log("(AddHTMLRank) Adding Director");
            document.getElementById('Director').add(option);
            return document.getElementById('Director').length - 1;
            break;
        case 1 /* Editor */:
            console.log("(AddHTMLRank) Adding Editor");
            document.getElementById('Editor').add(option);
            return document.getElementById('Director').length - 1;
            break;
        case 0 /* Viewer */:
            console.log("(AddHTMLRank) Adding Viewer");
            document.getElementById('Viewer').add(option);
            return document.getElementById('Director').length - 1;
            break;
        default:
            console.log("(AddHTMLRank) Rank Not Recognized");
            return -1;
    }
}

var Rank;
(function (Rank) {
    Rank[Rank["Viewer"] = 0] = "Viewer";
    Rank[Rank["Editor"] = 1] = "Editor";
    Rank[Rank["Director"] = 2] = "Director";
})(Rank || (Rank = {}));

// Establish WebSocket Connection with WeTube (Go) Client
var cws_addr;
var cws;
var myPeerRank = {};
var myPeerIndex = {};
var sws = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
sws.onmessage = function (event) {
    cws_addr = "ws://localhost" + JSON.parse(event.data) + "/ws";

    // PopulateHTMLRanks();
    ClientWebSocket();
    sws.close();
};
