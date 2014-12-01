/// <reference path="WeTubeClient.ts" />

// PlayerControls

function playVideo(): void {
	player.playVideo();
	myLocalWebSocket.send("Play");
}

function pauseVideo(): void {
	player.pauseVideo();
	myLocalWebSocket.send("Pause");
}

function stopVideo(): void {
	player.stopVideo();
	console.log("Stopping")
	myLocalWebSocket.send("Stop");
}

function seekTo(seconds: number): void {
	player.seekTo(seconds,true);
	myLocalWebSocket.send("SeekTo");
}