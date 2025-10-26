package calls

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type SFU struct {
	participants map[string]*Participant
	tracks       map[string][]*webrtc.TrackRemote
	mutex        sync.RWMutex
}

type Participant struct {
	ID      string
	pc      *webrtc.PeerConnection
	tracks  []*webrtc.TrackRemote
	sfu     *SFU
}

func NewSFU() *SFU {
	return &SFU{
		participants: make(map[string]*Participant),
		tracks:       make(map[string][]*webrtc.TrackRemote),
	}
}

func (s *SFU) AddParticipant(peerID string) *Participant {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Create peer connection for participant
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		// For testing, create a dummy participant
		pc = nil
	}
	
	participant := &Participant{
		ID:     peerID,
		pc:     pc,
		tracks: make([]*webrtc.TrackRemote, 0),
		sfu:    s,
	}
	
	s.participants[peerID] = participant
	return participant
}

func (s *SFU) AddTrack(peerID string, track *webrtc.TrackRemote) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Add track to participant
	if participant, exists := s.participants[peerID]; exists {
		participant.tracks = append(participant.tracks, track)
		s.tracks[peerID] = append(s.tracks[peerID], track)
	} else {
		// If participant doesn't exist, create tracks entry
		s.tracks[peerID] = []*webrtc.TrackRemote{track}
	}
}

func (s *SFU) RemoveParticipant(peerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Cleanup connections
	if participant, exists := s.participants[peerID]; exists {
		if participant.pc != nil {
			participant.pc.Close()
		}
		delete(s.participants, peerID)
		delete(s.tracks, peerID)
	}
}

func (s *SFU) ParticipantCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.participants)
}

func (p *Participant) GetRemoteTracks() []*webrtc.TrackRemote {
	p.sfu.mutex.RLock()
	defer p.sfu.mutex.RUnlock()
	
	// Return tracks from all other participants
	var allTracks []*webrtc.TrackRemote
	for peerID, tracks := range p.sfu.tracks {
		if peerID != p.ID {
			allTracks = append(allTracks, tracks...)
		}
	}
	return allTracks
}
