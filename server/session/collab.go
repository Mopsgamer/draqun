package session

import (
	"sync"
	"time"
)

type CollabParticipant struct {
	UserId      uint64 `json:"user_id"`
	Moniker     string `json:"moniker"`
	AvatarUrl   string `json:"avatar_url"`
	OutputMuted bool   `json:"output_muted"`
	InputMuted  bool   `json:"input_muted"`
}

type CollabChannelSession struct {
	ChannelId      uint64                       `json:"channel_id"`
	ChannelName    string                       `json:"channel_name"`
	SubchannelName string                       `json:"subchannel_name"` // e.g. "main", "team-a", "team-b"
	InputMuted     bool                         `json:"input_muted"`
	OutputMuted    bool                         `json:"output_muted"`
	Participants   map[uint64]*CollabParticipant `json:"participants"`
}

type UserCollabSession struct {
	UserId            uint64                           `json:"user_id"`
	GroupId           uint64                           `json:"group_id"`
	GlobalInputMuted  bool                             `json:"global_input_muted"`
	GlobalOutputMuted bool                             `json:"global_output_muted"`
	VideoEnabled      bool                             `json:"video_enabled"`
	BitrateKbps       int                              `json:"bitrate_kbps"`
	SignalStrength    string                           `json:"signal_strength"`
	Channels          map[uint64]*CollabChannelSession `json:"channels"`
	UserMutes         map[uint64]bool                  `json:"user_mutes"`
	LastSeen          time.Time                        `json:"last_seen"`
}

type collabHub struct {
	mu       sync.RWMutex
	sessions map[uint64]*UserCollabSession // Key: userId
}

var CollabHub = &collabHub{
	sessions: make(map[uint64]*UserCollabSession),
}

func (h *collabHub) GetUserSession(userId uint64) *UserCollabSession {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[userId]
}

func (h *collabHub) GetActiveSubchannels(channelId uint64) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomSet := make(map[string]bool)
	for _, sess := range h.sessions {
		if ch, ok := sess.Channels[channelId]; ok && len(ch.Participants) > 0 {
			room := ch.SubchannelName
			if room == "" {
				room = "1"
			}
			roomSet[room] = true
		}
	}

	var rooms []string
	for r := range roomSet {
		rooms = append(rooms, r)
	}
	return rooms
}

func (h *collabHub) GetRoomParticipants(channelId uint64, subchannel string) []CollabParticipant {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var list []CollabParticipant
	for _, sess := range h.sessions {
		if ch, ok := sess.Channels[channelId]; ok {
			room := ch.SubchannelName
			if room == "" {
				room = "1"
			}
			if room == subchannel {
				if part, hasPart := ch.Participants[sess.UserId]; hasPart {
					list = append(list, *part)
				}
			}
		}
	}
	return list
}

