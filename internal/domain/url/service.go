package url

import (
	"math/rand"
	"net/url"
	"time"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type ShortCodeService struct {
	rnd *rand.Rand
}

func NewShortCodeService() ShortCodeService {
	return ShortCodeService{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s ShortCodeService) EncodeBase62(value int64) string {
	if value == 0 {
		return string(alphabet[0])
	}

	encoded := make([]byte, 0)
	for value > 0 {
		remainder := value % 62
		encoded = append(encoded, alphabet[int(remainder)])
		value /= 62
	}

	reverse(encoded)
	return string(encoded)
}

func (s ShortCodeService) GenerateID() int64 {
	min := int64Pow(62, 7)
	max := int64Pow(62, 8) - 1
	return min + s.rnd.Int63n(max-min+1)
}

func (s ShortCodeService) ValidateLongURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}

	return parsed.Scheme != "" && parsed.Host != ""
}

func reverse(values []byte) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func int64Pow(base int64, exp int) int64 {
	result := int64(1)
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
