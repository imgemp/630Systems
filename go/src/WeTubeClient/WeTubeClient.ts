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
function playVideo(): void {
  var cmd = {
    Action: "Play",
    Argument: null,
    Target: null,
  };
  var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
  Seen[msg.ID] = true
  cws.send(JSON.stringify(msg));
  console.log("(playVideo) Play")
}

function pauseVideo(): void {
  var cmd = {
    Action: "Pause",
    Argument: null,
    Target: null,
  };
  var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
  Seen[msg.ID] = true
  cws.send(JSON.stringify(msg));
  console.log("(pauseVideo) Pause")
}

function stopVideo(): void {
  var cmd = {
    Action: "Stop",
    Argument: null,
    Target: null,
  };
  var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
  Seen[msg.ID] = true
  cws.send(JSON.stringify(msg));
  console.log("(stopVideo) Stop")
}

function seekTo(seconds: number): void {
  var cmd = {
    Action: "SeekTo",
    Argument: seconds.toString(),
    Target: null,
  };
  var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
  Seen[msg.ID] = true
  cws.send(JSON.stringify(msg));
  console.log("(seekTo) SeekTo "+seconds.toString()+" Seconds")
}

function ChangeRankHTML(fromRank: string,toRank: string,index: number): void {
  if (index > 0) {
    var option = (<HTMLSelectElement>document.getElementById(fromRank)).options[index];
    (<HTMLSelectElement>document.getElementById(fromRank)).remove(index);
    (<HTMLSelectElement>document.getElementById(toRank)).add(option);
    myPeerRank[option.text] = Str2Rank(toRank)
    myPeerIndex[option.text] = (<HTMLSelectElement>document.getElementById(toRank)).length-1
    console.log("(ChangeRankHTML) "+fromRank+" to "+toRank+": "+option.text);
  }
}

function PromoteEditor(): void {
  var index = (<HTMLSelectElement>document.getElementById("Editor")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Editor")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Director",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(PromoteEditor) Editor->Director")
  }
}

function DemoteDirector(): void {
  var index = (<HTMLSelectElement>document.getElementById("Director")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Director")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Editor",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(DemoteDirector) Director->Editor")
  }
}

function PromoteViewer(): void {
  var index = (<HTMLSelectElement>document.getElementById("Viewer")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Viewer")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Editor",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(PromoteViewer) Viewer->Editor");
  }
}

function DemoteEditor(): void {
  var index = (<HTMLSelectElement>document.getElementById("Editor")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Editor")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Viewer",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(DemoteEditor) Editor->Viewer")
  }
}

function KingViewer(): void {
  var index = (<HTMLSelectElement>document.getElementById("Viewer")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Viewer")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Director",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(KingViewer) Viewer->Director")
  }
}

