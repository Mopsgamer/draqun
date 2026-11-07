import { domLoaded } from "./lib.ts";

let pc: RTCPeerConnection | null = null;
let localStream: MediaStream | null = null;
let remoteAudio: HTMLAudioElement | null = null;

const rtcConfig: RTCConfiguration = {
	iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
};

export async function startWebRTCSession(
	groupId: string,
	channelId: string,
	subchannelName: string,
) {
	try {
		stopWebRTCSession();

		pc = new RTCPeerConnection(rtcConfig);

		remoteAudio = new Audio();
		remoteAudio.autoplay = true;

		pc.ontrack = (event) => {
			if (event.track.kind === "audio" && remoteAudio) {
				remoteAudio.srcObject =
					event.streams[0] || new MediaStream([event.track]);
			} else if (event.track.kind === "video") {
				renderRemoteVideoTrack(
					event.streams[0] || new MediaStream([event.track]),
				);
			}
		};

		// Check camera toggle in DOM
		const videoBtn =
			document.getElementById("collab-btn-video") ||
			document.getElementById("voice-btn-video");
		const wantVideo = videoBtn?.classList.contains("text-primary") ?? false;

		localStream = await navigator.mediaDevices.getUserMedia({
			audio: true,
			video: wantVideo,
		});

		if (wantVideo) {
			renderLocalVideo(localStream);
		}

		localStream.getTracks().forEach((track) => {
			if (pc && localStream) {
				pc.addTrack(track, localStream);
			}
		});

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
	} catch (e) {
		console.warn("WebRTC startup notice (e.g. no audio/video hardware):", e);
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
			"fixed bottom-4 right-4 z-50 flex flex-wrap gap-2 p-2 bg-surface-overlay/90 backdrop-blur rounded-xl shadow-2xl border border-base-content/10 max-w-sm sm:max-w-md";
		document.body.appendChild(dock);
	}

	let tile = document.getElementById(`video-tile-${id}`);
	if (!tile) {
		tile = document.createElement("div");
		tile.id = `video-tile-${id}`;
		tile.className =
			"relative rounded-lg overflow-hidden bg-black aspect-video w-44 sm:w-52 border border-base-content/10";

		const video = document.createElement("video");
		video.autoplay = true;
		video.playsInline = true;
		video.muted = isMuted;
		video.className = "w-full h-full object-cover";
		tile.appendChild(video);

		const tag = document.createElement("span");
		tag.className =
			"absolute bottom-1 left-1 badge badge-neutral badge-xs text-[10px] opacity-80";
		tag.textContent = label;
		tile.appendChild(tag);

		const closeBtn = document.createElement("button");
		closeBtn.className =
			"btn btn-ghost btn-circle btn-xs absolute top-1 right-1 text-white hover:bg-black/50";
		closeBtn.innerHTML = '<span class="icon icon-[lucide--x] size-3"></span>';
		closeBtn.onclick = () => tile?.remove();
		tile.appendChild(closeBtn);

		dock.appendChild(tile);
	}

	const v = tile.querySelector<HTMLVideoElement>("video");
	if (v && v.srcObject !== stream) {
		v.srcObject = stream;
	}
}

export function renderLocalVideo(stream: MediaStream) {
	renderVideoTile("local", "You", stream, true);
}

export function renderRemoteVideoTrack(stream: MediaStream) {
	renderVideoTile("remote", "Participant", stream, false);
}

function removeVideoDock() {
	const dock = document.getElementById("collab-video-dock");
	if (dock) dock.remove();
}

function checkAndSyncVoice() {
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

	if (groupId && channelId && !pc) {
		startWebRTCSession(groupId, channelId, subchannelName);
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
}

domLoaded.then(() => {
	checkAndSyncVoice();

	// Reconnect or hook onto HTMX voice/collab events
	document.body.addEventListener("htmx:afterOnLoad", (evt: Event) => {
		const customEvt = evt as CustomEvent;
		const target = customEvt.detail?.target as HTMLElement | null;
		if (
			target &&
			(target.id === "collab-header-container" ||
				target.closest("#collab-header-container") ||
				target.id === "voice-header-container" ||
				target.closest("#voice-header-container"))
		) {
			checkAndSyncVoice();
		}
	});
});