func (h *collabHub) JoinSubchannel(
	userId uint64,
	moniker, avatarUrl string,
	groupId uint64,
	channelId uint64,
	channelName string,
	subchannelName string,
) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subchannelName == "" {
		subchannelName = "main"
	}

	sess, exists := h.sessions[userId]
	if !exists {
		sess = &UserCollabSession{
			UserId:            userId,
			GroupId:           groupId,
			GlobalInputMuted:  false,
			GlobalOutputMuted: false,
			VideoEnabled:      false,
			BitrateKbps:       64,
			SignalStrength:    "good",
			Channels:          make(map[uint64]*CollabChannelSession),
			UserMutes:         make(map[uint64]bool),
			LastSeen:          time.Now(),
		}
		h.sessions[userId] = sess
	}
	sess.GroupId = groupId
	sess.LastSeen = time.Now()

	chSess, chExists := sess.Channels[channelId]
	if !chExists {
		chSess = &CollabChannelSession{
			ChannelId:      channelId,
			ChannelName:    channelName,
			SubchannelName: subchannelName,
			InputMuted:     false,
			OutputMuted:    false,
			Participants:   make(map[uint64]*CollabParticipant),
		}
		sess.Channels[channelId] = chSess
	} else {
		chSess.SubchannelName = subchannelName
	}

	selfPart := &CollabParticipant{
		UserId:      userId,
		Moniker:     moniker,
		AvatarUrl:   avatarUrl,
		OutputMuted: false,
		InputMuted:  sess.GlobalInputMuted || chSess.InputMuted,
	}

	// Update participants matching same channel and virtual subchannel
	for _, otherSess := range h.sessions {
		if otherCh, ok := otherSess.Channels[channelId]; ok {
			if otherCh.SubchannelName == subchannelName {
				otherCh.Participants[userId] = &CollabParticipant{
					UserId:      userId,
					Moniker:     moniker,
					AvatarUrl:   avatarUrl,
					OutputMuted: otherSess.UserMutes[userId],
					InputMuted:  sess.GlobalInputMuted || chSess.InputMuted,
				}
				if otherSess.UserId != userId {
					if otherPart, hasPart := otherCh.Participants[otherSess.UserId]; hasPart {
						chSess.Participants[otherSess.UserId] = &CollabParticipant{
							UserId:      otherPart.UserId,
							Moniker:     otherPart.Moniker,
							AvatarUrl:   otherPart.AvatarUrl,
							OutputMuted: sess.UserMutes[otherSess.UserId],
							InputMuted:  otherPart.InputMuted,
						}
					}
				}
			} else {
				// Different virtual subchannel, remove cross-visibility
				delete(otherCh.Participants, userId)
				delete(chSess.Participants, otherSess.UserId)
			}
		}
	}
	chSess.Participants[userId] = selfPart

	return sess
}

func (h *collabHub) LeaveChannel(userId uint64, channelId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}

	delete(sess.Channels, channelId)

	for _, otherSess := range h.sessions {
		if otherCh, ok := otherSess.Channels[channelId]; ok {
			delete(otherCh.Participants, userId)
		}
	}

	return sess
}

func (h *collabHub) DisconnectAll(userId uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return
	}

	for chId := range sess.Channels {
		for _, otherSess := range h.sessions {
			if otherCh, ok := otherSess.Channels[chId]; ok {
				delete(otherCh.Participants, userId)
			}
		}
	}

	delete(h.sessions, userId)
}

func (h *collabHub) ToggleGlobalInput(userId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	sess.GlobalInputMuted = !sess.GlobalInputMuted
	return sess
}

func (h *collabHub) ToggleGlobalOutput(userId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	sess.GlobalOutputMuted = !sess.GlobalOutputMuted
	return sess
}

func (h *collabHub) ToggleVideo(userId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	sess.VideoEnabled = !sess.VideoEnabled
	return sess
}

func (h *collabHub) SetBitrate(userId uint64, bitrate int) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	sess.BitrateKbps = bitrate
	return sess
}

func (h *collabHub) ToggleChannelInput(userId uint64, channelId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	if ch, ok := sess.Channels[channelId]; ok {
		ch.InputMuted = !ch.InputMuted
	}
	return sess
}

func (h *collabHub) ToggleChannelOutput(userId uint64, channelId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	if ch, ok := sess.Channels[channelId]; ok {
		ch.OutputMuted = !ch.OutputMuted
	}
	return sess
}

func (h *collabHub) ToggleUserMute(userId uint64, targetUserId uint64) *UserCollabSession {
	h.mu.Lock()
	defer h.mu.Unlock()

	sess, exists := h.sessions[userId]
	if !exists {
		return nil
	}
	sess.UserMutes[targetUserId] = !sess.UserMutes[targetUserId]
	for _, ch := range sess.Channels {
		if p, ok := ch.Participants[targetUserId]; ok {
			p.OutputMuted = sess.UserMutes[targetUserId]
		}
	}
	return sess
}
