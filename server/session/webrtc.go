package session

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type WebRTCClient struct {
	UserId         uint64
	GroupId        uint64
	ChannelId      uint64
	SubchannelName string
	PeerConnection *webrtc.PeerConnection
	AudioTrack     *webrtc.TrackLocalStaticRTP
	VideoTracks    map[string]*webrtc.TrackLocalStaticRTP
	VideoSSRCs     map[string]uint32
}

var videoSourceTypes = []string{"camera", "screen", "file"}

type webRTCHub struct {
	mu         sync.RWMutex
	clients    map[uint64]*WebRTCClient // Key: userId
	generation map[uint64]uint64
}

var WebRTCHub = &webRTCHub{
	clients:    make(map[uint64]*WebRTCClient),
	generation: make(map[uint64]uint64),
}

var webrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func (h *webRTCHub) HandleOffer(
	userId, groupId, channelId uint64,
	subchannel string,
	videoMids map[string]string,
	offer webrtc.SessionDescription,
) (*webrtc.SessionDescription, error) {
	midSources := make(map[string]string, len(videoSourceTypes))
	for _, source := range videoSourceTypes {
		mid, ok := videoMids[source]
		if !ok || mid == "" {
			return nil, fmt.Errorf("missing video transceiver MID for %q", source)
		}
		if _, exists := midSources[mid]; exists {
			return nil, fmt.Errorf("duplicate video transceiver MID %q", mid)
		}
		midSources[mid] = source
	}

	h.mu.Lock()
	h.generation[userId]++
	generation := h.generation[userId]
	old := h.clients[userId]
	delete(h.clients, userId)
	h.mu.Unlock()
	if old != nil && old.PeerConnection != nil {
		_ = old.PeerConnection.Close()
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
	if err != nil {
		return nil, err
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio",
		"draqun-collab",
	)
	if err != nil {
		_ = peerConnection.Close()
		return nil, err
	}

	// Add local tracks so client can receive forwarded media
	audioSender, err := peerConnection.AddTrack(audioTrack)
	if err != nil {
		_ = peerConnection.Close()
		return nil, err
	}

	videoTracks := make(map[string]*webrtc.TrackLocalStaticRTP, len(videoSourceTypes))
	videoSenders := make(map[string]*webrtc.RTPSender, len(videoSourceTypes))
	for _, source := range videoSourceTypes {
		videoTrack, trackErr := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
			"collab-video-"+source,
			"draqun-collab-"+source,
		)
		if trackErr != nil {
			_ = peerConnection.Close()
			return nil, trackErr
		}
		videoSender, addTrackErr := peerConnection.AddTrack(videoTrack)
		if addTrackErr != nil {
			_ = peerConnection.Close()
			return nil, addTrackErr
		}
		var videoTransceiver *webrtc.RTPTransceiver
		for _, transceiver := range peerConnection.GetTransceivers() {
			if transceiver.Sender() == videoSender {
				videoTransceiver = transceiver
				break
			}
		}
		if videoTransceiver == nil {
			_ = peerConnection.Close()
			return nil, fmt.Errorf("missing %s video transceiver", source)
		}
		var vp8Codecs []webrtc.RTPCodecParameters
		for _, codec := range videoSender.GetParameters().Codecs {
			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeVP8) {
				vp8Codecs = append(vp8Codecs, codec)
			}
		}
		if len(vp8Codecs) == 0 {
			_ = peerConnection.Close()
			return nil, errors.New("VP8 is unavailable for collaboration video relay")
		}
		if codecErr := videoTransceiver.SetCodecPreferences(vp8Codecs); codecErr != nil {
			_ = peerConnection.Close()
			return nil, fmt.Errorf("set %s video codec preference: %w", source, codecErr)
		}
		videoTracks[source] = videoTrack
		videoSenders[source] = videoSender
	}

	client := &WebRTCClient{
		UserId:         userId,
		GroupId:        groupId,
		ChannelId:      channelId,
		SubchannelName: subchannel,
		PeerConnection: peerConnection,
		AudioTrack:     audioTrack,
		VideoTracks:    videoTracks,
		VideoSSRCs:     make(map[string]uint32, len(videoSourceTypes)),
	}
	go drainSenderRTCP(h, client, audioSender, "")
	for source, sender := range videoSenders {
		go drainSenderRTCP(h, client, sender, source)
	}

	// Forward inbound RTP to other clients in same channel + subchannel
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		isAudio := remoteTrack.Kind() == webrtc.RTPCodecTypeAudio
		isVideo := remoteTrack.Kind() == webrtc.RTPCodecTypeVideo
		videoSource := ""
		if isVideo {
			transceiver := receiver.RTPTransceiver()
			if transceiver == nil {
				slog.Warn("collaboration video track has no RTP transceiver", "user_id", userId)
				return
			}
			videoSource = midSources[transceiver.Mid()]
			if videoSource == "" {
				slog.Warn(
					"collaboration video track has unknown MID",
					"user_id", userId,
					"mid", transceiver.Mid(),
				)
				return
			}
		}
		slog.Info(
			"collaboration media track received",
			"user_id", userId,
			"kind", remoteTrack.Kind(),
			"source", videoSource,
		)

		trackDone := make(chan struct{})
		loggedWriteErrors := make(map[uint64]bool)
		loggedFirstRelay := make(map[uint64]bool)
		packetCount := 0

		if isVideo {
			h.mu.Lock()
			client.VideoSSRCs[videoSource] = uint32(remoteTrack.SSRC())
			h.mu.Unlock()
			_ = peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
			})

			// Periodically request keyframes every 1 second to ensure new joiners and track changes decode smoothly
			go func() {
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-trackDone:
						return
					case <-ticker.C:
						if peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed ||
							peerConnection.ConnectionState() == webrtc.PeerConnectionStateFailed {
							return
						}
						_ = peerConnection.WriteRTCP([]rtcp.Packet{
							&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())},
						})
					}
				}
			}()
		}

		defer close(trackDone)

		for {
			packet, _, readErr := remoteTrack.ReadRTP()
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					return
				}
				slog.Warn("failed to read collaboration RTP", "user_id", userId, "error", readErr)
				return
			}
			packetCount++

			// Check input mute on sender: if global input is muted OR channel input is muted, drop audio
			if isAudio {
				senderSess := CollabHub.GetUserSession(userId)
				if senderSess != nil {
					if senderSess.GlobalInputMuted {
						continue
					}
					if chSess, ok := senderSess.Channels[channelId]; ok && chSess.InputMuted {
						continue
					}
				}
			}

			// RTP header extensions are negotiated per peer connection and cannot
			// be forwarded unchanged to a different peer connection.
			packet.Header.Extensions = nil
			packet.Header.Extension = false

			h.mu.RLock()
			matchedPeers := 0
			for _, peer := range h.clients {
				if peer.UserId == userId {
					continue
				}
				if peer.GroupId != groupId || peer.ChannelId != channelId || peer.SubchannelName != subchannel {
					continue
				}
				matchedPeers++

				// Check output mute on receiver
				userSess := CollabHub.GetUserSession(peer.UserId)
				if userSess != nil {
					if isAudio && (userSess.GlobalOutputMuted || userSess.UserMutes[userId]) {
						continue
					}
					if isAudio {
						if chSess, ok := userSess.Channels[channelId]; ok && chSess.OutputMuted {
							continue
						}
					}
				}

				var writeErr error
				relayed := false
				if isAudio && peer.AudioTrack != nil {
					writeErr = peer.AudioTrack.WriteRTP(packet)
					relayed = writeErr == nil
				} else if isVideo && peer.VideoTracks[videoSource] != nil {
					writeErr = peer.VideoTracks[videoSource].WriteRTP(packet)
					relayed = writeErr == nil
				}
				if writeErr != nil && !loggedWriteErrors[peer.UserId] {
					loggedWriteErrors[peer.UserId] = true
					slog.Warn(
						"failed to relay collaboration RTP",
						"sender_id", userId,
						"receiver_id", peer.UserId,
						"kind", remoteTrack.Kind(),
						"source", videoSource,
						"error", writeErr,
					)
				}
				if relayed && !loggedFirstRelay[peer.UserId] {
					loggedFirstRelay[peer.UserId] = true
					slog.Info(
						"first collaboration RTP packet relayed",
						"sender_id", userId,
						"receiver_id", peer.UserId,
						"channel_id", channelId,
						"subchannel", subchannel,
						"kind", remoteTrack.Kind(),
						"source", videoSource,
					)
				}
			}
			h.mu.RUnlock()
			if packetCount == 1 {
				slog.Info(
					"first collaboration RTP packet received",
					"sender_id", userId,
					"channel_id", channelId,
					"subchannel", subchannel,
					"kind", remoteTrack.Kind(),
					"source", videoSource,
					"matched_peers", matchedPeers,
				)
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		slog.Info("collaboration ICE state changed", "user_id", userId, "state", state.String())
	})
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		slog.Info("collaboration peer connection state changed", "user_id", userId, "state", state)
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			h.mu.Lock()
			if curr, ok := h.clients[userId]; ok && curr == client {
				delete(h.clients, userId)
			}
			h.mu.Unlock()
		}
	})

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		_ = peerConnection.Close()
		return nil, err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		_ = peerConnection.Close()
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		_ = peerConnection.Close()
		return nil, err
	}
	<-gatherComplete

	h.mu.Lock()
	if h.generation[userId] != generation {
		h.mu.Unlock()
		_ = peerConnection.Close()
		return nil, errors.New("collaboration offer was superseded")
	}
	h.clients[userId] = client

	type keyframeRequest struct {
		pc   *webrtc.PeerConnection
		ssrc uint32
	}
	var keyframeRequests []keyframeRequest
	for _, peer := range h.clients {
		if peer.UserId != userId && peer.GroupId == groupId && peer.ChannelId == channelId && peer.SubchannelName == subchannel {
			if peer.PeerConnection != nil {
				for _, ssrc := range peer.VideoSSRCs {
					if ssrc != 0 {
						keyframeRequests = append(keyframeRequests, keyframeRequest{
							pc:   peer.PeerConnection,
							ssrc: ssrc,
						})
					}
				}
			}
		}
	}
	h.mu.Unlock()

	// Notify existing video senders in room to send a keyframe for this new peer.
	for _, request := range keyframeRequests {
		_ = request.pc.WriteRTCP([]rtcp.Packet{
			&rtcp.PictureLossIndication{MediaSSRC: request.ssrc},
		})
	}
	return peerConnection.LocalDescription(), nil
}

