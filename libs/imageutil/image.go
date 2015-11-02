package imageutil

import (
	"image"
	"image/draw"
	"io"
	"io/ioutil"

	"github.com/bamiaux/rez"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/sprucehealth/backend/libs/errors"
	"github.com/sprucehealth/backend/libs/golog"
)

// JPEGQuality represents the quality at which to store/resize JPEG images
const JPEGQuality = 95

var resizeFilter = rez.NewLanczosFilter(3)

type flipDirection int

const (
	flipVertical flipDirection = 1 << iota
	flipHorizontal
)

type subImage interface {
	SubImage(r image.Rectangle) image.Image
}

// Options for image related function
type Options struct {
	AllowScaleUp bool
	Crop         bool
}

var defaultResizeOps = &Options{
	AllowScaleUp: false,
	Crop:         true,
}

/*
ResizeImage resizes and optionally crops an image returning the new image.

The resize calculation/algorithm works like this:

If we're only given one dimension (width or height) then calculate the other one
based on the aspect ratio of the original image. In this case no cropping is performed
and all we need to do is resize the image to the calculated size.

If both width and height and provided and cropping is NOT requested then the width and
height service as a bounding box and the image is resized down to fit within that box.

If both width and height are provided and cropping is requested then we'll likely need
to crop unless aspect ratio of the request width and height matches the original image
exactly. If cropping is requires then the original image is first resized to be large
enough (but no larger) in order to fit the request image size, and it's then cropped.
For instance a 640x480 original image being request to resize to 320x320 is first
resized to 426x320 and then cropped from the center to the final size of 320x320.
*/
func ResizeImage(img image.Image, width, height int, ops *Options) (image.Image, error) {
	// Since 0 means unbounded in the chosen dimension then both width and height being 0 is the same as the original image.
	if width <= 0 && height <= 0 {
		return img, nil
	}

	if ops == nil {
		ops = defaultResizeOps
	}

	size := calcSize(img.Bounds().Dx(), img.Bounds().Dy(), width, height, ops)

	// Create a new image that matches the format of the original. The rez
	// package can only resize into the same format as the source.
	var resizedImg image.Image
	rr := image.Rect(0, 0, size.rw, size.rh)
	switch m := img.(type) {
	case *image.YCbCr:
		resizedImg = image.NewYCbCr(rr, m.SubsampleRatio)
	case *image.RGBA:
		resizedImg = image.NewRGBA(rr)
	case *image.NRGBA:
		resizedImg = image.NewNRGBA(rr)
	case *image.Gray:
		resizedImg = image.NewGray(rr)
	default:
		// Convert to RGBA
		newImg := image.NewRGBA(img.Bounds())
		draw.Draw(newImg, newImg.Bounds(), img, img.Bounds().Min, draw.Src)
		img = newImg
		resizedImg = image.NewRGBA(rr)
	}

	if err := rez.Convert(resizedImg, img, resizeFilter); err != nil {
		return nil, errors.Trace(err)
	}

	if size.crop {
		// It's safe to assume that resizeImg implements the SubImage interface
		// because above we matched on specific image types that all have the
		// SubImage method.
		x0 := (size.rw - size.w) / 2
		y0 := (size.rh - size.h) / 2
		resizedImg = resizedImg.(subImage).SubImage(image.Rect(x0, y0, x0+size.w, y0+size.h))
	}

	return resizedImg, nil
}

type sizeOp struct {
	w, h   int // final size of image
	rw, rh int // resized size of image before crop
	crop   bool
}

