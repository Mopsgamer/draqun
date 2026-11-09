import { domLoaded } from "./lib.ts";

let pc: RTCPeerConnection | null = null;
let localAudioStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;
let isSyncingVideo = false;

// Independent active streams for simultaneous sharing
let cameraStream: MediaStream | null = null;
let screenStream: MediaStream | null = null;
let fileStream: MediaStream | null = null;
let currentSharedFileName: string | null = null;
let activeMediaElement: HTMLVideoElement | HTMLAudioElement | null = null;

// Track sender mapping
let activeVideoSender: RTCRtpSender | null = null;

const rtcConfig: RTCConfiguration = {
	iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
};

export async function startWebRTCSession(
	groupId: string,
	channelId: string,
	subchannelName: string,
) {
	if (isStarting) return;
	isStarting = true;
	try {
		stopWebRTCSession();
		currentGroupId = groupId;
		currentChannelId = channelId;
		currentSubchannelName = subchannelName;

		pc = new RTCPeerConnection(rtcConfig);

		remoteAudio = new Audio();
		remoteAudio.autoplay = true;

		pc.ontrack = (event) => {
			if (event.track.kind === "audio" && remoteAudio) {
				remoteAudio.srcObject =
					event.streams[0] || new MediaStream([event.track]);
			} else if (event.track.kind === "video") {
				const remoteStream = event.streams[0] || new MediaStream([event.track]);
				const track = event.track;

				const onTrackActive = () => {
					renderRemoteVideoTrack(remoteStream);
				};
				const onTrackInactive = () => {
					removeRemoteVideo();
				};

				track.addEventListener("unmute", onTrackActive);
				track.addEventListener("mute", onTrackInactive);
				track.addEventListener("ended", onTrackInactive);

				if (!track.muted) {
					renderRemoteVideoTrack(remoteStream);
				}
			}
		};

		// Check camera toggle in DOM
		const videoBtn =
			document.getElementById("collab-btn-video") ||
			document.getElementById("voice-btn-video");
		const wantVideo = videoBtn?.classList.contains("text-primary") ?? false;

		localAudioStream = null;
		try {
			const micStream = await navigator.mediaDevices.getUserMedia({
				audio: true,
				video: false,
			});
			localAudioStream = micStream;
		} catch (err: any) {
			console.warn("Could not obtain audio stream:", err);
		}

		if (localAudioStream) {
			localAudioStream.getAudioTracks().forEach((track) => {
				if (pc && localAudioStream) {
					pc.addTrack(track, localAudioStream);
				}
			});
		} else {
			if (pc.addTransceiver) {
				pc.addTransceiver("audio", { direction: "recvonly" });
			}
		}

		if (pc.addTransceiver) {
			pc.addTransceiver("video", { direction: "sendrecv" });
		}

		if (wantVideo) {
			await startCameraStream();
		}

		await negotiateWebRTC(groupId, channelId, subchannelName);
	} catch (e) {
		console.warn("WebRTC startup notice:", e);
	} finally {
		isStarting = false;
	}
}

async function negotiateWebRTC(
	groupId: string,
	channelId: string,
	subchannelName: string,
) {
	if (!pc) return;
	const offer = await pc.createOffer();
	await pc.setLocalDescription(offer);

	// Wait for ICE gathering
	await new Promise<void>((resolve) => {
		if (!pc || pc.iceGatheringState === "complete") {
			resolve();
		} else {
			const checkState = () => {
				if (pc?.iceGatheringState === "complete") {
					pc.removeEventListener("icegatheringstatechange", checkState);
					resolve();
				}
			};
			pc.addEventListener("icegatheringstatechange", checkState);
			setTimeout(resolve, 1500); // safety fallback
		}
	});

	const res = await fetch(`/groups/${groupId}/collab/signal`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({
			channel_id: parseInt(channelId, 10),
			subchannel_name: subchannelName || "1",
			sdp: pc.localDescription?.sdp,
			type: pc.localDescription?.type,
		}),
	});

	if (res.ok) {
		const answer = await res.json();
		await pc.setRemoteDescription(new RTCSessionDescription(answer));
	}
}

export function stopWebRTCSession() {
	stopCameraStream();
	stopScreenShare();
	stopFileShare();

	if (localAudioStream) {
		localAudioStream.getTracks().forEach((t) => t.stop());
		localAudioStream = null;
	}
	if (pc) {
		pc.close();
		pc = null;
		activeVideoSender = null;
	}
	if (remoteAudio) {
		remoteAudio.srcObject = null;
		remoteAudio = null;
	}
	currentGroupId = null;
	currentChannelId = null;
	currentSubchannelName = null;
	removeRemoteVideo();
}

