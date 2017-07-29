package labeler

import (
	"bytes"
	"image"
	_ "image/gif" // import gif support
	"image/jpeg"
	_ "image/png" // import png support
	"io"
)

func compressOnce(out io.Writer, in io.Reader, quality int) error {
	option := jpeg.Options{Quality: quality}
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}
	return jpeg.Encode(out, img, &option)
}

// Compress compresses a photo until it is smaller than size
func Compress(out io.Writer, in io.Reader, size int) error {
	var buffer1 bytes.Buffer
	var buffer2 bytes.Buffer

	_, err := buffer2.ReadFrom(in)
	if err != nil {
		return err
	}

	oldImg := &buffer1
	newImg := &buffer2

	quality := 90.0

	for newImg.Len() > size {
		tmp := oldImg
		oldImg = newImg
		newImg = tmp

		err := compressOnce(newImg, oldImg, int(quality))
		if err != nil {
			return err
		}

		oldImg.Reset()
		quality = quality * 0.8
	}

	_, err = newImg.WriteTo(out)
	if err != nil {
		return err
	}

	return nil
}
