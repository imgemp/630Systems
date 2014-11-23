// PlayerControls

function playVideo(): void {
  player.playVideo();
}

function pauseVideo(): void {
  player.pauseVideo();
}

function stopVideo(): void {
  player.stopVideo();
}

function seekTo(seconds: number): void {
	player.seekTo(seconds,true);
}