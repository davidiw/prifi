package sign

import (
	"github.com/dedis/prifi/coco/coconet"
)

// Functions used in collective signing
// That are direclty related to the generation/ verification/ sending
// of the Simple Combined Public Key Signature

// Send children challenges
func (sn *Node) SendChildrenChallenges(chm *ChallengeMessage) error {
	for _, child := range sn.Children() {
		var messg coconet.BinaryMarshaler
		messg = &SigningMessage{Type: Challenge, Chm: chm}

		if err := child.Put(messg); err != nil {
			return err
		}
	}

	return nil
}
