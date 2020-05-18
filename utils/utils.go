package mao_utils

func IsSameBytes(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i, val := range left {
		if right[i] != val {
			return false
		}
	}
	return true
}