// calcSize calculates the resultant size based on the original image size and requested criteria.
// It's broken out to ease testing.
func calcSize(imgWidth, imgHeight, reqWidth, reqHeight int, ops *Options) sizeOp {
	// Never return a larger image than the original.
	if !ops.AllowScaleUp {
		if reqWidth > imgWidth {
			reqWidth = imgWidth
		}
		if reqHeight > imgHeight {
			reqHeight = imgHeight
		}
	}

	// If only given one dimension then calculate the other dimension based on the aspect ratio.
	var cropOrBound bool
	if reqWidth <= 0 {
		reqWidth = imgWidth * reqHeight / imgHeight
	} else if reqHeight <= 0 {
		reqHeight = imgHeight * reqWidth / imgWidth
	} else {
		cropOrBound = true
	}

	resizeWidth := reqWidth
	resizeHeight := reqHeight
	crop := false
	if cropOrBound {
		imgRatio := float64(imgWidth) / float64(imgHeight)
		reqRatio := float64(reqWidth) / float64(reqHeight)
		if ops.Crop {
			if imgRatio > reqRatio {
				crop = true
				resizeWidth = imgWidth * reqHeight / imgHeight
			} else if imgRatio < reqRatio {
				crop = true
				resizeHeight = imgHeight * reqWidth / imgWidth
			}
		} else {
			if imgRatio == reqRatio {
				resizeWidth = reqWidth
				resizeHeight = reqHeight
			} else if imgRatio > reqRatio {
				resizeHeight = imgHeight * reqWidth / imgWidth
			} else {
				resizeWidth = imgWidth * reqHeight / imgHeight
			}
			reqWidth = resizeWidth
			reqHeight = resizeHeight
		}
	}

	return sizeOp{
		w:    reqWidth,
		h:    reqHeight,
		rw:   resizeWidth,
		rh:   resizeHeight,
		crop: crop,
	}
}

// ResizeImageFromReader takes the provided reader and resizes the image. It rotates the image
// based on EXIF orientation attributes.
func ResizeImageFromReader(r io.Reader, width, height int, ops *Options) (image.Image, string, error) {
	// This doesn't use DecodeAndOrient since it can more efficiently orient after the resize.

	img, imf, ex, err := DecodeImageAndExif(r)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	var orient int
	if ex != nil {
		if tag, err := ex.Get(exif.Orientation); err == nil {
			orient, err = tag.Int(0)
			if err != nil {
				orient = 0
			}
		}
	}

	// Reoriented desired dimensions since we're fixing orientation after resize
	if orient >= 5 && orient <= 8 {
		width, height = height, width
	}

	imgOut, err := ResizeImage(img, width, height, ops)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	if orient != 0 {
		return orientImage(imgOut, orient), imf, nil
	}
	return imgOut, imf, nil
}

// DecodeAndOrient decoders an image from the reader and fixes the orientation based
// on EXIF attributes. It returns the oriented image and the format.
func DecodeAndOrient(r io.Reader) (image.Image, string, error) {
	img, imf, ex, err := DecodeImageAndExif(r)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	var orient int
	if ex != nil {
		if tag, err := ex.Get(exif.Orientation); err == nil {
			orient, err = tag.Int(0)
			if err != nil {
				orient = 0
			}
		}
	}

	if orient != 0 {
		return orientImage(img, orient), imf, nil
	}

	return img, imf, nil
}

// orientImage returns the oriented version of the image. The oritentation is
// the exif value for the orientation.
func orientImage(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return flip(img, flipHorizontal)
	case 3:
		return rotate(img, 180)
	case 4:
		return flip(img, flipVertical)
	case 5:
		return flip(rotate(img, -90), flipHorizontal)
	case 6:
		return rotate(img, -90)
	case 7:
		return flip(rotate(img, 90), flipHorizontal)
	case 8:
		return rotate(img, 90)
	}
	return img
}

func rotate(im image.Image, angle int) image.Image {
	var rotated *image.NRGBA
	// trigonometric (i.e counter clock-wise)
	switch angle {
	case 90:
		newH, newW := im.Bounds().Dx(), im.Bounds().Dy()
		rotated = image.NewNRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				rotated.Set(rotated.Bounds().Min.X+x, rotated.Bounds().Min.Y+y, im.At(im.Bounds().Min.X+newH-1-y, im.Bounds().Min.Y+x))
			}
		}
	case -90:
		newH, newW := im.Bounds().Dx(), im.Bounds().Dy()
		rotated = image.NewNRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				rotated.Set(rotated.Bounds().Min.X+x, rotated.Bounds().Min.Y+y, im.At(im.Bounds().Min.X+y, im.Bounds().Min.Y+newW-1-x))
			}
		}
	case 180, -180:
		newW, newH := im.Bounds().Dx(), im.Bounds().Dy()
		rotated = image.NewNRGBA(image.Rect(0, 0, newW, newH))
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				rotated.Set(rotated.Bounds().Min.X+x, rotated.Bounds().Min.Y+y, im.At(im.Bounds().Min.X+newW-1-x, im.Bounds().Min.Y+newH-1-y))
			}
		}
	default:
		return im
	}
	return rotated
}

