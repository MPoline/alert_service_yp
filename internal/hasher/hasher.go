package hasher

type Hasher interface {
	CalculateHash([]byte, []byte) ([]byte, error)
}

func InitHasher(method string) Hasher {
	return NewSHA265Hasher()
}
