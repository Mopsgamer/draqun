package session

import (
	"errors"
	"fmt"
	"io"
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
	mu      sync.RWMutex
	clients map[uint64]*WebRTCClient // Key: userId
}

var WebRTCHub = &webRTCHub{
	clients: make(map[uint64]*WebRTCClient),
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
	h.mu.Lock()
	defer h.mu.Unlock()

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

	// Clean up existing client if present
	if old, exists := h.clients[userId]; exists {
		if old.PeerConnection != nil {
			_ = old.PeerConnection.Close()
		}
		delete(h.clients, userId)
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
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		_ = peerConnection.Close()
		return nil, err
	}

	videoTracks := make(map[string]*webrtc.TrackLocalStaticRTP, len(videoSourceTypes))
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
		if _, trackErr = peerConnection.AddTrack(videoTrack); trackErr != nil {
			_ = peerConnection.Close()
			return nil, trackErr
		}
		videoTracks[source] = videoTrack
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

	// Forward inbound RTP to other clients in same channel + subchannel
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		isAudio := remoteTrack.Kind() == webrtc.RTPCodecTypeAudio
		isVideo := remoteTrack.Kind() == webrtc.RTPCodecTypeVideo
		videoSource := ""
		if isVideo {
			transceiver := receiver.RTPTransceiver()
			if transceiver == nil {
				return
			}
			videoSource = midSources[transceiver.Mid()]
			if videoSource == "" {
				return
			}
		}

		trackDone := make(chan struct{})

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
				return
			}

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

			h.mu.RLock()
			for _, peer := range h.clients {
				if peer.UserId == userId {
					continue
				}
				if peer.ChannelId != channelId || peer.SubchannelName != subchannel {
					continue
				}

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

				if isAudio && peer.AudioTrack != nil {
					_ = peer.AudioTrack.WriteRTP(packet)
				} else if isVideo && peer.VideoTracks[videoSource] != nil {
					_ = peer.VideoTracks[videoSource].WriteRTP(packet)
				}
			}
			h.mu.RUnlock()
		}
	})

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
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

	h.clients[userId] = client

	// Notify existing video senders in room to send a keyframe for this new peer
	for _, peer := range h.clients {
		if peer.UserId != userId && peer.ChannelId == channelId && peer.SubchannelName == subchannel {
			if peer.PeerConnection != nil {
				for _, ssrc := range peer.VideoSSRCs {
					if ssrc != 0 {
						_ = peer.PeerConnection.WriteRTCP([]rtcp.Packet{
							&rtcp.PictureLossIndication{MediaSSRC: ssrc},
						})
					}
				}
			}
		}
	}
	return peerConnection.LocalDescription(), nil
}

func (h *webRTCHub) Close(userId uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userId]; ok {
		if c.PeerConnection != nil {
			_ = c.PeerConnection.Close()
		}
		delete(h.clients, userId)
	}
}
