package icons

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	"github.com/gen2brain/webp"
)

func testPNG(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, size, size))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestConvert(t *testing.T) {
	tests := []struct {
		profile Profile
		size    int
		want    int
	}{
		{Icon, 128, 64},
		{Icon, 64, 64},
		{Icon, 32, 32},
		{Lossless, 22, 22},
	}
	for _, test := range tests {
		raw, err := convert(testPNG(t, test.size), test.profile)
		if err != nil {
			t.Fatalf("%s at %d: %v", test.profile, test.size, err)
		}
		config, err := webp.DecodeConfig(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("%s at %d: %v", test.profile, test.size, err)
		}
		if config.Width != test.want || config.Height != test.want {
			t.Errorf("%s at %d: got %dx%d, want %d", test.profile, test.size, config.Width, config.Height, test.want)
		}
	}
}

func TestHash(t *testing.T) {
	source := Source{Resource: "res:/ui/texture/icons/bpo.png", Profile: Icon, MD5: "45518159d7ae1ce0cb4e80242e22b44b"}
	if source.Hash() != source.Hash() {
		t.Error("hash is not stable")
	}
	lossless := source
	lossless.Profile = Lossless
	if source.Hash() == lossless.Hash() {
		t.Error("profiles share a hash")
	}
	moved := source
	moved.Resource = "res:/elsewhere.png"
	if source.Hash() != moved.Hash() {
		t.Error("hash depends on where the image lives, not only on its content")
	}
}
