/// <reference path="WeTubeClient.ts" />
// PlayerControls
function playVideo() {
    player.playVideo();
    myLocalWebSocket.send("Play");
}
function pauseVideo() {
    player.pauseVideo();
    myLocalWebSocket.send("Pause");
}
function stopVideo() {
    player.stopVideo();
    console.log("Stopping");
    myLocalWebSocket.send("Stop");
}
function seekTo(seconds) {
    player.seekTo(seconds, true);
    myLocalWebSocket.send("SeekTo");
}
