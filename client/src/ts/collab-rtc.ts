import { domLoaded } from "./lib.ts";
import type * as HTMX from "htmx.org";

type HtmxApi = Omit<typeof HTMX.default, "ajax"> & {
	ajax: (
		verb: Uppercase<HTMX.HttpVerb>,
		path: string,
		context: Element | string | HTMX.HtmxAjaxHelperContext,
	) => Promise<void>;
};

function getHtmx(): HtmxApi | undefined {
	return (globalThis as typeof globalThis & { htmx?: HtmxApi }).htmx;
}

// Low-level WebRTC, Audio, and Video Stream state only
let pc: RTCPeerConnection | null = null;
let localAudioStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;
let voiceDisconnectTimer: ReturnType<typeof setTimeout> | null = null;
let staleMediaStateChecked = false;
let statsMonitorTimer: ReturnType<typeof setInterval> | null = null;
let lastReportedQuality: string | null = null;

type VideoSource = "camera" | "screen" | "file";

const videoSources: VideoSource[] = ["camera", "screen", "file"];

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
let activeFileUrl: string | null = null;
let activeFileImage: HTMLImageElement | null = null;
let videoTransceivers: Partial<Record<VideoSource, RTCRtpTransceiver>> = {};
let remoteVideoStreams: Partial<Record<VideoSource, MediaStream>> = {};

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
		const htmx = getHtmx();
		const cameraButton = document.getElementById("collab-btn-video");
		const cameraIsEnabled =
			cameraStream !== null || cameraButton?.classList.contains("text-primary");
		if (cameraIsEnabled) {
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
			if (success && !htmx) {
				await stopCameraStream();
				throw new Error(
					"HTMX is unavailable; camera state could not be updated.",
				);
			}
			if (success) {
				if (!htmx) {
					throw new Error(
						"HTMX is unavailable; camera state could not be updated.",
					);
				}
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
		const htmx = getHtmx();
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
			if (success && !htmx) {
				await stopScreenShare();
				throw new Error(
					"HTMX is unavailable; screen-share state could not be updated.",
				);
			}
			if (success) {
				if (!htmx) {
					throw new Error(
						"HTMX is unavailable; screen-share state could not be updated.",
					);
				}
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
		try {
			const groupId = getGroupId();
			const success = await shareFile(file);
			if (success) {
				if (!groupId) {
					throw new Error("No collaboration group is active.");
				}
				const htmx = getHtmx();
				if (!htmx) {
					throw new Error(
						"HTMX is unavailable; presentation state could not be updated.",
					);
				}
				htmx.ajax(
					"POST",
					`/groups/${groupId}/collab/toggle-presentation?enabled=true&filename=${encodeURIComponent(file.name)}`,
					{
						target: "#collab-dropdown-menu-content",
						swap: "outerHTML",
					},
				);
			}
		} catch (error) {
			console.error("Could not start file presentation:", error);
			await stopFileShare();
		} finally {
			input.value = "";
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

(
	globalThis as typeof globalThis & { collabClient: typeof collabClient }
).collabClient = collabClient;

// WebRTC Session
export async function startWebRTCSession(
	groupId: string,
	channelId: string,
	subchannelName: string,
) {
	if (isStarting) return;
	isStarting = true;
	const preserveLocalMedia =
		currentGroupId === groupId &&
		currentChannelId === channelId &&
		currentSubchannelName === subchannelName;
	try {
		stopWebRTCSession(preserveLocalMedia);
		currentGroupId = groupId;
		currentChannelId = channelId;
		currentSubchannelName = subchannelName;

		if (!staleMediaStateChecked) {
			staleMediaStateChecked = true;
			const hasStaleMediaState = [
				"collab-btn-video",
				"collab-btn-screenshare",
				"collab-btn-presentation",
			].some((id) =>
				document.getElementById(id)?.classList.contains("text-primary"),
			);
			if (hasStaleMediaState) {
				const htmx = getHtmx();
				if (!htmx) {
					throw new Error(
						"HTMX is unavailable; stale collaboration media cannot be cleared.",
					);
				}
				await htmx.ajax("POST", `/groups/${groupId}/collab/reset-media`, {
					target: "#collab-dropdown-menu-content",
					swap: "outerHTML",
				});
			}
		}

		pc = new RTCPeerConnection(rtcConfig);
		pc.oniceconnectionstatechange = () => {
			if (pc) {
				console.info("Collaboration ICE state:", pc.iceConnectionState);
			}
		};
		pc.onconnectionstatechange = () => {
			if (pc) {
				console.info("Collaboration peer state:", pc.connectionState);
			}
		};
		videoTransceivers = {
			camera: pc.addTransceiver("video", { direction: "sendrecv" }),
			screen: pc.addTransceiver("video", { direction: "sendrecv" }),
			file: pc.addTransceiver("video", { direction: "sendrecv" }),
		};

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
				const source =
					videoSources.find(
						(candidate) =>
							videoTransceivers[candidate]?.mid === event.transceiver.mid,
					) ??
					videoSources.find(
						(candidate) => videoTransceivers[candidate] === event.transceiver,
					) ??
					videoSources.find(
						(candidate) => stream.id === `draqun-collab-${candidate}`,
					);
				if (!source) {
					console.warn(
						`Received collaboration video on an unknown transceiver: ${event.transceiver.mid ?? event.track.id}`,
					);
					return;
				}
				console.info("Received collaboration video track:", {
					source,
					mid: event.transceiver.mid,
					trackId: event.track.id,
					readyState: event.track.readyState,
				});
				remoteVideoStreams[source] = stream;
				attachRemoteVideo(stream, source);
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

		await negotiateWebRTC(groupId, channelId, subchannelName);
		await updatePeerVideoTrack();

		attachAllActiveStreams();
	} catch (e) {
		console.warn("WebRTC startup notice:", e);
		stopWebRTCSession(preserveLocalMedia);
	} finally {
		isStarting = false;
	}
}

async function negotiateWebRTC(
	groupId: string,
	channelId: string,
	subchannelName: string,
) {
	const peerConnection = pc;
	if (!peerConnection) return;
	const offer = await peerConnection.createOffer();
	await peerConnection.setLocalDescription(offer);

	await new Promise<void>((resolve, reject) => {
		if (peerConnection.iceGatheringState === "complete") {
			resolve();
			return;
		}

		const timeout = window.setTimeout(() => {
			peerConnection.removeEventListener("icegatheringstatechange", checkState);
			reject(
				new Error("Timed out waiting for ICE candidates before signaling."),
			);
		}, 10000);
		const checkState = () => {
			if (peerConnection.iceGatheringState === "complete") {
				window.clearTimeout(timeout);
				peerConnection.removeEventListener(
					"icegatheringstatechange",
					checkState,
				);
				resolve();
			}
		};
		peerConnection.addEventListener("icegatheringstatechange", checkState);
		checkState();
	});

	const videoMids: Partial<Record<VideoSource, string>> = {};
	for (const source of videoSources) {
		const mid = videoTransceivers[source]?.mid;
		if (mid === null || mid === undefined) {
			throw new Error(`Missing ${source} video transceiver MID`);
		}
		videoMids[source] = mid;
	}

	const res = await fetch(`/groups/${groupId}/collab/signal`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({
			channel_id: parseInt(channelId, 10),
			subchannel_name: subchannelName || "1",
			sdp: peerConnection.localDescription?.sdp,
			type: peerConnection.localDescription?.type,
			video_mids: videoMids,
		}),
	});

	if (!res.ok) {
		throw new Error(
			`WebRTC signaling failed (${res.status}): ${await res.text()}`,
		);
	}
	const answer = await res.json();
	await peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
	startStatsMonitor(groupId);
}

function startStatsMonitor(groupId: string) {
	if (statsMonitorTimer) clearInterval(statsMonitorTimer);
	const timer = setInterval(async () => {
		if (
			!pc ||
			pc.connectionState === "closed" ||
			pc.connectionState === "failed"
		) {
			clearInterval(timer);
			statsMonitorTimer = null;
			return;
		}
		try {
			const stats = await pc.getStats();
			let quality = "good";
			let maxRtt = 0;
			const videoStats: Array<Record<string, string | number | null>> = [];
			const videoSenders = videoSources.map((source) => {
				const transceiver = videoTransceivers[source];
				const track = transceiver?.sender.track;
				return {
					source,
					mid: transceiver?.mid ?? null,
					trackId: track?.id ?? null,
					trackState: track?.readyState ?? null,
					trackEnabled: track?.enabled ?? null,
				};
			});
			stats.forEach((report) => {
				if (report.type === "candidate-pair" && report.state === "succeeded") {
					const rtt = report.currentRoundTripTime || 0;
					if (rtt > maxRtt) maxRtt = rtt;
				}
				if (report.type === "outbound-rtp" && report.kind === "video") {
					const outbound = report as RTCOutboundRtpStreamStats;
					videoStats.push({
						direction: "outbound",
						source:
							videoSources.find(
								(source) => videoTransceivers[source]?.mid === outbound.mid,
							) ?? null,
						mid: outbound.mid ?? null,
						bytes: outbound.bytesSent ?? 0,
						packets: outbound.packetsSent ?? 0,
						frames: outbound.framesEncoded ?? 0,
					});
				} else if (report.type === "inbound-rtp" && report.kind === "video") {
					const inbound = report as RTCInboundRtpStreamStats;
					videoStats.push({
						direction: "inbound",
						source:
							videoSources.find(
								(source) => videoTransceivers[source]?.mid === inbound.mid,
							) ?? null,
						mid: inbound.mid ?? null,
						bytes: inbound.bytesReceived ?? 0,
						packets: inbound.packetsReceived ?? 0,
						frames: inbound.framesDecoded ?? 0,
					});
				}
			});
			console.info("Collaboration video stats:", {
				connectionState: pc.connectionState,
				iceConnectionState: pc.iceConnectionState,
				senders: videoSenders,
				video: videoStats,
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
				const htmx = getHtmx();
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
		} catch (error) {
			console.warn("Could not read collaboration WebRTC stats:", error);
		}
	}, 3000);
	statsMonitorTimer = timer;
}

export function stopWebRTCSession(preserveLocalMedia = false) {
	if (voiceDisconnectTimer) {
		clearTimeout(voiceDisconnectTimer);
		voiceDisconnectTimer = null;
	}
	const peerConnection = pc;
	pc = null;
	peerConnection?.close();
	if (statsMonitorTimer) {
		clearInterval(statsMonitorTimer);
		statsMonitorTimer = null;
	}
	lastReportedQuality = null;
	if (!preserveLocalMedia) {
		stopCameraStream();
		stopScreenShare();
		stopFileShare();
	}
	teardownAudioMixer();

	if (localAudioStream) {
		localAudioStream.getTracks().forEach((t) => t.stop());
		localAudioStream = null;
	}
	activeAudioSender = null;
	if (remoteAudio) {
		remoteAudio.srcObject = null;
		remoteAudio = null;
	}
	currentGroupId = null;
	currentChannelId = null;
	currentSubchannelName = null;
	videoTransceivers = {};
	remoteVideoStreams = {};
}

async function updatePeerVideoTrack() {
	if (!pc) return;

	const streams: Record<VideoSource, MediaStream | null> = {
		camera: cameraStream,
		screen: screenStream,
		file: fileStream,
	};
	await Promise.all(
		videoSources.map(async (source) => {
			const sender = videoTransceivers[source]?.sender;
			if (!sender) {
				throw new Error(`Missing ${source} video sender`);
			}
			const track = streams[source]?.getVideoTracks()[0] ?? null;
			if (sender.track !== track) {
				await sender.replaceTrack(track);
			}
		}),
	);
}

function getMediaCaptureFunction(
	element: HTMLVideoElement,
): (() => MediaStream) | null {
	const mediaElement = element as HTMLVideoElement & {
		captureStream?: () => MediaStream;
		mozCaptureStream?: () => MediaStream;
	};
	const capture = mediaElement.captureStream ?? mediaElement.mozCaptureStream;
	return capture ? () => capture.call(mediaElement) : null;
}

function releaseActiveFileSource() {
	if (activeFileImage) {
		activeFileImage.src = "";
		activeFileImage = null;
	}
	if (activeMediaElement) {
		activeMediaElement.pause();
		activeMediaElement.removeAttribute("src");
		activeMediaElement.load();
		activeMediaElement.remove();
		activeMediaElement = null;
	}
	if (activeFileUrl) {
		URL.revokeObjectURL(activeFileUrl);
		activeFileUrl = null;
	}
}

// Media Capture Helpers
export async function startCameraStream(): Promise<boolean> {
	try {
		if (!cameraStream) {
			cameraStream = await navigator.mediaDevices.getUserMedia({ video: true });
			const track = cameraStream.getVideoTracks()[0];
			track?.addEventListener(
				"ended",
				() => {
					stopCameraStream();
					const groupId = getGroupId();
					const htmx = getHtmx();
					if (groupId && htmx) {
						htmx.ajax(
							"POST",
							`/groups/${groupId}/collab/toggle-video?enabled=false`,
							{
								target: "#collab-dropdown-menu-content",
								swap: "outerHTML",
							},
						);
					}
				},
				{ once: true },
			);
		}
		attachCameraVideo();
		await updatePeerVideoTrack();
		return true;
	} catch (e) {
		console.warn("Could not capture camera stream:", e);
		if (cameraStream) {
			cameraStream.getTracks().forEach((track) => track.stop());
			cameraStream = null;
		}
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
		if (!vTrack) {
			displayStream.getTracks().forEach((track) => track.stop());
			return false;
		}

		screenStream = displayStream;
		attachScreenVideo();

		vTrack.addEventListener("ended", () => {
			stopScreenShare();
			const groupId = getGroupId();
			const htmx = getHtmx();
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
		if (screenStream) {
			screenStream.getTracks().forEach((track) => track.stop());
			screenStream = null;
			updateScreenAudioMixer();
		}
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
	await stopFileShare();

	const fileType = file.type.toLowerCase();
	const fileName = file.name.toLowerCase();
	const isAudio =
		fileType.startsWith("audio/") ||
		/\.(mp3|wav|ogg|m4a|aac|flac)$/i.test(fileName);
	const isVideo =
		fileType.startsWith("video/") ||
		/\.(mp4|webm|mov|m4v|ogv)$/i.test(fileName);
	const isImage =
		fileType.startsWith("image/") ||
		/\.(png|jpe?g|gif|webp|bmp|svg)$/i.test(fileName);
	const isPresentationOrDoc =
		fileType.includes("presentation") ||
		fileType.includes("powerpoint") ||
		fileType.includes("pdf") ||
		isImage ||
		/\.(pptx?|pdf)$/i.test(fileName);

	if (!isAudio && !isVideo && !isPresentationOrDoc) {
		console.error(
			`Unsupported presentation file type: ${file.type || file.name}`,
		);
		return false;
	}

	activeFileUrl = URL.createObjectURL(file);
	try {
		if (isVideo || isAudio) {
			const mediaEl = document.createElement("video");
			mediaEl.src = activeFileUrl;
			mediaEl.preload = "auto";
			mediaEl.loop = true;
			mediaEl.playsInline = true;
			activeMediaElement = mediaEl;
			document.body.appendChild(mediaEl);

			try {
				await mediaEl.play();
			} catch (playError) {
				mediaEl.muted = true;
				await mediaEl.play();
				console.warn(
					"File playback started muted due to browser autoplay policy.",
					playError,
				);
			}

			const capture = getMediaCaptureFunction(mediaEl);
			if (!capture) {
				throw new Error(
					"This browser does not support capturing media file playback.",
				);
			}
			const stream = capture();
			if (isVideo && stream.getVideoTracks().length === 0) {
				throw new Error("The selected video did not produce a video track.");
			}

			if (isAudio || stream.getVideoTracks().length === 0) {
				const canvas = document.createElement("canvas");
				canvas.width = 1280;
				canvas.height = 720;
				const ctx = canvas.getContext("2d");
				if (!ctx) {
					throw new Error("Could not create the audio presentation canvas.");
				}
				ctx.fillStyle = "#1e1e24";
				ctx.fillRect(0, 0, canvas.width, canvas.height);
				ctx.fillStyle = "#38bdf8";
				ctx.font = "bold 36px sans-serif";
				ctx.textAlign = "center";
				ctx.fillText(`Now Playing: ${file.name}`, canvas.width / 2, 340);
				ctx.fillStyle = "#94a3b8";
				ctx.font = "20px sans-serif";
				ctx.fillText(
					`Audio Stream (${Math.round(file.size / 1024)} KB)`,
					canvas.width / 2,
					400,
				);
				const videoTrack = canvas.captureStream(15).getVideoTracks()[0];
				if (videoTrack) stream.addTrack(videoTrack);
			}

			fileStream = stream;
			updateFileAudioMixer();
		} else {
			const canvas = document.createElement("canvas");
			canvas.width = 1280;
			canvas.height = 720;
			const ctx = canvas.getContext("2d");
			if (!ctx) {
				throw new Error("Could not create the presentation canvas.");
			}

			ctx.fillStyle = "#1e1e24";
			ctx.fillRect(0, 0, canvas.width, canvas.height);
			if (isImage) {
				const image = new Image();
				activeFileImage = image;
				image.src = activeFileUrl;
				await image.decode();
				const scale = Math.min(
					canvas.width / image.width,
					canvas.height / image.height,
				);
				const x = canvas.width / 2 - (image.width / 2) * scale;
				const y = canvas.height / 2 - (image.height / 2) * scale;
				ctx.drawImage(image, x, y, image.width * scale, image.height * scale);
			} else {
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
			fileStream = canvas.captureStream(30);
		}

		attachFileVideo();
		await updatePeerVideoTrack();
		return true;
	} catch (error) {
		console.error("Could not capture the selected presentation:", error);
		await stopFileShare();
		return false;
	}
}

export async function stopFileShare() {
	if (fileStream) {
		fileStream.getTracks().forEach((t) => t.stop());
		fileStream = null;
	}
	if (activeMediaElement) {
		releaseActiveFileSource();
	} else if (activeFileImage || activeFileUrl) {
		releaseActiveFileSource();
	}
	const v = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	if (v) v.srcObject = null;
	updateFileAudioMixer();
	await updatePeerVideoTrack();
}

// Media Attachment to Server-Rendered Elements
function attachVideoStream(
	video: HTMLVideoElement,
	stream: MediaStream,
	label: string,
) {
	if (video.srcObject !== stream) {
		video.srcObject = stream;
	}
	video.muted = true;
	video.playsInline = true;
	video.autoplay = true;

	const play = () => {
		video.play().catch((error) => {
			console.warn(`Could not play ${label} video:`, error);
		});
	};
	if (video.readyState >= HTMLMediaElement.HAVE_METADATA) {
		play();
	} else {
		video.addEventListener("loadedmetadata", play, { once: true });
	}
}

function attachCameraVideo() {
	const video = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (video && cameraStream) {
		attachVideoStream(video, cameraStream, "local camera");
	}
}

function attachScreenVideo() {
	const video = document.getElementById(
		"collab-screen-video",
	) as HTMLVideoElement | null;
	if (video && screenStream) {
		attachVideoStream(video, screenStream, "local screen share");
	}
}

function attachFileVideo() {
	const video = document.getElementById(
		"collab-file-video",
	) as HTMLVideoElement | null;
	if (video && fileStream) {
		attachVideoStream(video, fileStream, "local presentation");
	}
}

function attachRemoteVideo(stream: MediaStream, source: VideoSource) {
	const remoteVideos = document.querySelectorAll<HTMLVideoElement>(
		`video[data-remote-stream-type="${source}"]`,
	);
	remoteVideos.forEach((video) => {
		attachVideoStream(video, stream, `remote ${source}`);
	});
}

function attachAllActiveStreams() {
	setTimeout(() => {
		attachCameraVideo();
		attachScreenVideo();
		attachFileVideo();
		for (const source of videoSources) {
			const stream = remoteVideoStreams[source];
			if (stream) {
				attachRemoteVideo(stream, source);
			}
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
		if (!voiceDisconnectTimer) {
			voiceDisconnectTimer = setTimeout(() => {
				voiceDisconnectTimer = null;
				const currentContainer = document.getElementById(
					"collab-header-container",
				);
				if (
					!currentContainer ||
					currentContainer.classList.contains("hidden") ||
					!currentContainer.hasAttribute("data-voice-active")
				) {
					stopWebRTCSession();
				}
			}, 500);
		}
		return;
	}
	if (voiceDisconnectTimer) {
		clearTimeout(voiceDisconnectTimer);
		voiceDisconnectTimer = null;
	}

	const groupId = container.getAttribute("data-group-id");
	const firstCh = container.querySelector(".channel-collab-entry");
	const channelId = firstCh?.getAttribute("data-channel-id");
	const subchannelName = firstCh?.getAttribute("data-subchannel-name") || "1";

	if (groupId && channelId) {
		if (
			!pc ||
			currentGroupId !== groupId ||
			currentChannelId !== channelId ||
			currentSubchannelName !== subchannelName
		) {
			startWebRTCSession(groupId, channelId, subchannelName);
			return;
		}
	}

	// Sync local mic mute
	const micBtn = document.getElementById("collab-btn-mic");
	const isMicMuted = micBtn?.classList.contains("text-error") ?? false;
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
			remoteAudio.play().catch((error) => {
				console.warn("Could not resume collaboration audio playback:", error);
			});
		}
	}

	// Allow OOB dock markup to arrive before treating a missing tile as a stopped share.
	const videoTile = document.getElementById("video-tile-local");
	const activeCameraStream = cameraStream;
	if (!videoTile && activeCameraStream) {
		setTimeout(() => {
			if (
				!document.getElementById("video-tile-local") &&
				cameraStream === activeCameraStream
			) {
				void stopCameraStream();
			}
		}, 250);
	}

	const screenTile = document.getElementById("video-tile-screen");
	const activeScreenStream = screenStream;
	if (!screenTile && activeScreenStream) {
		setTimeout(() => {
			if (
				!document.getElementById("video-tile-screen") &&
				screenStream === activeScreenStream
			) {
				void stopScreenShare();
			}
		}, 250);
	}

	const fileTile = document.getElementById("video-tile-file");
	const activeFileStream = fileStream;
	if (!fileTile && activeFileStream) {
		setTimeout(() => {
			if (
				!document.getElementById("video-tile-file") &&
				fileStream === activeFileStream
			) {
				void stopFileShare();
			}
		}, 250);
	}

	attachAllActiveStreams();
}

// Audio mixer helpers
function setupAudioMixer() {
	try {
		const AudioContextClass =
			window.AudioContext ||
			(window as Window & { webkitAudioContext?: typeof AudioContext })
				.webkitAudioContext;
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
		const audioSender = senders.find((s) => s.track?.kind === "audio");
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
