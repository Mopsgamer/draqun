import { domLoaded } from "./lib.ts";

let pc: RTCPeerConnection | null = null;
let localStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;
let currentGroupId: string | null = null;
let currentChannelId: string | null = null;
let currentSubchannelName: string | null = null;
let isStarting = false;

const rtcConfig: RTCConfiguration = {
	iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
};

function showPermissionAlert(message: string) {
	let container = document.getElementById("collab-alerts-container");
	if (!container) {
		container = document.createElement("div");
		container.id = "collab-alerts-container";
		container.className = "toast toast-top toast-end z-50";
		document.body.appendChild(container);
	}
	const alert = document.createElement("div");
	alert.className =
		"alert alert-warning text-xs shadow-lg flex items-center gap-2 py-2 px-3";
	alert.innerHTML = `
		<span class="icon icon-[lucide--alert-triangle] size-4"></span>
		<span>${message}</span>
	`;
	container.appendChild(alert);
	setTimeout(() => {
		alert.remove();
		if (container && container.children.length === 0) {
			container.remove();
		}
	}, 5000);
}

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

				// Only render tile when track is unmuted / receiving frames
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
					// Check if stream actually plays frames before rendering black box
					const tempVideo = document.createElement("video");
					tempVideo.srcObject = remoteStream;
					tempVideo.muted = true;
					tempVideo.playsInline = true;
					tempVideo.onloadedmetadata = () => {
						renderRemoteVideoTrack(remoteStream);
					};
					tempVideo.play().catch(() => {});
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
		} catch (err) {
			console.warn("Could not obtain audio/video stream:", err);
			if (wantVideo) {
				// Retry with audio only
				try {
					localStream = await navigator.mediaDevices.getUserMedia({
						audio: true,
					});
					showPermissionAlert(
						"Camera access denied or unavailable. Joined with audio only.",
					);
				} catch (audioErr) {
					console.warn("Could not obtain audio stream either:", audioErr);
					showPermissionAlert(
						"Microphone access denied or unavailable. Joined in listen-only mode.",
					);
				}
			} else {
				showPermissionAlert(
					"Microphone access denied or unavailable. Joined in listen-only mode.",
				);
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
	removeVideoDock();
}

function renderVideoTile(
	id: string,
	label: string,
	stream: MediaStream,
	isMuted: boolean,
) {
	let dock = document.getElementById("collab-video-dock");
	if (!dock) {
		dock = document.createElement("div");
		dock.id = "collab-video-dock";
		dock.className =
			"fixed bottom-6 right-6 z-50 flex flex-wrap gap-3 pointer-events-none";
		document.body.appendChild(dock);
	}

	let tile = document.getElementById(`video-tile-${id}`);
	if (!tile) {
		tile = document.createElement("div");
		tile.id = `video-tile-${id}`;
		tile.className =
			"pointer-events-auto relative rounded-xl overflow-hidden bg-neutral shadow-2xl border border-base-content/15 aspect-video w-48 sm:w-60 transition-transform duration-200";

		const video = document.createElement("video");
		video.autoplay = true;
		video.playsInline = true;
		video.muted = isMuted;
		video.className = "w-full h-full object-cover";
		tile.appendChild(video);

		const tag = document.createElement("span");
		tag.className =
			"absolute bottom-1.5 left-1.5 badge badge-neutral badge-xs text-[10px] font-medium opacity-90 shadow";
		tag.textContent = label;
		tile.appendChild(tag);

		const closeBtn = document.createElement("button");
		closeBtn.className =
			"btn btn-ghost btn-circle btn-xs absolute top-1.5 right-1.5 text-white/90 hover:text-white hover:bg-black/40";
		closeBtn.innerHTML = '<span class="icon icon-[lucide--x] size-3.5"></span>';
		closeBtn.onclick = () => {
			tile?.remove();
			checkDockEmpty();
		};
		tile.appendChild(closeBtn);

		dock.appendChild(tile);
	}

	const v = tile.querySelector<HTMLVideoElement>("video");
	if (v && v.srcObject !== stream) {
		v.srcObject = stream;
	}
}

function checkDockEmpty() {
	const dock = document.getElementById("collab-video-dock");
	if (dock && dock.children.length === 0) {
		dock.remove();
	}
}

export function renderLocalVideo(stream: MediaStream) {
	renderVideoTile("local", "You", stream, true);
}

export function renderRemoteVideoTrack(stream: MediaStream) {
	renderVideoTile("remote", "Participant", stream, false);
}

function removeLocalVideo() {
	const tile = document.getElementById("video-tile-local");
	if (tile) tile.remove();
	checkDockEmpty();
}

function removeRemoteVideo() {
	const tile = document.getElementById("video-tile-remote");
	if (tile) tile.remove();
	checkDockEmpty();
}

function removeVideoDock() {
	const dock = document.getElementById("collab-video-dock");
	if (dock) dock.remove();
}

async function syncVideoTrack(wantVideo: boolean) {
	if (!pc) return;

	if (wantVideo) {
		let videoTrack = localStream?.getVideoTracks()[0];
		if (videoTrack) {
			videoTrack.enabled = true;
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
			} catch (e) {
				console.warn("Could not capture video track:", e);
				showPermissionAlert("Camera access denied or unavailable.");
			}
		}
	} else {
		const videoTrack = localStream?.getVideoTracks()[0];
		if (videoTrack) {
			videoTrack.enabled = false;
			videoTrack.stop();
			localStream?.removeTrack(videoTrack);
			const sender = pc.getSenders().find((s) => s.track === videoTrack);
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