async function updatePeerVideoTrack() {
	if (!pc) return;

	// Determine primary outbound video track (priority: Screen > File > Camera)
	let outboundTrack: MediaStreamTrack | null = null;
	let outboundStream: MediaStream | null = null;

	if (screenStream && screenStream.getVideoTracks().length > 0) {
		outboundTrack = screenStream.getVideoTracks()[0];
		outboundStream = screenStream;
	} else if (fileStream && fileStream.getVideoTracks().length > 0) {
		outboundTrack = fileStream.getVideoTracks()[0];
		outboundStream = fileStream;
	} else if (cameraStream && cameraStream.getVideoTracks().length > 0) {
		outboundTrack = cameraStream.getVideoTracks()[0];
		outboundStream = cameraStream;
	}

	const senders = pc.getSenders();
	const videoSender =
		activeVideoSender ||
		senders.find((s) => s.track && s.track.kind === "video") ||
		senders.find((s) => !s.track && (s as any).kind === "video");

	if (outboundTrack) {
		if (videoSender) {
			await videoSender.replaceTrack(outboundTrack);
			activeVideoSender = videoSender;
		} else {
			if (outboundStream) {
				activeVideoSender = pc.addTrack(outboundTrack, outboundStream);
			}
			if (currentGroupId && currentChannelId && currentSubchannelName) {
				await negotiateWebRTC(
					currentGroupId,
					currentChannelId,
					currentSubchannelName,
				);
			}
		}
	} else {
		if (videoSender) {
			await videoSender.replaceTrack(null);
		}
	}
}

// 1. Camera Handling
export async function startCameraStream() {
	try {
		if (cameraStream) {
			renderLocalCamera(cameraStream);
			return;
		}
		const vStream = await navigator.mediaDevices.getUserMedia({ video: true });
		cameraStream = vStream;
		renderLocalCamera(vStream);
		await updatePeerVideoTrack();
	} catch (e: any) {
		console.warn("Could not capture camera stream:", e);
	}
}

export async function stopCameraStream() {
	if (cameraStream) {
		cameraStream.getTracks().forEach((t) => t.stop());
		cameraStream = null;
	}
	removeLocalCamera();
	await updatePeerVideoTrack();
}

export function renderLocalCamera(stream: MediaStream) {
	const dock = document.getElementById("collab-video-dock");
	const localTile = document.getElementById("video-tile-local");
	const localVideo = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (dock) dock.classList.remove("hidden");
	if (localTile) localTile.classList.remove("hidden");
	if (localVideo) {
		if (localVideo.srcObject !== stream) {
			localVideo.srcObject = stream;
		}
		localVideo.muted = true;
		localVideo.play().catch(() => {});
	}
}

function removeLocalCamera() {
	const localTile = document.getElementById("video-tile-local");
	const localVideo = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (localTile) localTile.classList.add("hidden");
	if (localVideo) localVideo.srcObject = null;
	checkDockVisibility();
}

// 2. Screen Share Handling
export async function toggleScreenShare() {
	if (screenStream) {
		await stopScreenShare();
	} else {
		await startScreenShare();
	}
}

export async function startScreenShare() {
	try {
		const displayStream = await navigator.mediaDevices.getDisplayMedia({
			video: true,
			audio: true,
		});

		const vTrack = displayStream.getVideoTracks()[0];
		if (!vTrack) return;

		screenStream = displayStream;
		updateScreenShareIcon(true);
		renderScreenShare(displayStream);

		vTrack.addEventListener("ended", () => {
			stopScreenShare();
		});

		// Include screen audio if present
		const aTrack = displayStream.getAudioTracks()[0];
		if (aTrack && pc) {
			pc.addTrack(aTrack, displayStream);
		}

		await updatePeerVideoTrack();
	} catch (e: any) {
		console.warn("Screen share canceled or failed:", e);
		updateScreenShareIcon(false);
	}
}

export async function stopScreenShare() {
	if (screenStream) {
		screenStream.getTracks().forEach((t) => t.stop());
		screenStream = null;
	}
	updateScreenShareIcon(false);
	removeScreenShare();
	await updatePeerVideoTrack();
}