func drainSenderRTCP(h *webRTCHub, receiver *WebRTCClient, sender *webrtc.RTPSender, source string) {
	buf := make([]byte, 1500)
	for {
		n, _, err := sender.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Warn(
					"failed to read collaboration sender RTCP",
					"user_id", receiver.UserId,
					"source", source,
					"error", err,
				)
			}
			return
		}
		if source == "" {
			continue
		}
		packets, err := rtcp.Unmarshal(buf[:n])
		if err != nil {
			slog.Warn(
				"failed to parse collaboration sender RTCP",
				"user_id", receiver.UserId,
				"source", source,
				"error", err,
			)
			continue
		}
		forwardVideoFeedback(h, receiver, source, packets)
	}
}

func forwardVideoFeedback(h *webRTCHub, receiver *WebRTCClient, source string, packets []rtcp.Packet) {
	type publisher struct {
		peerConnection *webrtc.PeerConnection
		ssrc           uint32
		userId         uint64
	}

	h.mu.RLock()
	publishers := make([]publisher, 0)
	for _, peer := range h.clients {
		if peer.UserId == receiver.UserId ||
			peer.GroupId != receiver.GroupId ||
			peer.ChannelId != receiver.ChannelId ||
			peer.SubchannelName != receiver.SubchannelName {
			continue
		}
		if ssrc := peer.VideoSSRCs[source]; ssrc != 0 && peer.PeerConnection != nil {
			publishers = append(publishers, publisher{
				peerConnection: peer.PeerConnection,
				ssrc:           ssrc,
				userId:         peer.UserId,
			})
		}
	}
	h.mu.RUnlock()

	for _, publisher := range publishers {
		var feedback []rtcp.Packet
		for _, packet := range packets {
			switch packet := packet.(type) {
			case *rtcp.TransportLayerNack:
				rewritten := *packet
				rewritten.MediaSSRC = publisher.ssrc
				feedback = append(feedback, &rewritten)
			case *rtcp.PictureLossIndication:
				rewritten := *packet
				rewritten.MediaSSRC = publisher.ssrc
				feedback = append(feedback, &rewritten)
			case *rtcp.FullIntraRequest:
				rewritten := *packet
				rewritten.MediaSSRC = publisher.ssrc
				rewritten.FIR = append([]rtcp.FIREntry(nil), packet.FIR...)
				for i := range rewritten.FIR {
					rewritten.FIR[i].SSRC = publisher.ssrc
				}
				feedback = append(feedback, &rewritten)
			}
		}
		if len(feedback) == 0 {
			continue
		}
		if err := publisher.peerConnection.WriteRTCP(feedback); err != nil {
			slog.Warn(
				"failed to forward collaboration video feedback",
				"receiver_id", receiver.UserId,
				"publisher_id", publisher.userId,
				"source", source,
				"error", err,
			)
		}
	}
}

func (h *webRTCHub) Close(userId uint64) {
	h.mu.Lock()
	h.generation[userId]++
	client := h.clients[userId]
	delete(h.clients, userId)
	h.mu.Unlock()

	if client != nil && client.PeerConnection != nil {
		_ = client.PeerConnection.Close()
	}
}
