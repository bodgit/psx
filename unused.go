package psx

type unusedFrame struct {
	AvailableBlocks byte
	Reserved        [3]byte
	_               [4]byte
	LinkOrder       uint16
	_               [118]byte
}

func newUnusedFrame() unusedFrame {
	return unusedFrame{
		AvailableBlocks: blockUnavailable,
		Reserved:        [3]byte{0xff, 0xff, 0xff},
		LinkOrder:       lastLink,
	}
}