export function renderScreenShare(stream: MediaStream) {
	const dock = document.getElementById("collab-video-dock");
	const screenTile = document.getElementById("video-tile-screen");
	const screenVideo = document.getElementById(
		"collab-screen-video",
	) as HTMLVideoElement | null;
	if (dock) dock.classList.remove("hidden");
	if (screenTile) screenTile.classList.remove("hidden");
	if (screenVideo) {
		if (screenVideo.srcObject !== stream) {
			screenVideo.srcObject = stream;
		}
		screenVideo.muted = true;
		screenVideo.play().catch(() => {});
	}
}

function removeScreenShare() {
	const screenTile = document.getElementById("video-tile-screen");
	const screenVideo = document.getElementById(
		"collab-screen-video",
	) as HTMLVideoElement | null;
	if (screenTile) screenTile.classList.add("hidden");
	if (screenVideo) screenVideo.srcObject = null;
	checkDockVisibility();
}

function updateScreenShareIcon(sharing: boolean) {
	const icon = document.getElementById("collab-screenshare-icon");
	const btn = document.getElementById("collab-btn-screenshare");
	if (!icon || !btn) return;
	if (sharing) {
		icon.classList.remove("icon-[lucide--screen-share]");
		icon.classList.add("icon-[lucide--screen-share-off]");
		btn.classList.add("text-primary");
		btn.setAttribute("title", "Stop sharing screen");
	} else {
		icon.classList.remove("icon-[lucide--screen-share-off]");
		icon.classList.add("icon-[lucide--screen-share]");
		btn.classList.remove("text-primary");
		btn.setAttribute("title", "Share screen");
	}
}

// 3. File / Presentation Handling
export async function shareFile(file: File) {
	stopFileShare();

	currentSharedFileName = file.name;
	const fileType = file.type;
	const isAudio = fileType.startsWith("audio/");
	const isVideo = fileType.startsWith("video/");
	const isPresentationOrDoc =
		fileType.includes("presentation") ||
		fileType.includes("powerpoint") ||
		fileType.includes("pdf") ||
		fileType.startsWith("image/") ||
		file.name.endsWith(".pptx") ||
		file.name.endsWith(".ppt");

	const fileUrl = URL.createObjectURL(file);

	if (isVideo || isAudio) {
		const mediaEl = isVideo
			? document.createElement("video")
			: (document.createElement("audio") as unknown as HTMLVideoElement);
		mediaEl.src = fileUrl;
		mediaEl.crossOrigin = "anonymous";
		mediaEl.loop = true;
		mediaEl.muted = false;
		activeMediaElement = mediaEl;

		await mediaEl.play().catch(() => {});

		const captureFn =
			(mediaEl as any).captureStream || (mediaEl as any).mozCaptureStream;
		if (captureFn) {
			const stream: MediaStream = captureFn.call(mediaEl);
			fileStream = stream;

			const aTrack = stream.getAudioTracks()[0];
			if (aTrack && pc) {
				pc.addTrack(aTrack, stream);
			}

			renderFileShare(stream, file.name);
			await updatePeerVideoTrack();
		}
	} else if (isPresentationOrDoc) {
		const canvas = document.createElement("canvas");
		canvas.width = 1280;
		canvas.height = 720;
		const ctx = canvas.getContext("2d");

		let animActive = true;
		if (fileType.startsWith("image/")) {
			const img = new Image();
			img.src = fileUrl;
			img.onload = async () => {
				const drawImageFrame = () => {
					if (!animActive) return;
					if (ctx) {
						ctx.fillStyle = "#1e1e24";
						ctx.fillRect(0, 0, canvas.width, canvas.height);
						const scale = Math.min(
							canvas.width / img.width,
							canvas.height / img.height,
						);
						const x = canvas.width / 2 - (img.width / 2) * scale;
						const y = canvas.height / 2 - (img.height / 2) * scale;
						ctx.drawImage(img, x, y, img.width * scale, img.height * scale);
					}
					requestAnimationFrame(drawImageFrame);
				};
				requestAnimationFrame(drawImageFrame);
			};
		} else {
			if (ctx) {
				ctx.fillStyle = "#1e1e24";
				ctx.fillRect(0, 0, canvas.width, canvas.height);
				ctx.fillStyle = "#e2e8f0";
				ctx.font = "bold 32px sans-serif";
				ctx.textAlign = "center";
				ctx.fillText(`Presentation: ${file.name}`, canvas.width / 2, 320);
				ctx.font = "20px sans-serif";
				ctx.fillStyle = "#94a3b8";
				ctx.fillText(
					`Sharing file presentation (${Math.round(file.size / 1024)} KB)`,
					canvas.width / 2,
					380,
				);
			}
		}

		const renderLoop = () => {
			if (!animActive) return;
			requestAnimationFrame(renderLoop);
		};
		requestAnimationFrame(renderLoop);

		const stream = canvas.captureStream(30);
		fileStream = stream;

		const vTrack = stream.getVideoTracks()[0];
		if (vTrack) {
			vTrack.addEventListener("ended", () => {
				animActive = false;
			});
		}

		renderFileShare(stream, file.name);
		await updatePeerVideoTrack();
	}
}

