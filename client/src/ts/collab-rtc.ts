import { domLoaded } from "./lib.ts";

let pc: RTCPeerConnection | null = null;
let localStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;
let isSyncingVideo = false;

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

		localStream = null;
		try {
			localStream = await navigator.mediaDevices.getUserMedia({
				audio: true,
				video: wantVideo,
			});
		} catch (err: any) {
			console.warn("Could not obtain audio/video stream:", err);
			if (wantVideo) {
				try {
					localStream = await navigator.mediaDevices.getUserMedia({
						audio: true,
					});
				} catch (audioErr: any) {
					console.warn("Could not obtain audio stream either:", audioErr);
				}
			}
		}

		if (localStream) {
			if (wantVideo && localStream.getVideoTracks().length > 0) {
				renderLocalVideo(localStream);
			}

			localStream.getTracks().forEach((track) => {
				if (pc && localStream) {
					pc.addTrack(track, localStream);
				}
			});
		} else {
			// No local tracks available (receive-only transceivers)
			if (pc.addTransceiver) {
				pc.addTransceiver("audio", { direction: "recvonly" });
				pc.addTransceiver("video", { direction: "recvonly" });
			}
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
	if (localStream) {
		localStream.getTracks().forEach((t) => t.stop());
		localStream = null;
	}
	if (pc) {
		pc.close();
		pc = null;
	}
	if (remoteAudio) {
		remoteAudio.srcObject = null;
		remoteAudio = null;
	}
	currentGroupId = null;
	currentChannelId = null;
	currentSubchannelName = null;
	removeLocalVideo();
	removeRemoteVideo();
}

export function renderLocalVideo(stream: MediaStream) {
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

export function renderRemoteVideoTrack(stream: MediaStream) {
	const dock = document.getElementById("collab-video-dock");
	const remoteTile = document.getElementById("video-tile-remote");
	const remoteVideo = document.getElementById(
		"collab-remote-video",
	) as HTMLVideoElement | null;
	if (dock) dock.classList.remove("hidden");
	if (remoteTile) remoteTile.classList.remove("hidden");
	if (remoteVideo) {
		if (remoteVideo.srcObject !== stream) {
			remoteVideo.srcObject = stream;
		}
		remoteVideo.muted = false;
		remoteVideo.play().catch(() => {});
	}
}

function removeLocalVideo() {
	const localTile = document.getElementById("video-tile-local");
	const localVideo = document.getElementById(
		"collab-local-video",
	) as HTMLVideoElement | null;
	if (localTile) localTile.classList.add("hidden");
	if (localVideo) localVideo.srcObject = null;
	checkDockVisibility();
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
	const remoteTile = document.getElementById("video-tile-remote");
	const isLocalVisible = localTile && !localTile.classList.contains("hidden");
	const isRemoteVisible =
		remoteTile && !remoteTile.classList.contains("hidden");
	if (dock && !isLocalVisible && !isRemoteVisible) {
		dock.classList.add("hidden");
	}
}

async function syncVideoTrack(wantVideo: boolean) {
	if (!pc || isStarting || isSyncingVideo) return;
	isSyncingVideo = true;

	try {
		const existingTrack = localStream?.getVideoTracks()[0];
		if (wantVideo) {
			if (existingTrack) {
				existingTrack.enabled = true;
				if (localStream) renderLocalVideo(localStream);
			} else {
				try {
					const vStream = await navigator.mediaDevices.getUserMedia({
						video: true,
					});
					const newTrack = vStream.getVideoTracks()[0];
					if (newTrack) {
						if (!localStream) {
							localStream = new MediaStream();
						}
						localStream.addTrack(newTrack);
						pc.addTrack(newTrack, localStream);
						renderLocalVideo(localStream);
						if (currentGroupId && currentChannelId && currentSubchannelName) {
							await negotiateWebRTC(
								currentGroupId,
								currentChannelId,
								currentSubchannelName,
							);
						}
					}
				} catch (e: any) {
					console.warn("Could not capture video track:", e);
				}
			}
		} else {
			if (existingTrack) {
				existingTrack.enabled = false;
				existingTrack.stop();
				localStream?.removeTrack(existingTrack);
				const sender = pc.getSenders().find((s) => s.track === existingTrack);
				if (sender) {
					pc.removeTrack(sender);
				}
				removeLocalVideo();
				if (currentGroupId && currentChannelId && currentSubchannelName) {
					await negotiateWebRTC(
						currentGroupId,
						currentChannelId,
						currentSubchannelName,
					);
				}
			} else {
				removeLocalVideo();
			}
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
	if (localStream) {
		localStream.getAudioTracks().forEach((track) => {
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
	if (wantVideo && localStream && localStream.getVideoTracks().length > 0) {
		renderLocalVideo(localStream);
	}
	syncVideoTrack(wantVideo);
}

domLoaded.then(() => {
	checkAndSyncVoice();

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
		}
	};

	document.body.addEventListener("htmx:afterSwap", onHtmxUpdate);
	document.body.addEventListener("htmx:afterSettle", onHtmxUpdate);
	document.body.addEventListener("htmx:oobAfterSwap", onHtmxUpdate);
});
