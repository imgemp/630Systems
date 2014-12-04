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
  player.playVideo();
  var cmd = {
    Action: "Play",
    Argument: null,
    Target: null,
  };
  var msg = {Body: cmd, PI: myPeerInfo};
  cws.send(JSON.stringify(msg));
  console.log("(playVideo) Play")
}

function pauseVideo(): void {
  player.pauseVideo();
  var cmd = {
    Action: "Pause",
    Argument: null,
    Target: null,
  };
  var msg = {Body: cmd, PI: myPeerInfo};
  cws.send(JSON.stringify(msg));
  console.log("(pauseVideo) Pause")
}

function stopVideo(): void {
  player.stopVideo();
  var cmd = {
    Action: "Stop",
    Argument: null,
    Target: null,
  };
  var msg = {Body: cmd, PI: myPeerInfo};
  cws.send(JSON.stringify(msg));
  console.log("(stopVideo) Stop")
}

function seekTo(seconds: number): void {
  player.seekTo(seconds,true);
  var cmd = {
    Action: "SeekTo",
    Argument: seconds.toString(),
    Target: null,
  };
  var msg = {Body: cmd, PI: myPeerInfo};
  cws.send(JSON.stringify(msg));
  console.log("(seekTo) SeekTo "+seconds.toString()+" Seconds")
}

function ChangeRank(fromRank: string,toRank: string): void {
  var index = (<HTMLSelectElement>document.getElementById(fromRank)).selectedIndex;
  var option = (<HTMLSelectElement>document.getElementById(fromRank)).options[index];
  (<HTMLSelectElement>document.getElementById(fromRank)).remove(index);
  (<HTMLSelectElement>document.getElementById(toRank)).add(option);
  console.log("(ChangeRank) "+fromRank+" to "+toRank+": "+option.text);
}

function PromoteEditor(): void {
  ChangeRank('Editor','Master');
}

function DemoteMaster(): void {
  ChangeRank('Master','Editor');
}

function PromoteViewer(): void {
  ChangeRank('Viewer','Editor');
}

function DemoteEditor(): void {
  ChangeRank('Editor','Viewer')
}

function KingViewer(): void {
  ChangeRank('Viewer','Master');
}

function CrushMaster(): void {
  ChangeRank('Master','Viewer');
}

// Connect to Client WebSocket
function ClientWebSocket() {
    cws = new WebSocket(cws_addr, "protocolOne");
    cws.onopen = function (event) {
      var cmd = {Action: "NewPeer", Argument: null, Target: null};
      var msg = {Body: cmd, PI: myPeerInfo};
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

// Update myPeerInfo & HTML Ranks
function UpdatePeers(PI: any) {
  for (var addr in PI) {
    if (!myPeerInfo[addr]) {
      myPeerInfo[addr] = PI[addr];
      AddHTMLRank(addr,PI[addr]);
    }
  }
}

// Handle Peer Messages
function HandleMessage(msg: any) {
  switch(msg.Body.Action) {
    case "NewPeer":
      console.log("(HandleMessage) NewPeer");
      UpdatePeers(msg.PI)
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
      player.seekTo(msg.Body.Argument,true);
      break;
    default:
      console.log("(HandleMessage) Command Not Recognized");
  }
}

// Populate HTML Ranks on Startup
function PopulateHTMLRanks() {
  for (var addr in myPeerInfo) {
    AddHTMLRank(addr,myPeerInfo[addr]);
  }
}

// Update Single HTML Rank - might want to make this check to see if addr is already in rank list or somewhere in ranks
function AddHTMLRank(addr: string, rank: number) {
  var option = document.createElement("option");
  option.text = addr;
  switch(rank) {
  case Rank.Master:
    console.log("(UpdateHTMLRank) Adding Master");
    (<HTMLSelectElement>document.getElementById('Master')).add(option);
    break;
  case Rank.Editor:
    console.log("(UpdateHTMLRank) Adding Editor");
    (<HTMLSelectElement>document.getElementById('Editor')).add(option);
    break;
  case Rank.Viewer:
    console.log("(UpdateHTMLRank) Adding Viewer");
    (<HTMLSelectElement>document.getElementById('Viewer')).add(option);
    break;
  default:   
    console.log("(UpdateHTMLRank) Rank Not Recognized");
  }
}

enum Rank {
  Viewer = 0,
  Editor = 1,
  Master = 2
}

// Establish WebSocket Connection with WeTube (Go) Client
var cws_addr: string;
var cws: WebSocket;
var myPeerInfo: any;
var sws = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
sws.onmessage = function (event) {
  var init = JSON.parse(event.data)
  cws_addr = "ws://localhost:"+init.Port+"/ws"
  myPeerInfo = init.PI;
  PopulateHTMLRanks();
  ClientWebSocket();
  sws.close();
}