export async function stopFileShare() {
	if (fileStream) {
		fileStream.getTracks().forEach((t) => t.stop());
		fileStream = null;
	}
	if (activeMediaElement) {
		activeMediaElement.pause();
		activeMediaElement.src = "";
		activeMediaElement.remove();
		activeMediaElement = null;
	}
	currentSharedFileName = null;
	removeFileShare();
	await updatePeerVideoTrack();
}

export function renderFileShare(stream: MediaStream, fileName?: string) {
	const dock = document.getElementById("collab-video-dock");
	const fileTile = document.getElementById("video-tile-file");
	const fileVideo = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	const fileBadge = document.getElementById("collab-file-badge");

	if (fileBadge) {
		const name = fileName || currentSharedFileName;
		fileBadge.textContent = name
			? `Your presentation (${name})`
			: "Your presentation";
	}
	if (dock) dock.classList.remove("hidden");
	if (fileTile) fileTile.classList.remove("hidden");
	if (fileVideo) {
		if (fileVideo.srcObject !== stream) {
			fileVideo.srcObject = stream;
		}
		fileVideo.muted = true;
		fileVideo.play().catch(() => {});
	}
}

function removeFileShare() {
	const fileTile = document.getElementById("video-tile-file");
	const fileVideo = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	if (fileTile) fileTile.classList.add("hidden");
	if (fileVideo) fileVideo.srcObject = null;
	checkDockVisibility();
}

// 4. Remote Video Handling
export function renderRemoteVideoTrack(stream: MediaStream) {
	const dock = document.getElementById("collab-video-dock");
	const remoteTile = document.getElementById("video-tile-remote");
	const remoteBadge = document.getElementById("collab-remote-badge");
	const remoteVideo = document.getElementById(
		"collab-remote-video",
	) as HTMLVideoElement | null;
	if (dock) dock.classList.remove("hidden");
	if (remoteTile) remoteTile.classList.remove("hidden");
	if (remoteBadge) {
		remoteBadge.textContent = "Participant";
	}
	if (remoteVideo) {
		if (remoteVideo.srcObject !== stream) {
			remoteVideo.srcObject = stream;
		}
		remoteVideo.muted = false;
		remoteVideo.play().catch(() => {});
	}
}

function removeRemoteVideo() {
	const remoteTile = document.getElementById("video-tile-remote");
	const remoteVideo = document.getElementById(
		"collab-remote-video",
	) as HTMLVideoElement | null;
	if (remoteTile) remoteTile.classList.add("hidden");
	if (remoteVideo) remoteVideo.srcObject = null;
	checkDockVisibility();
}

function checkDockVisibility() {
	const dock = document.getElementById("collab-video-dock");
	const localTile = document.getElementById("video-tile-local");
	const screenTile = document.getElementById("video-tile-screen");
	const fileTile = document.getElementById("video-tile-file");
	const remoteTile = document.getElementById("video-tile-remote");

	const hasVisibleTile =
		(localTile && !localTile.classList.contains("hidden")) ||
		(screenTile && !screenTile.classList.contains("hidden")) ||
		(fileTile && !fileTile.classList.contains("hidden")) ||
		(remoteTile && !remoteTile.classList.contains("hidden"));

	if (dock && !hasVisibleTile) {
		dock.classList.add("hidden");
	}
}

export function toggleTileFullscreen(tileId: string) {
	const tile = document.getElementById(tileId);
	if (!tile) return;

	if (document.fullscreenElement) {
		document.exitFullscreen?.().catch(() => {});
	} else {
		tile.requestFullscreen?.().catch(() => {});
	}
}

