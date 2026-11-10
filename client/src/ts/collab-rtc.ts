import { domLoaded } from "./lib.ts";

// Low-level WebRTC, Audio, and Video Stream state only
let pc: RTCPeerConnection | null = null;
let localAudioStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;
let statsMonitorTimer: any = null;
let lastReportedQuality: string | null = null;

// Web Audio mixer
let audioContext: AudioContext | null = null;
let mixedAudioDestination: MediaStreamAudioDestinationNode | null = null;
let micAudioSource: MediaStreamAudioSourceNode | null = null;
let screenAudioSource: MediaStreamAudioSourceNode | null = null;
let fileAudioSource: MediaStreamAudioSourceNode | null = null;
let activeAudioSender: RTCRtpSender | null = null;

// Media Streams
let cameraStream: MediaStream | null = null;
let screenStream: MediaStream | null = null;
let fileStream: MediaStream | null = null;
let activeMediaElement: HTMLVideoElement | HTMLAudioElement | null = null;
let activeVideoSender: RTCRtpSender | null = null;
let remoteVideoStream: MediaStream | null = null;

const rtcConfig: RTCConfiguration = {
	iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
};

function getGroupId(): string | null {
	const container = document.getElementById("collab-header-container");
	return container?.getAttribute("data-group-id") || currentGroupId;
}

