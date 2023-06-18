package bot

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

// pixels is a struct that stores all pixels colors.
type pixels struct {
	r []int
	g []int
	b []int
}

// handleMessage handles messages sent by user.
func (b *Bot) handleMessage(msg *tgbotapi.Message) error {
	if msg.Photo != nil || msg.Document != nil {
		var fileID string
		if msg.Photo != nil {
			fileID = msg.Photo[0].FileID
		}
		if msg.Document != nil {
			fileID = msg.Document.FileID
		}
		averageColor, err := b.getAverageColor(fileID, msg.From.ID)
		if err != nil {
			return errors.Wrap(err, "getting average color of source image")
		}
		preview, err := createColorPreview(averageColor, msg)
		if err != nil {
			return errors.Wrap(err, "creating color preview")
		}
		averageColorHex := fmt.Sprintf("#%02x%02x%02x", averageColor[0], averageColor[1], averageColor[2])
		preview.Caption = averageColorHex
		_, err = b.bot.Send(preview)
		if err != nil {
			return err
		}
	}

	return nil
}

// getAverageColor gets average color of image sent by user.
func (b *Bot) getAverageColor(imageID string, userID int64) (avgColor []int, err error) {
	downloadURL, err := getURL(b, imageID)
	if err != nil {
		return []int{}, errors.Wrap(err, "getting source image URL")
	}
	imgBytes, err := downloadFileBytes(downloadURL)
	if err != nil {
		return []int{}, errors.Wrap(err, "downloading source image")
	}
	img, _, err := image.Decode(imgBytes)
	if err != nil {
		return []int{}, errors.Wrap(err, "decoding source image")
	}
	bounds := img.Bounds()
	imgSize := (bounds.Max.Y - bounds.Min.Y) * (bounds.Max.X - bounds.Min.X)
	var pix pixels
	pix.r = make([]int, 0, imgSize)
	pix.g = make([]int, 0, imgSize)
	pix.b = make([]int, 0, imgSize)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixR, pixG, pixB, _ := img.At(x, y).RGBA()
			pix.r = append(pix.r, int(pixR/257))
			pix.g = append(pix.g, int(pixG/257))
			pix.b = append(pix.b, int(pixB/257))
		}
	}
	avgColor = []int{avg(pix.r), avg(pix.g), avg(pix.b)}
	return avgColor, nil
}

// getURL creates and returns download URL for file sent by user.
func getURL(b *Bot, fileID string) (URL string, err error) {
	var fileConfig tgbotapi.FileConfig
	fileConfig.FileID = fileID
	file, err := b.bot.GetFile(fileConfig)
	if err != nil {
		return "", err
	}
	URL = file.Link(b.bot.Token)
	return URL, nil
}

// downloadFileBytes downloads and returns file bytes from given URL.
func downloadFileBytes(URL string) (*bytes.Buffer, error) {
	// Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return nil, errors.Wrap(err, "getting URL response")
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Received non 200 response code")
	}

	// Write the bytes from img
	w := new(bytes.Buffer)
	_, err = io.Copy(w, response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "building image bytes")
	}

	return w, nil
}

// avg counts an average from given integer slice.
func avg(numbers []int) int {
	sum := 0
	for _, v := range numbers {
		sum += v
	}
	avg := sum / len(numbers)
	return avg
}

// createColorPreview Creates Color preview.
func createColorPreview(avgColor []int, msg *tgbotapi.Message) (tgbotapi.PhotoConfig, error) {
	// Create image template
	width := 500
	height := 500

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Set color for each pixel
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{uint8(avgColor[0]), uint8(avgColor[1]), uint8(avgColor[2]), 255})
		}
	}

	// Encode as jpeg bytes
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return tgbotapi.PhotoConfig{}, errors.Wrap(err, "encoding color preview image")
	}
	photo := tgbotapi.NewPhoto(msg.From.ID, tgbotapi.FileBytes{Name: "color", Bytes: buf.Bytes()})
	return photo, nil
}