func flip(im image.Image, dir flipDirection) image.Image {
	if dir == 0 {
		return im
	}
	ycbcr := false
	var nrgba image.Image
	dx, dy := im.Bounds().Dx(), im.Bounds().Dy()
	di, ok := im.(draw.Image)
	if !ok {
		if _, ok := im.(*image.YCbCr); !ok {
			golog.Errorf("failed to flip iamge: input does not satisfy draw.Image")
			return im
		}
		// because YCbCr does not implement Set, we replace it with a new NRGBA
		ycbcr = true
		nrgba = image.NewNRGBA(image.Rect(0, 0, dx, dy))
		di, ok = nrgba.(draw.Image)
		if !ok {
			golog.Errorf("failed to flip image: could not cast an NRGBA to a draw.Image")
			return im
		}
	}
	if dir&flipHorizontal != 0 {
		for y := 0; y < dy; y++ {
			for x := 0; x < dx/2; x++ {
				old := im.At(im.Bounds().Min.X+x, im.Bounds().Min.Y+y)
				di.Set(di.Bounds().Min.X+x, di.Bounds().Min.Y+y, im.At(im.Bounds().Min.X+dx-1-x, im.Bounds().Min.Y+y))
				di.Set(di.Bounds().Min.X+dx-1-x, di.Bounds().Min.Y+y, old)
			}
		}
	}
	if dir&flipVertical != 0 {
		for y := 0; y < dy/2; y++ {
			for x := 0; x < dx; x++ {
				old := im.At(im.Bounds().Min.X+x, im.Bounds().Min.Y+y)
				di.Set(di.Bounds().Min.X+x, di.Bounds().Min.Y+y, im.At(im.Bounds().Min.X+x, im.Bounds().Min.Y+dy-1-y))
				di.Set(di.Bounds().Min.X+x, di.Bounds().Min.Y+dy-1-y, old)
			}
		}
	}
	if ycbcr {
		return nrgba
	}
	return im
}

// DecodeImageAndExif decodes an image and its EXIF attributes from the provided reader.
func DecodeImageAndExif(r io.Reader) (image.Image, string, *exif.Exif, error) {
	// Use a pipe to avoid having to buffer the image data in memory because
	// both the image decoder and exif decoder need a Reader.
	pr, pw := io.Pipe()
	defer pw.Close()

	exifCh := make(chan *exif.Exif, 1)
	go func(r io.Reader, exifCh chan *exif.Exif) {
		var ex *exif.Exif
		defer func() {
			exifCh <- ex
			// The exif decoder isn't guaranteed to consume the entire file
			// so make sure to drain off all data from the pipe or it'll clog.
			io.Copy(ioutil.Discard, r)
		}()
		var err error
		ex, err = exif.Decode(r)
		if err != nil {
			ex = nil
		}
	}(pr, exifCh)

	ir := io.TeeReader(r, pw)
	img, imf, err := image.Decode(ir)
	if err != nil {
		return nil, "", nil, err
	}
	// Make sure to consume all the data in case the image was just at the beginning and the
	// exif data is at the end. This should in general be a noop, and I'm not sure exactly
	// where the exif data is most commonly.
	io.Copy(ioutil.Discard, ir)
	pw.Close()

	exifData := <-exifCh

	return img, imf, exifData, nil
}

// DecodeImageConfigAndExif decodes the image config, format, and EXIF from the reader.
// The dimensions in the returned config are updated to match the orientation.
func DecodeImageConfigAndExif(r io.ReadSeeker) (image.Config, string, *exif.Exif, error) {
	ex, err := exif.Decode(r)
	if err != nil {
		ex = nil
	}
	r.Seek(0, 0)
	cnf, imf, err := image.DecodeConfig(r)
	if err != nil {
		return cnf, "", nil, errors.Trace(err)
	}
	if ex != nil {
		var orient int
		if tag, err := ex.Get(exif.Orientation); err == nil {
			orient, err = tag.Int(0)
			if err != nil {
				orient = 0
			}
		}
		if orient >= 5 && orient <= 8 {
			cnf.Width, cnf.Height = cnf.Height, cnf.Width
		}
	}
	return cnf, imf, ex, nil
}