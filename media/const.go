package media

const (
	JPEG = iota
	PNG
)

var JPEG_PNG map[int][]byte

func init() {
	JPEG_PNG = make(map[int][]byte)
	JPEG_PNG[JPEG] = []byte{0xFF, 0xD8}
	JPEG_PNG[PNG] = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
}