function updateFullscreenIcons() {
	const fsEl = document.fullscreenElement;
	const tiles = [
		{ tileId: "video-tile-local", btnId: "collab-btn-fullscreen-local" },
		{ tileId: "video-tile-screen", btnId: "collab-btn-fullscreen-screen" },
		{ tileId: "video-tile-file", btnId: "collab-btn-fullscreen-file" },
		{ tileId: "video-tile-remote", btnId: "collab-btn-fullscreen-remote" },
	];

	for (const { tileId, btnId } of tiles) {
		const tile = document.getElementById(tileId);
		const btn = document.getElementById(btnId);
		if (!btn || !tile) continue;

		const icon = btn.querySelector(".icon");
		const isTileFs = fsEl !== null && (fsEl === tile || tile.contains(fsEl));
		if (icon) {
			if (isTileFs || (fsEl !== null && fsEl === tile)) {
				icon.classList.remove("icon-[lucide--maximize-2]");
				icon.classList.add("icon-[lucide--minimize-2]");
				btn.setAttribute("title", "Exit fullscreen");
			} else {
				icon.classList.remove("icon-[lucide--minimize-2]");
				icon.classList.add("icon-[lucide--maximize-2]");
				btn.setAttribute("title", "Toggle fullscreen");
			}
		}
	}
}

export function toggleDockUI() {
	const tilesContainer = document.getElementById("collab-dock-tiles-container");
	const toggleIcon = document.getElementById("collab-dock-toggle-icon");
	if (!tilesContainer || !toggleIcon) return;

	const isHidden = tilesContainer.classList.contains("hidden");
	if (isHidden) {
		tilesContainer.classList.remove("hidden");
		toggleIcon.classList.remove("icon-[lucide--eye]");
		toggleIcon.classList.add("icon-[lucide--eye-off]");
	} else {
		tilesContainer.classList.add("hidden");
		toggleIcon.classList.remove("icon-[lucide--eye-off]");
		toggleIcon.classList.add("icon-[lucide--eye]");
	}
}

async function syncVideoTrack(wantVideo: boolean) {
	if (!pc || isStarting || isSyncingVideo) return;
	isSyncingVideo = true;
	try {
		if (wantVideo) {
			await startCameraStream();
		} else {
			await stopCameraStream();
		}
	} finally {
		isSyncingVideo = false;
	}
}

export function checkAndSyncVoice() {
	const container =
		document.getElementById("collab-header-container") ||
		document.getElementById("voice-header-container");
	if (
		!container ||
		container.classList.contains("hidden") ||
		!container.hasAttribute("data-voice-active")
	) {
		stopWebRTCSession();
		return;
	}

	const groupId = container.getAttribute("data-group-id");
	const firstCh =
		container.querySelector(".channel-collab-entry") ||
		container.querySelector(".channel-voice-entry");
	const channelId = firstCh?.getAttribute("data-channel-id");
	const subchannelName = firstCh?.getAttribute("data-subchannel-name") || "1";

	if (groupId && channelId) {
		if (
			!pc ||
			currentChannelId !== channelId ||
			currentSubchannelName !== subchannelName
		) {
			startWebRTCSession(groupId, channelId, subchannelName);
			return;
		}
	}

	// Sync local mic mute
	const micBtn =
		document.getElementById("collab-btn-mic") ||
		document.getElementById("voice-btn-mic");
	const isMicMuted = micBtn?.querySelector(".icon-[lucide--mic-off]") !== null;
	if (localAudioStream) {
		localAudioStream.getAudioTracks().forEach((track) => {
			track.enabled = !isMicMuted;
		});
	}

	// Sync local deafen/headphones
	const audioBtn =
		document.getElementById("collab-btn-audio") ||
		document.getElementById("voice-btn-audio");
	const isDeafened = audioBtn?.classList.contains("text-error") ?? false;
	if (remoteAudio) {
		remoteAudio.muted = isDeafened;
	}

	// Sync video toggle
	const videoBtn =
		document.getElementById("collab-btn-video") ||
		document.getElementById("voice-btn-video");
	const wantVideo = videoBtn?.classList.contains("text-primary") ?? false;
	if (wantVideo && cameraStream && cameraStream.getVideoTracks().length > 0) {
		renderLocalCamera(cameraStream);
	}
	syncVideoTrack(wantVideo);
}

