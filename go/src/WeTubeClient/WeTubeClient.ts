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
    for (var other_addr in myPeerRank) {
      if ((myPeerRank[other_addr] == Str2Rank(fromRank)) && (myPeerIndex[other_addr] > index)) {
        myPeerIndex[other_addr] -= 1
      }
    }
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
      Argument: null,
      Target: addr,
    };
    var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
    Seen[msg.ID] = true
    cws.send(JSON.stringify(msg));
    console.log("(DropPeer) "+addr);
    if ((<HTMLSelectElement>document.getElementById('Director')).length < 2) { // length is 1 at minimum for "Director" header
      if (CountVoters() < 2) {
        ChangeRankHTML(Rank2Str(myPeerRank[psoc]),"Director",myPeerIndex[psoc]);
      } else {
        StartElection();
      }
    }
  }
}

function PeerList(): string[] {
  var PL = []
  for (var addr in myPeerRank) {
    PL.push(addr);
  }
  return PL
}

function MakeVote(): string {
  var PL = PeerList();
  var randomPick = Math.floor((Math.random() * PL.length) + 0);
  var myVote = PL[randomPick];
  return myVote
}

function CountVotes(): number {
  var count = 0;
  for (var addr in Votes) {
    count += 1;
  }
  return count
}

function CountVoters(): number {
  var count = 0;
  for (var addr in myPeerRank) {
    count += 1;
  }
  return count
}

function StartElection(): void {
  Votes = {};
  var numVoters = (<HTMLSelectElement>document.getElementById('Director')).length + (<HTMLSelectElement>document.getElementById('Editor')).length + (<HTMLSelectElement>document.getElementById('Viewer')).length - 3;
  var myVote = MakeVote();
  Votes[psoc] = myVote;
  var cmd = {
    Action: "Vote",
    Argument: psoc,
    Target: myVote,
  };
  var msg = {ID: Math.random().toString(), Body: cmd, PR: myPeerRank, Addr: null};
  Seen[msg.ID] = true
  cws.send(JSON.stringify(msg));
  console.log("(StartElection) myVote: "+myVote);
}

function isMajority(): any[] {
  var candidates: any = {};
  var winner: string = psoc;
  candidates[winner] = 0;
  var maj: boolean = false;
  for (var voter in Votes) {
    var candidate = Votes[voter];
    if (candidates[candidate] == undefined) {
      candidates[candidate] = 1;
    } else {
      candidates[candidate] += 1;
    }
    if (candidates[candidate] > candidates[winner]) {
      maj = true;
      winner = candidate;
    } else if (candidate == winner) {
      maj = true
    } else if (candidates[candidate] == candidates[winner]) {
      maj = true; // break ties with larger address
      if (parseInt(candidate.slice(1)) > parseInt(winner.slice(1))) {
        winner = candidate;
      }
    }
  }
  return [maj,winner]
}

function Vote(cmd: any): void {
  if (!(cmd.Argument in Votes)) {
    Votes[cmd.Argument] = cmd.Target;
  }
  if (!(psoc in Votes)) {
    Votes[psoc] = MakeVote();
  }
  if (CountVotes() == CountVoters()) {
    console.log("Votes Are In");
    console.log(Votes);
    var result = isMajority();
    var isMaj = result[0];
    if (isMaj) {
      var winner = result[1];
      console.log("Winner is "+winner);
      ChangeRankHTML(Rank2Str(myPeerRank[winner]),"Director",myPeerIndex[winner]);
    } else {
      StartElection();
    }
  }
}

function HandleMessage(msg: any): void {
  if (msg.Body.Action == "NewPeer") {
    console.log("(HandleMessage) NewPeer");
    UpdatePeers(msg.PR);
  } else if (msg.Body.Action == "DropPeer") {
    console.log("(HandleMessage) DropPeer");
    DropPeer(msg.Body.Target);
  } else if (msg.Body.Action == "Vote") {
    console.log("(HandleMessage) Vote");
    Vote(msg.Body);
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
          console.log("(HandleMessage/Boomerang/ChangeRank) *Mote ",Rank2Str(myPeerRank[msg.Body.Target]));
          ChangeRankHTML(Rank2Str(myPeerRank[msg.Body.Target]),msg.Body.Argument,myPeerIndex[msg.Body.Target]);
          cws.send(JSON.stringify(msg));
          if (msg.Body.Argument == "Director") {
            Votes = {};
          }
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
                if (msg.Body.Argument == "Director") {
                  Votes = {};
                }
                break;
              case Rank.Viewer:
                console.log("(HandleMessage/Director/ChangeRank) *Mote Viewer");
                ChangeRankHTML("Viewer",msg.Body.Argument,myPeerIndex[msg.Body.Target])
                cws.send(JSON.stringify(msg));
                if (msg.Body.Argument == "Director") {
                  Votes = {};
                }
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
    Votes = {};
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
var Votes: any = {};
var sws = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
sws.onmessage = function (event) {
  cws_addr = "ws://localhost"+JSON.parse(event.data)+"/ws"
  psoc = ":"+(parseInt(JSON.parse(event.data).slice(1))+1).toString()
  document.getElementById('psoc').innerHTML += psoc;
  console.log(psoc)
  ClientWebSocket();
  sws.close();
}