// Global Collab Controller exposed to window for HTML onclick/onchange handlers
export const collabClient = {
	async triggerCameraToggle() {
		const groupId = getGroupId();
		if (!groupId) return;
		const htmx = (globalThis as any).htmx;
		if (cameraStream) {
			await stopCameraStream();
			if (htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-video?enabled=false`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		} else {
			const success = await startCameraStream();
			if (success && htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-video?enabled=true`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		}
	},

	async triggerScreenShare() {
		const groupId = getGroupId();
		if (!groupId) return;
		const htmx = (globalThis as any).htmx;
		if (screenStream) {
			await stopScreenShare();
			if (htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-screenshare?enabled=false`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		} else {
			const success = await startScreenShare();
			if (success && htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-screenshare?enabled=true`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		}
	},

	async handleFileInput(input: HTMLInputElement) {
		const file = input.files?.[0];
		if (!file) return;
		const groupId = getGroupId();
		const success = await shareFile(file);
		if (success && groupId) {
			const htmx = (globalThis as any).htmx;
			if (htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-presentation?enabled=true&filename=${encodeURIComponent(file.name)}`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		}
	},

	toggleTileFullscreen(tileId: string) {
		const tile = document.getElementById(tileId);
		if (!tile) return;
		if (document.fullscreenElement) {
			document.exitFullscreen?.().catch(() => {});
		} else {
			tile.requestFullscreen?.().catch(() => {});
		}
	},

	toggleDockUI() {
		const tilesContainer = document.getElementById(
			"collab-dock-tiles-container",
		);
		if (!tilesContainer) return;
		tilesContainer.classList.toggle("hidden");
	},
};

(globalThis as any).collabClient = collabClient;

// WebRTC Session
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
			if (event.track.kind === "audio") {
				const audioStream = event.streams[0] || new MediaStream([event.track]);
				const audioBtn = document.getElementById("collab-btn-audio");
				const isDeafened = audioBtn?.classList.contains("text-error") ?? false;
				if (remoteAudio) {
					remoteAudio.muted = isDeafened;
					remoteAudio.srcObject = audioStream;
					remoteAudio.play().catch(() => {});
				}
			} else if (event.track.kind === "video") {
				const stream = event.streams[0] || new MediaStream([event.track]);
				remoteVideoStream = stream;
				attachRemoteVideo(stream);
			}
		};

		localAudioStream = null;
		try {
			localAudioStream = await navigator.mediaDevices.getUserMedia({
				audio: {
					echoCancellation: true,
					noiseSuppression: true,
					autoGainControl: true,
				},
				video: false,
			});
		} catch (err) {
			console.warn("Could not capture local mic audio:", err);
		}

		setupAudioMixer();

		const outboundAudioTrack = getMixedOutboundAudioTrack();
		if (outboundAudioTrack && mixedAudioDestination) {
			const existingSenders = pc.getSenders();
			const audioSender = existingSenders.find(
				(s) => s.track && s.track.kind === "audio",
			);
			if (audioSender) {
				await audioSender.replaceTrack(outboundAudioTrack);
				activeAudioSender = audioSender;
			} else {
				activeAudioSender = pc.addTrack(
					outboundAudioTrack,
					mixedAudioDestination.stream,
				);
			}
		} else if (
			localAudioStream &&
			localAudioStream.getAudioTracks().length > 0
		) {
			const existingSenders = pc.getSenders();
			const audioSender = existingSenders.find(
				(s) => s.track && s.track.kind === "audio",
			);
			if (audioSender) {
				await audioSender.replaceTrack(localAudioStream.getAudioTracks()[0]);
				activeAudioSender = audioSender;
			} else {
				activeAudioSender = pc.addTrack(
					localAudioStream.getAudioTracks()[0],
					localAudioStream,
				);
			}
		} else if (pc.addTransceiver) {
			pc.addTransceiver("audio", { direction: "recvonly" });
		}

		// Configure video transceiver
		const existingVideoSender = pc
			.getSenders()
			.find(
				(s) =>
					(s.track && s.track.kind === "video") ||
					(!s.track && (s as any).kind === "video"),
			);
		if (!existingVideoSender && pc.addTransceiver) {
			pc.addTransceiver("video", { direction: "sendrecv" });
		}

		await negotiateWebRTC(groupId, channelId, subchannelName);

		const videoBtn = document.getElementById("collab-btn-video");
		const wantVideo = videoBtn?.classList.contains("text-primary") ?? false;
		if (wantVideo) {
			await startCameraStream();
		}
		attachAllActiveStreams();
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
			setTimeout(resolve, 1200);
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
		startStatsMonitor(groupId);
	}
}

function startStatsMonitor(groupId: string) {
	if (statsMonitorTimer) clearInterval(statsMonitorTimer);
	statsMonitorTimer = setInterval(async () => {
		if (
			!pc ||
			pc.connectionState === "closed" ||
			pc.connectionState === "failed"
		) {
			clearInterval(statsMonitorTimer);
			statsMonitorTimer = null;
			return;
		}
		try {
			const stats = await pc.getStats();
			let quality = "good";
			let maxRtt = 0;
			stats.forEach((report) => {
				if (report.type === "candidate-pair" && report.state === "succeeded") {
					const rtt = report.currentRoundTripTime || 0;
					if (rtt > maxRtt) maxRtt = rtt;
				}
			});
			if (maxRtt > 0.4) {
				quality = "poor";
			} else if (maxRtt > 0.15) {
				quality = "fair";
			} else {
				quality = "good";
			}
			if (quality !== lastReportedQuality) {
				lastReportedQuality = quality;
				const htmx = (globalThis as any).htmx;
				if (htmx) {
					htmx.ajax(
						"POST",
						`/groups/${groupId}/collab/signal-strength?quality=${quality}`,
						{
							target: "#collab-dropdown-menu-content",
							swap: "outerHTML",
						},
					);
				}
			}
		} catch (_) {}
	}, 3000);
}

export function stopWebRTCSession() {
	if (statsMonitorTimer) {
		clearInterval(statsMonitorTimer);
		statsMonitorTimer = null;
	}
	lastReportedQuality = null;
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
	remoteVideoStream = null;
}

async function updatePeerVideoTrack() {
	if (!pc) return;

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
		}
		if (currentGroupId && currentChannelId && currentSubchannelName) {
			await negotiateWebRTC(
				currentGroupId,
				currentChannelId,
				currentSubchannelName,
			);
		}
	} else if (videoSender) {
		await videoSender.replaceTrack(null);
		if (currentGroupId && currentChannelId && currentSubchannelName) {
			await negotiateWebRTC(
				currentGroupId,
				currentChannelId,
				currentSubchannelName,
			);
		}
	}
}

// Media Capture Helpers
export async function startCameraStream(): Promise<boolean> {
	try {
		if (!cameraStream) {
			cameraStream = await navigator.mediaDevices.getUserMedia({ video: true });
		}
		attachCameraVideo();
		await updatePeerVideoTrack();
		return true;
	} catch (e) {
		console.warn("Could not capture camera stream:", e);
		return false;
	}
}

export async function stopCameraStream() {
	if (cameraStream) {
		cameraStream.getTracks().forEach((t) => t.stop());
		cameraStream = null;
	}
	const v = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (v) v.srcObject = null;
	await updatePeerVideoTrack();
}

export async function startScreenShare(): Promise<boolean> {
	try {
		const displayStream = await navigator.mediaDevices.getDisplayMedia({
			video: true,
			audio: true,
		});
		const vTrack = displayStream.getVideoTracks()[0];
		if (!vTrack) return false;

		screenStream = displayStream;
		attachScreenVideo();

		vTrack.addEventListener("ended", () => {
			stopScreenShare();
			const groupId = getGroupId();
			const htmx = (globalThis as any).htmx;
			if (groupId && htmx) {
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-screenshare?enabled=false`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		});

		updateScreenAudioMixer();
		await updatePeerVideoTrack();
		return true;
	} catch (e) {
		console.warn("Screen share cancelled or failed:", e);
		return false;
	}
}

export async function stopScreenShare() {
	if (screenStream) {
		screenStream.getTracks().forEach((t) => t.stop());
		screenStream = null;
	}
	const v = document.getElementById(
		"collab-screen-video",
	) as HTMLVideoElement | null;
	if (v) v.srcObject = null;
	updateScreenAudioMixer();
	await updatePeerVideoTrack();
}

export async function shareFile(file: File): Promise<boolean> {
	stopFileShare();

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
		const mediaEl = document.createElement("video");
		mediaEl.src = fileUrl;
		mediaEl.crossOrigin = "anonymous";
		mediaEl.loop = true;
		mediaEl.playsInline = true;
		// Keep local element unmuted so the sender can hear what they are playing
		mediaEl.muted = false;
		activeMediaElement = mediaEl;

		await mediaEl.play().catch(() => {});
		const captureFn =
			(mediaEl as any).captureStream || (mediaEl as any).mozCaptureStream;

		if (captureFn) {
			let stream: MediaStream = captureFn.call(mediaEl);

			// For audio-only files (MP3/WAV/OGG), generate a placeholder visualization canvas stream
			// with the audio track so peers receive both the audio and a visual tile.
			if (isAudio || stream.getVideoTracks().length === 0) {
				const canvas = document.createElement("canvas");
				canvas.width = 1280;
				canvas.height = 720;
				const ctx = canvas.getContext("2d");
				if (ctx) {
					ctx.fillStyle = "#1e1e24";
					ctx.fillRect(0, 0, canvas.width, canvas.height);
					ctx.fillStyle = "#38bdf8";
					ctx.font = "bold 36px sans-serif";
					ctx.textAlign = "center";
					ctx.fillText(`🎵 Now Playing: ${file.name}`, canvas.width / 2, 340);
					ctx.fillStyle = "#94a3b8";
					ctx.font = "20px sans-serif";
					ctx.fillText(
						`Audio Stream (${Math.round(file.size / 1024)} KB)`,
						canvas.width / 2,
						400,
					);
				}
				const canvasStream = canvas.captureStream(15);
				const vTrack = canvasStream.getVideoTracks()[0];
				if (vTrack) {
					stream.addTrack(vTrack);
				}
			}

			fileStream = stream;
			updateFileAudioMixer();
			attachFileVideo();
			await updatePeerVideoTrack();
			return true;
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
			img.onload = () => {
				const drawImage = () => {
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
					requestAnimationFrame(drawImage);
				};
				requestAnimationFrame(drawImage);
			};
		} else if (ctx) {
			ctx.fillStyle = "#1e1e24";
			ctx.fillRect(0, 0, canvas.width, canvas.height);
			ctx.fillStyle = "#e2e8f0";
			ctx.font = "bold 32px sans-serif";
			ctx.textAlign = "center";
			ctx.fillText(`Presentation: ${file.name}`, canvas.width / 2, 320);
			ctx.font = "20px sans-serif";
			ctx.fillStyle = "#94a3b8";
			ctx.fillText(
				`Sharing file (${Math.round(file.size / 1024)} KB)`,
				canvas.width / 2,
				380,
			);
		}

		const stream = canvas.captureStream(30);
		fileStream = stream;
		attachFileVideo();
		await updatePeerVideoTrack();
		return true;
	}
	return false;
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
	const v = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	if (v) v.srcObject = null;
	updateFileAudioMixer();
	await updatePeerVideoTrack();
}

// Media Attachment to Server-Rendered Elements
function attachCameraVideo() {
	const video = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (video && cameraStream) {
		if (video.srcObject !== cameraStream) {
			video.srcObject = cameraStream;
		}
		video.muted = true;
		video.play().catch(() => {});
	}
}

function attachScreenVideo() {
	const video = document.getElementById(
		"collab-screen-video",
	) as HTMLVideoElement | null;
	if (video && screenStream) {
		if (video.srcObject !== screenStream) {
			video.srcObject = screenStream;
		}
		video.muted = true;
		video.play().catch(() => {});
	}
}

function attachFileVideo() {
	const video = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	if (video && fileStream) {
		if (video.srcObject !== fileStream) {
			video.srcObject = fileStream;
		}
		video.muted = true;
		video.play().catch(() => {});
	}
}

function attachRemoteVideo(stream: MediaStream) {
	// Find any server-rendered remote video element that hasn't been attached or needs current stream
	const remoteVideos = document.querySelectorAll<HTMLVideoElement>(
		"video[id^='collab-remote-video-'], #collab-remote-video, #collab-video-dock video:not(#collab-local-video):not(#collab-screen-video):not(#collab-file-video)",
	);
	remoteVideos.forEach((video) => {
		if (video.srcObject !== stream) {
			video.srcObject = stream;
		}
		video.playsInline = true;
		video.autoplay = true;
		const playPromise = video.play();
		if (playPromise !== undefined) {
			playPromise.catch(() => {
				// Fallback to muted playback if autoplay policy prevents unmuted play
				video.muted = true;
				video.play().catch(() => {});
			});
		}
	});
}

function attachAllActiveStreams() {
	setTimeout(() => {
		attachCameraVideo();
		attachScreenVideo();
		attachFileVideo();
		if (remoteVideoStream) {
			attachRemoteVideo(remoteVideoStream);
		}
	}, 50);
}

export function checkAndSyncVoice() {
	const container = document.getElementById("collab-header-container");
	if (
		!container ||
		container.classList.contains("hidden") ||
		!container.hasAttribute("data-voice-active")
	) {
		stopWebRTCSession();
		return;
	}

	const groupId = container.getAttribute("data-group-id");
	const firstCh = container.querySelector(".channel-collab-entry");
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
	const micBtn = document.getElementById("collab-btn-mic");
	const isMicMuted = micBtn?.querySelector(".icon-[lucide--mic-off]") !== null;
	if (localAudioStream) {
		localAudioStream.getAudioTracks().forEach((track) => {
			track.enabled = !isMicMuted;
		});
	}

	// Sync local deafen/headphones
	const audioBtn = document.getElementById("collab-btn-audio");
	const isDeafened = audioBtn?.classList.contains("text-error") ?? false;
	if (remoteAudio) {
		remoteAudio.muted = isDeafened;
		if (!isDeafened && remoteAudio.paused) {
			remoteAudio.play().catch(() => {});
		}
	}

	// Sync camera state with server DOM: do not auto-prompt getUserMedia without user action!
	const videoTile = document.getElementById("video-tile-local");
	if (!videoTile && cameraStream) {
		stopCameraStream();
	}

	// Sync screen share state with server DOM: do not auto-prompt getDisplayMedia without user action!
	const screenTile = document.getElementById("video-tile-screen");
	if (!screenTile && screenStream) {
		stopScreenShare();
	}

	// Sync file share state with server DOM
	const fileTile = document.getElementById("video-tile-file");
	if (!fileTile && fileStream) {
		stopFileShare();
	}

	attachAllActiveStreams();
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

domLoaded.then(() => {
	checkAndSyncVoice();

	const onHtmxUpdate = () => {
		checkAndSyncVoice();
	};

	document.body.addEventListener("htmx:afterSwap", onHtmxUpdate);
	document.body.addEventListener("htmx:afterSettle", onHtmxUpdate);
	document.body.addEventListener("htmx:oobAfterSwap", onHtmxUpdate);
});
