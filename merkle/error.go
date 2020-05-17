package merkle

type InvalidProof struct{}

func (mi InvalidProof) Error() string{
	return "Payload is Invalid"
}