function CrushDirector(): void {
  var index = (<HTMLSelectElement>document.getElementById("Director")).selectedIndex;
  if (index > 0) {
    var addr: string = (<HTMLSelectElement>document.getElementById("Director")).options[index].text;
    var cmd = {
      Action: "ChangeRank",
      Argument: "Viewer",
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(CrushDirector) Director->Viewer")
  }
}

// Connect to Client WebSocket
function ClientWebSocket(): void {
    cws = new WebSocket(cws_addr, "protocolOne");
    cws.onopen = function (event) {
      var cmd = {Action: "NewPeer", Argument: null, Target: null};
      var msg = {Body: cmd, PR: myPeerRank};
      cws.send(JSON.stringify(msg));
      console.log("(ClientWebSocket/onopen)");
      console.log(msg);
    };
    cws.onmessage = function (event) {
      var msg = JSON.parse(event.data)
      console.log("(ClientWebSocket/onmessage) "+event.data.trim());
      HandleMessage(msg);
    }
    cws.onclose = function (event) {
      console.log("(ClientWebSocket) WebSocket Closing...",event.code,event.reason);
    }
}

// Update myPeerRank & HTML Ranks
function UpdatePeers(PR: any): void {
  for (var addr in PR) {
    if (!(addr in myPeerRank)) {
      myPeerRank[addr] = PR[addr];
      var index = AddHTMLRank(addr,PR[addr]);
      if (index >= 0) {
        myPeerIndex[addr] = index
      }
    }
  }
}

function Rank2Str(rank: number): string {
  switch(rank) {
    case Rank.Director:
      return "Director"
    case Rank.Editor:
      return "Editor"
    case Rank.Viewer:
      return "Viewer"
    default:
      return null
  }
}

function Str2Rank(rank: string): number {
  switch(rank) {
    case "Director":
      return Rank.Director
    case "Editor":
      return Rank.Editor
    case "Viewer":
      return Rank.Viewer
    default:
      return null
  }
}

function RemovePeerHTML(addr: string) {
  var rank = myPeerRank[addr]
  var index = myPeerIndex[addr]
  for (var other_addr in myPeerRank) {
    if ((myPeerRank[other_addr] == rank) && (myPeerIndex[other_addr] > index)) {
      myPeerIndex[other_addr] -= 1
    }
  }
  (<HTMLSelectElement>document.getElementById(Rank2Str(rank))).remove(index);
}

function DropPeer(addr: string) {
  if (addr in myPeerRank) {
    RemovePeerHTML(addr)
    delete myPeerRank[addr]
    delete myPeerIndex[addr]
    var cmd = {
      Action: "DropPeer",
      Argument: null, // put vote here?
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(DropPeer) "+addr);
  }
}

function HandleMessage(msg: any): void {
  if (msg.Body.Action == "NewPeer") {
    console.log("(HandleMessage) NewPeer");
    UpdatePeers(msg.PR);
  } else if (msg.Body.Action == "DropPeer") {
    console.log("(HandleMessage) DropPeer");
    DropPeer(msg.Body.Target);
    // check for empty director list and run election
  } else if (Seen[msg.ID]) {
    switch(msg.Body.Action) {
      case "Play":
        console.log("(HandleMessage/Boomerang) Play");
        cws.send(JSON.stringify(msg));
        player.playVideo();
        break;
      case "Pause":
        console.log("(HandleMessage/Boomerang) Pause");
        cws.send(JSON.stringify(msg));
        player.pauseVideo();
        break;
      case "Stop":
        console.log("(HandleMessage/Boomerang) Stop");
        cws.send(JSON.stringify(msg));
        player.stopVideo();
        break;
      case "SeekTo":
        console.log("(HandleMessage/Boomerang) SeekTo");
        cws.send(JSON.stringify(msg));
        player.seekTo(msg.Body.Argument,true);
        break;
      case "ChangeRank":
        console.log("(HandleMessage/Boomerang/ChangeRank) *Mote Editor");
        ChangeRankHTML(Rank2Str(myPeerRank[msg.Body.Target]),msg.Body.Argument,myPeerIndex[msg.Body.Target])
        cws.send(JSON.stringify(msg));
        break;
      default:
        console.log("(HandleMessage/Boomerang) Command "+msg.Body.Action+" Not Recognized");
    }
  } else {
    switch(myPeerRank[msg.Addr]) {
      case Rank.Director:
        switch(msg.Body.Action) {
          case "Play":
            console.log("(HandleMessage/Director) Play");
            cws.send(JSON.stringify(msg));
            player.playVideo();
            break;
          case "Pause":
            console.log("(HandleMessage/Director) Pause");
            cws.send(JSON.stringify(msg));
            player.pauseVideo();
            break;
          case "Stop":
            console.log("(HandleMessage/Director) Stop");
            cws.send(JSON.stringify(msg));
            player.stopVideo();
            break;
          case "SeekTo":
            console.log("(HandleMessage/Director) SeekTo");
            cws.send(JSON.stringify(msg));
            player.seekTo(msg.Body.Argument,true);
            break;
          case "ChangeRank":
            switch(myPeerRank[msg.Body.Target]) {
              case Rank.Director:
                console.log("(HandleMessage/Director/ChangeRank) Director Rank Cannot Be Changed");
                break;
              case Rank.Editor:
                console.log("(HandleMessage/Director/ChangeRank) *Mote Editor");
                ChangeRankHTML("Editor",msg.Body.Argument,myPeerIndex[msg.Body.Target])
                cws.send(JSON.stringify(msg));
                break;
              case Rank.Viewer:
                console.log("(HandleMessage/Director/ChangeRank) *Mote Viewer");
                ChangeRankHTML("Viewer",msg.Body.Argument,myPeerIndex[msg.Body.Target])
                cws.send(JSON.stringify(msg));
                break;
              default:
                console.log("(HandleMessage/Director/ChangeRank) My Rank Of "+myPeerRank[msg.Body.Target]+" Not Recognized");
            }
            break;
          default:
            console.log("(HandleMessage/Director) "+msg.Body.Action+" Command Not Recognized");
        }
        break;
      case Rank.Editor:
        switch(msg.Body.Action) {
          case "Play":
            console.log("(HandleMessage/Editor) Play");
            cws.send(JSON.stringify(msg));
            player.playVideo();
            break;
          case "Pause":
            console.log("(HandleMessage/Editor) Pause");
            cws.send(JSON.stringify(msg));
            player.pauseVideo();
            break;
          case "Stop":
            console.log("(HandleMessage/Editor) Stop");
            cws.send(JSON.stringify(msg));
            player.stopVideo();
            break;
          case "SeekTo":
            console.log("(HandleMessage/Editor) SeekTo");
            cws.send(JSON.stringify(msg));
            player.seekTo(msg.Body.Argument,true);
            break;
          case "ChangeRank":
            console.log("(HandleMessage/Editor) Editor Cannot Change Rank");
            break;
          default:
            console.log("(HandleMessage/Editor) "+msg.Body.Action+" Command Not Recognized");
        }
        break;
      case Rank.Viewer:
        switch(msg.Body.Action) {
          case "Play":
          case "Pause":
          case "Stop":
          case "SeekTo":
          case "ChangeRank":
            console.log("(HandleMessage/Viewer) Viewer Can Only Watch");
            break;
          default:
            console.log("(HandleMessage/Viewer) Command "+msg.Body.Action+" Not Recognized");
        }
        break;
      default:
        console.log("(HandleMessage) Rank Of "+myPeerRank[msg.Addr]+" Not Recognized");
    }
  }
}

// Update Single HTML Rank - might want to make this check to see if addr is already in rank list or somewhere in ranks
function AddHTMLRank(addr: string, rank: number): number {
  var option = document.createElement("option");
  option.text = addr;
  switch(rank) {
  case Rank.Director:
    console.log("(AddHTMLRank) Adding Director");
    (<HTMLSelectElement>document.getElementById('Director')).add(option);
    return (<HTMLSelectElement>document.getElementById('Director')).length-1
    break;
  case Rank.Editor:
    console.log("(AddHTMLRank) Adding Editor");
    (<HTMLSelectElement>document.getElementById('Editor')).add(option);
    return (<HTMLSelectElement>document.getElementById('Editor')).length-1
    break;
  case Rank.Viewer:
    console.log("(AddHTMLRank) Adding Viewer");
    (<HTMLSelectElement>document.getElementById('Viewer')).add(option);
    return (<HTMLSelectElement>document.getElementById('Viewer')).length-1
    break;
  default:   
    console.log("(AddHTMLRank) Rank Not Recognized");
    return -1
  }
}

enum Rank {
  Viewer = 0,
  Editor = 1,
  Director = 2
}

// Establish WebSocket Connection with WeTube (Go) Client
var cws_addr: string;
var cws: WebSocket;
var psoc: string;
var myPeerRank: any = {};
var myPeerIndex: any = {};
var Seen: any = {};
var sws = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
sws.onmessage = function (event) {
  cws_addr = "ws://localhost"+JSON.parse(event.data)+"/ws"
  psoc = ":"+(parseInt(JSON.parse(event.data).slice(1))+1).toString()
  document.getElementById('psoc').innerHTML = psoc;
  console.log(psoc)
  ClientWebSocket();
  sws.close();
}