function bindCollabEventHandlers() {
	// Fullscreen buttons
	const fsBindings = [
		{ id: "collab-btn-fullscreen-local", tile: "video-tile-local" },
		{ id: "collab-btn-fullscreen-screen", tile: "video-tile-screen" },
		{ id: "collab-btn-fullscreen-file", tile: "video-tile-file" },
		{ id: "collab-btn-fullscreen-remote", tile: "video-tile-remote" },
	];

	for (const { id, tile } of fsBindings) {
		const btn = document.getElementById(id);
		if (btn && !btn.hasAttribute("data-bound")) {
			btn.setAttribute("data-bound", "true");
			btn.addEventListener("click", () => {
				toggleTileFullscreen(tile);
			});
		}
	}

	// Dock hide/unhide toggle button
	const dockToggle = document.getElementById("collab-dock-toggle-btn");
	if (dockToggle && !dockToggle.hasAttribute("data-bound")) {
		dockToggle.setAttribute("data-bound", "true");
		dockToggle.addEventListener("click", () => {
			toggleDockUI();
		});
	}

	// Close buttons
	const closeLocal = document.getElementById("collab-btn-close-local");
	if (closeLocal && !closeLocal.hasAttribute("data-bound")) {
		closeLocal.setAttribute("data-bound", "true");
		closeLocal.addEventListener("click", () => {
			if (document.fullscreenElement) {
				document.exitFullscreen?.().catch(() => {});
			}
			const videoBtn = document.getElementById("collab-btn-video");
			if (videoBtn && videoBtn.classList.contains("text-primary")) {
				videoBtn.click();
			} else {
				stopCameraStream();
			}
		});
	}

	const closeScreen = document.getElementById("collab-btn-close-screen");
	if (closeScreen && !closeScreen.hasAttribute("data-bound")) {
		closeScreen.setAttribute("data-bound", "true");
		closeScreen.addEventListener("click", () => {
			if (document.fullscreenElement) {
				document.exitFullscreen?.().catch(() => {});
			}
			stopScreenShare();
		});
	}

	const closeFile = document.getElementById("collab-btn-close-file");
	if (closeFile && !closeFile.hasAttribute("data-bound")) {
		closeFile.setAttribute("data-bound", "true");
		closeFile.addEventListener("click", () => {
			if (document.fullscreenElement) {
				document.exitFullscreen?.().catch(() => {});
			}
			stopFileShare();
		});
	}

	const closeRemote = document.getElementById("collab-btn-close-remote");
	if (closeRemote && !closeRemote.hasAttribute("data-bound")) {
		closeRemote.setAttribute("data-bound", "true");
		closeRemote.addEventListener("click", () => {
			if (document.fullscreenElement) {
				document.exitFullscreen?.().catch(() => {});
			}
			removeRemoteVideo();
		});
	}

	// Screen share button
	const screenShareBtn = document.getElementById("collab-btn-screenshare");
	if (screenShareBtn && !screenShareBtn.hasAttribute("data-bound")) {
		screenShareBtn.setAttribute("data-bound", "true");
		screenShareBtn.addEventListener("click", () => {
			toggleScreenShare();
		});
	}

	// File presentation input
	const fileInput = document.getElementById(
		"collab-file-input",
	) as HTMLInputElement | null;
	if (fileInput && !fileInput.hasAttribute("data-bound")) {
		fileInput.setAttribute("data-bound", "true");
		fileInput.addEventListener("change", () => {
			if (fileInput.files && fileInput.files[0]) {
				shareFile(fileInput.files[0]);
			}
		});
	}
}

domLoaded.then(() => {
	checkAndSyncVoice();
	bindCollabEventHandlers();

	document.addEventListener("fullscreenchange", updateFullscreenIcons);

	// HTMX lifecycle handlers: afterSwap and afterSettle fire after DOM elements are replaced in the page
	const onHtmxUpdate = (evt: Event) => {
		const customEvt = evt as CustomEvent;
		const target = customEvt.detail?.target as HTMLElement | null;
		if (
			!target ||
			target.id === "collab-header-container" ||
			target.closest("#collab-header-container") ||
			target.id === "collab-dropdown-menu" ||
			target.closest("#collab-dropdown-menu") ||
			target.id === "collab-dropdown-menu-content" ||
			target.closest("#collab-dropdown-menu-content") ||
			target.id === "collab-quick-controls" ||
			target.closest("#collab-quick-controls") ||
			target.id === "channels" ||
			target.closest("#channels") ||
			target.id === "channels-view" ||
			target.id === "app-content"
		) {
			checkAndSyncVoice();
			bindCollabEventHandlers();
		}
	};

	document.body.addEventListener("htmx:afterSwap", onHtmxUpdate);
	document.body.addEventListener("htmx:afterSettle", onHtmxUpdate);
	document.body.addEventListener("htmx:oobAfterSwap", onHtmxUpdate);
});
