package url

import "testing"

func TestEncodeBase62(t *testing.T) {
	service := NewShortCodeService()
	encoded := service.EncodeBase62(11157)
	if encoded != "2TX" {
		t.Fatalf("expected 2TX, got %s", encoded)
	}
}

func TestGenerateIDRange(t *testing.T) {
	service := NewShortCodeService()
	id := service.GenerateID()

	min := int64Pow(62, 7)
	max := int64Pow(62, 8) - 1

	if id < min || id > max {
		t.Fatalf("id %d outside expected range [%d, %d]", id, min, max)
	}
}

func TestValidateLongURL(t *testing.T) {
	service := NewShortCodeService()

	if !service.ValidateLongURL("https://example.com/abc") {
		t.Fatal("expected valid URL")
	}

	if service.ValidateLongURL("not-a-valid-url") {
		t.Fatal("expected invalid URL")
	}
}
