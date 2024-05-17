package services

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/disintegration/imageorient"
)

type Image struct {
	ServicesBase
}

func (sf *Image) GetImageSize(url string) (int, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	img, _, err := imageorient.Decode(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	return img.Bounds().Dx(), img.Bounds().Dy(), nil
}
