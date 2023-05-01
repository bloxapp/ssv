package records

import (
	"crypto/sha256"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type HandshakeData struct {
	SenderPeerID    peer.ID
	RecipientPeerID peer.ID
	Timestamp       time.Time
	SenderPubKeyPem []byte
}

func (h *HandshakeData) Hash() [32]byte {
	sb := strings.Builder{}

	sb.WriteString(h.SenderPeerID.String())
	sb.WriteString(h.RecipientPeerID.String())
	sb.WriteString(strconv.FormatInt(h.Timestamp.Unix(), 10))
	sb.Write(h.SenderPubKeyPem)

	return sha256.Sum256([]byte(sb.String()))
}
