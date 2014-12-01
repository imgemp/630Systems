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

function playVideo(): void {
  player.playVideo();
  // sendToWebSocket(myLocalWebSocketAddr,"Play");
}

function pauseVideo(): void {
  player.pauseVideo();
  // sendToWebSocket(myLocalWebSocketAddr,"Pause");
}

function stopVideo(): void {
  player.stopVideo();
  var msg = {
    command: "Stop",
    argument: null
  };
  var json_msg = JSON.stringify(msg)
  sendToWebSocket(myLocalWebSocketAddr,json_msg);
}

function seekTo(seconds: number): void {
  player.seekTo(seconds,true);
  // sendToWebSocket(myLocalWebSocketAddr,"SeekTo");
}

console.log("Starting WeTubeClient (JS)")

// Create New Websocket
function sendToWebSocket(addr: string,m: string) {
  var ws = new WebSocket(addr, "protocolOne");
  ws.onopen = function (event) {
    ws.send(m);
    console.log(m);
  };
  ws.onmessage = function (event) {
    var msg = JSON.parse(event.data)
    console.log("Go Client: "+event.data.trim()); // this will turn into a command to be parsed and executed, should also update peer set
  }
  ws.onclose = function (event) {
    console.log("Websocket closing...");
  }
}

// Establish WebSocket Connection with WeTube (Go) Client
var myLocalWebSocketAddr: string;
// var myLocalWebSocket: WebSocket;
var tempWebSocket = new WebSocket("ws://localhost:8080/ws/js", "protocolOne");
tempWebSocket.onopen = function (event) {
  tempWebSocket.send("Which port should I use?");
  console.log("Which port should I use?");
};
tempWebSocket.onmessage = function (event) {
  console.log("WeTubeServer: Use port "+event.data)
  console.log("Connecting to websocket at ws://localhost:"+event.data+"/ws");
  myLocalWebSocketAddr = "ws://localhost:"+event.data+"/ws"
  sendToWebSocket(myLocalWebSocketAddr,"Who are my peers?");
  tempWebSocket.close();
}