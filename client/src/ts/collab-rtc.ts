import { domLoaded } from "./lib.ts";

let pc: RTCPeerConnection | null = null;
let localAudioStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;
let isSyncingVideo = false;

// Web Audio mixer for combining local microphone, screen audio, and presentation/file audio
let audioContext: AudioContext | null = null;
let mixedAudioDestination: MediaStreamAudioDestinationNode | null = null;
let micAudioSource: MediaStreamAudioSourceNode | null = null;
let screenAudioSource: MediaStreamAudioSourceNode | null = null;
let fileAudioSource: MediaStreamAudioSourceNode | null = null;
let activeAudioSender: RTCRtpSender | null = null;

// Independent active streams for simultaneous sharing
let cameraStream: MediaStream | null = null;
let screenStream: MediaStream | null = null;
let fileStream: MediaStream | null = null;
let currentSharedFileName: string | null = null;
let activeMediaElement: HTMLVideoElement | HTMLAudioElement | null = null;

// Track sender mapping
let activeVideoSender: RTCRtpSender | null = null;

// Remote active streams mapped by track/stream ID
interface RemoteTileInfo {
	element: HTMLElement;
	video: HTMLVideoElement;
	badge: HTMLElement;
	stream: MediaStream;
	track: MediaStreamTrack;
	streamType: "camera" | "screen" | "presentation";
}
const remoteTiles = new Map<string, RemoteTileInfo>();

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
				const audioBtn =
					document.getElementById("collab-btn-audio") ||
					document.getElementById("voice-btn-audio");
				const isDeafened = audioBtn?.classList.contains("text-error") ?? false;
				remoteAudio.muted = isDeafened;
				remoteAudio.srcObject =
					event.streams[0] || new MediaStream([event.track]);
				remoteAudio.play().catch(() => {});
			} else if (event.track.kind === "video") {
				const remoteStream = event.streams[0] || new MediaStream([event.track]);
				const track = event.track;

				const onTrackActive = () => {
					renderRemoteVideoTrack(remoteStream, track);
				};
				const onTrackInactive = () => {
					removeRemoteVideo(track.id);
				};

				track.addEventListener("unmute", onTrackActive);
				track.addEventListener("mute", onTrackInactive);
				track.addEventListener("ended", onTrackInactive);

				if (!track.muted) {
					renderRemoteVideoTrack(remoteStream, track);
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
				audio: {
					echoCancellation: true,
					noiseSuppression: true,
					autoGainControl: true,
				},
				video: false,
			});
			localAudioStream = micStream;
		} catch (err: any) {
			console.warn("Could not obtain audio stream:", err);
		}

		setupAudioMixer();

		const outboundAudioTrack = getMixedOutboundAudioTrack();
		if (outboundAudioTrack && mixedAudioDestination) {
			activeAudioSender = pc.addTrack(
				outboundAudioTrack,
				mixedAudioDestination.stream,
			);
		} else if (
			localAudioStream &&
			localAudioStream.getAudioTracks().length > 0
		) {
			activeAudioSender = pc.addTrack(
				localAudioStream.getAudioTracks()[0],
				localAudioStream,
			);
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

	teardownAudioMixer();

	if (localAudioStream) {
		localAudioStream.getTracks().forEach((t) => t.stop());
		localAudioStream = null;
	}
	if (pc) {
		pc.close();
		pc = null;
		activeVideoSender = null;
		activeAudioSender = null;
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

		// Include screen audio in mixer
		updateScreenAudioMixer();

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
	updateScreenAudioMixer();
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

			updateFileAudioMixer();

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
	updateFileAudioMixer();
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

// Helper to look up active participant moniker from connected channels UI
function getParticipantMoniker(): string {
	const memberEl =
		document.querySelector("#collab-channels-list [popover] span.truncate") ||
		document.querySelector("#collab-channels-list .avatar img[alt]") ||
		document.querySelector("#members .font-medium") ||
		document.querySelector(
			".channel-collab-entry [popover] button span.truncate",
		);

	if (memberEl) {
		const alt = memberEl.getAttribute("alt");
		if (alt && alt.trim() !== "") return alt.trim();
		const text = memberEl.textContent?.trim();
		if (text && text !== "") return text;
	}
	return "Partner";
}

// 4. Remote Video Handling with Dynamic Multi-Tile Support
export function renderRemoteVideoTrack(
	stream: MediaStream,
	track?: MediaStreamTrack,
) {
	const dock = document.getElementById("collab-video-dock");
	const remoteContainer = document.getElementById("collab-remote-tiles");
	if (dock) dock.classList.remove("hidden");

	const activeTrack = track || stream.getVideoTracks()[0];
	if (!activeTrack) return;

	const trackId = activeTrack.id;
	const trackLabel = (activeTrack.label || "").toLowerCase();
	const streamId = (stream.id || "").toLowerCase();

	// Determine stream type: Screen, Presentation, or Camera
	let streamType: "camera" | "screen" | "presentation" = "camera";
	let streamTypeName = "Camera";
	if (
		trackLabel.includes("screen") ||
		streamId.includes("screen") ||
		trackId.includes("screen")
	) {
		streamType = "screen";
		streamTypeName = "Screen";
	} else if (
		trackLabel.includes("presentation") ||
		trackLabel.includes("file") ||
		streamId.includes("presentation") ||
		trackId.includes("presentation")
	) {
		streamType = "presentation";
		streamTypeName = "Presentation";
	}

	const participantName = getParticipantMoniker();
	const badgeLabel = `${participantName}'s ${streamTypeName}`;

	let tileInfo = remoteTiles.get(trackId);

	if (!tileInfo) {
		// Use template video-tile-remote if it's the first and currently hidden
		const templateTile = document.getElementById("video-tile-remote");
		let tileEl: HTMLElement;
		let videoEl: HTMLVideoElement;
		let badgeEl: HTMLElement;

		if (
			templateTile &&
			templateTile.classList.contains("hidden") &&
			remoteTiles.size === 0
		) {
			tileEl = templateTile;
			videoEl = tileEl.querySelector("video") as HTMLVideoElement;
			badgeEl = tileEl.querySelector("#collab-remote-badge") as HTMLElement;
		} else {
			// Clone template for additional remote video stream
			tileEl = (
				templateTile
					? templateTile.cloneNode(true)
					: document.createElement("div")
			) as HTMLElement;
			tileEl.id = `video-tile-remote-${trackId}`;
			videoEl = tileEl.querySelector("video") as HTMLVideoElement;
			badgeEl = tileEl.querySelector("span") as HTMLElement;

			// Update buttons IDs/attributes
			const fsBtn = tileEl.querySelector(
				"button[title*='fullscreen']",
			) as HTMLButtonElement | null;
			const closeBtn = tileEl.querySelector(
				"button[title*='video']",
			) as HTMLButtonElement | null;

			if (fsBtn) {
				fsBtn.id = `collab-btn-fullscreen-remote-${trackId}`;
				fsBtn.addEventListener("click", () => {
					toggleTileFullscreen(tileEl.id);
				});
			}
			if (closeBtn) {
				closeBtn.id = `collab-btn-close-remote-${trackId}`;
				closeBtn.addEventListener("click", () => {
					removeRemoteVideo(trackId);
				});
			}

			if (remoteContainer) {
				remoteContainer.appendChild(tileEl);
			}
		}

		tileEl.classList.remove("hidden");
		tileEl.setAttribute("data-track-id", trackId);
		tileEl.setAttribute("data-stream-type", streamType);

		tileInfo = {
			element: tileEl,
			video: videoEl,
			badge: badgeEl,
			stream,
			track: activeTrack,
			streamType,
		};
		remoteTiles.set(trackId, tileInfo);
	}

	if (tileInfo.badge) {
		tileInfo.badge.textContent = badgeLabel;
		tileInfo.badge.setAttribute("title", badgeLabel);
	}

	if (tileInfo.video) {
		if (tileInfo.video.srcObject !== stream) {
			tileInfo.video.srcObject = stream;
		}
		tileInfo.video.muted = false;
		tileInfo.video.play().catch(() => {});
	}
}

export function removeRemoteVideo(trackId?: string) {
	if (trackId && remoteTiles.has(trackId)) {
		const tile = remoteTiles.get(trackId);
		if (tile) {
			if (tile.element.id === "video-tile-remote") {
				tile.element.classList.add("hidden");
				if (tile.video) tile.video.srcObject = null;
			} else {
				tile.element.remove();
			}
			remoteTiles.delete(trackId);
		}
	} else if (!trackId) {
		for (const [id, tile] of remoteTiles.entries()) {
			if (tile.element.id === "video-tile-remote") {
				tile.element.classList.add("hidden");
				if (tile.video) tile.video.srcObject = null;
			} else {
				tile.element.remove();
			}
		}
		remoteTiles.clear();
	}
	checkDockVisibility();
}

function checkDockVisibility() {
	const dock = document.getElementById("collab-video-dock");
	const localTile = document.getElementById("video-tile-local");
	const screenTile = document.getElementById("video-tile-screen");
	const fileTile = document.getElementById("video-tile-file");

	let hasVisibleRemote = false;
	for (const tile of remoteTiles.values()) {
		if (!tile.element.classList.contains("hidden")) {
			hasVisibleRemote = true;
			break;
		}
	}

	const hasVisibleTile =
		(localTile && !localTile.classList.contains("hidden")) ||
		(screenTile && !screenTile.classList.contains("hidden")) ||
		(fileTile && !fileTile.classList.contains("hidden")) ||
		hasVisibleRemote;

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
		if (!isDeafened && remoteAudio.paused) {
			remoteAudio.play().catch(() => {});
		}
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

// Audio mixer helpers
function setupAudioMixer() {
	try {
		const AudioContextClass =
			window.AudioContext || (window as any).webkitAudioContext;
		if (!AudioContextClass) return;

		if (!audioContext || audioContext.state === "closed") {
			audioContext = new AudioContextClass();
		}
		if (audioContext.state === "suspended") {
			audioContext.resume().catch(() => {});
		}

		if (!mixedAudioDestination) {
			mixedAudioDestination = audioContext.createMediaStreamDestination();
		}

		// Connect mic stream
		if (localAudioStream && localAudioStream.getAudioTracks().length > 0) {
			if (micAudioSource) {
				micAudioSource.disconnect();
				micAudioSource = null;
			}
			micAudioSource = audioContext.createMediaStreamSource(localAudioStream);
			micAudioSource.connect(mixedAudioDestination);
		}

		updateScreenAudioMixer();
		updateFileAudioMixer();
	} catch (e) {
		console.warn("Failed to set up audio mixer:", e);
	}
}

function syncOutboundAudioSender() {
	if (!pc) return;
	const track = getMixedOutboundAudioTrack();
	if (!track) return;

	if (activeAudioSender) {
		activeAudioSender.replaceTrack(track).catch(() => {});
	} else {
		const senders = pc.getSenders();
		const audioSender = senders.find(
			(s) =>
				(s.track && s.track.kind === "audio") || (s as any).kind === "audio",
		);
		if (audioSender) {
			activeAudioSender = audioSender;
			audioSender.replaceTrack(track).catch(() => {});
		} else if (mixedAudioDestination) {
			activeAudioSender = pc.addTrack(track, mixedAudioDestination.stream);
		}
	}
}

function updateScreenAudioMixer() {
	try {
		if (screenAudioSource) {
			screenAudioSource.disconnect();
			screenAudioSource = null;
		}
		if (
			audioContext &&
			mixedAudioDestination &&
			screenStream &&
			screenStream.getAudioTracks().length > 0
		) {
			screenAudioSource = audioContext.createMediaStreamSource(screenStream);
			screenAudioSource.connect(mixedAudioDestination);
		}
		syncOutboundAudioSender();
	} catch (e) {
		console.warn("Error updating screen audio mixer:", e);
	}
}

function updateFileAudioMixer() {
	try {
		if (fileAudioSource) {
			fileAudioSource.disconnect();
			fileAudioSource = null;
		}
		if (
			audioContext &&
			mixedAudioDestination &&
			fileStream &&
			fileStream.getAudioTracks().length > 0
		) {
			fileAudioSource = audioContext.createMediaStreamSource(fileStream);
			fileAudioSource.connect(mixedAudioDestination);
		}
		syncOutboundAudioSender();
	} catch (e) {
		console.warn("Error updating file audio mixer:", e);
	}
}

function getMixedOutboundAudioTrack(): MediaStreamTrack | null {
	if (
		mixedAudioDestination &&
		mixedAudioDestination.stream.getAudioTracks().length > 0
	) {
		return mixedAudioDestination.stream.getAudioTracks()[0];
	}
	if (localAudioStream && localAudioStream.getAudioTracks().length > 0) {
		return localAudioStream.getAudioTracks()[0];
	}
	return null;
}

function teardownAudioMixer() {
	if (micAudioSource) {
		try {
			micAudioSource.disconnect();
		} catch (_) {}
		micAudioSource = null;
	}
	if (screenAudioSource) {
		try {
			screenAudioSource.disconnect();
		} catch (_) {}
		screenAudioSource = null;
	}
	if (fileAudioSource) {
		try {
			fileAudioSource.disconnect();
		} catch (_) {}
		fileAudioSource = null;
	}
	if (mixedAudioDestination) {
		mixedAudioDestination = null;
	}
	if (audioContext && audioContext.state !== "closed") {
		try {
			audioContext.close();
		} catch (_) {}
		audioContext = null;
	}
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
