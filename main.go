package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/disintegration/imaging"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	url    = "https://base64.guru/converter/encode/url"
	auth   = "Bearer key-2srPk0uOtQVZdzjAFw3c9rhyaUhzHHbZ4CRsWJHKDvcZnmabUCFNerZrROOft7c9j6yT6UfHUENmxi4cas9gDe8pogoozxA4"
	apiUrl = "https://api.getimg.ai/v1/stable-diffusion-xl/image-to-image"
)

const (
	maxHeight = 1536
	maxWidth  = 1536
)

type Req struct {
	Link  string `json:"link"`
	Style string `json:"style"`
}

type RespAI struct {
	Url string `json:"url,omitempty"`
}

type ErrRepsAi struct {
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})
	e.POST("/api", SupportHandler)
	//e.POST("/api/2", SupportHandler)

	if err := e.Start(":" + os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}

func MainHandler(c echo.Context) error {
	var r Req
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	cli := resty.New()

	// should be added service where we send info + handle link
	resp, err := cli.R().
		SetFormData(map[string]string{
			"form_is_submited": "base64-converter-encode-url",
			"form_action_url":  " /converter/encode/url",
			"url":              r.Link,
			"encode":           "1",
		}).Post(url)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if resp.IsError() {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	return c.JSON(http.StatusOK, string(resp.Body()))
}

func SupportHandler(c echo.Context) error {
	var r Req
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	resp, err := http.Get(r.Link)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	img, err := imaging.Decode(bytes.NewReader(d))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	if width > maxWidth || height > maxHeight {
		newWidth, newHeight := calculateNewDimensions(width, height, maxWidth, maxHeight)
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	encodeImg := base64.StdEncoding.EncodeToString(buf.Bytes())

	body, err := StyleFactory(r.Style, encodeImg)

	cli := resty.New()

	var rr RespAI
	resp2, err := cli.R().
		SetHeaders(map[string]string{"Authorization": auth}).
		SetBody(body).
		SetResult(&rr).
		Post(apiUrl)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if resp2.IsError() {
		return c.JSON(http.StatusUnprocessableEntity, err)
	}

	return c.JSON(http.StatusOK, struct {
		Url string `json:"url,omitempty"`
		//Base64 string `json:"base64,omitempty" json:"base64,omitempty"`
	}{
		Url: rr.Url,
		//Base64: encodeImg,
	})
}

func calculateNewDimensions(width, height, maxWidth, maxHeight int) (newWidth, newHeight int) {
	aspectRatio := float64(width) / float64(height)

	if width > maxWidth {
		newWidth = maxWidth
		newHeight = int(float64(newWidth) / aspectRatio)
	} else {
		newWidth = width
		newHeight = height
	}

	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(newHeight) * aspectRatio)
	}

	return newWidth, newHeight

}

func StyleFactory(style string, encodeImg string) (map[string]any, error) {
	switch style {
	case "Anime":
		return map[string]any{
			"model":           "stable-diffusion-xl-v1-0",
			"prompt":          "anime style picture",
			"negative_prompt": "Disfigured, cartoon, blurry",
			"image":           encodeImg,
			"strength":        0.5,
			"steps":           50,
			"guidance":        7.5,
			"seed":            1,
			"scheduler":       "euler",
			"output_format":   "jpeg",
			"response_format": "url",
		}, nil
	case "GTA":
		return map[string]any{
			"model":           "stable-diffusion-xl-v1-0",
			"prompt":          "GTA character style portrait of a person, bold outlines, realistic textures, slightly exaggerated features, urban background, cinematic composition, trending on artstation",
			"negative_prompt": "blurry, distorted, low resolution, monochrome, disfigured hands, deformed fingers, missing fingers, extra fingers, poorly drawn hands, mutated hands, bad anatomy, malformed limbs",
			"strength":        0.5,
			"image":           encodeImg,
			"steps":           50,
			"guidance":        11,
			"seed":            1,
			"scheduler":       "euler",
			"output_format":   "jpeg",
			"response_format": "url",
		}, nil
	case "Disney":
		return map[string]any{
			"model":           "stable-diffusion-xl-v1-0",
			"prompt":          "Disney/Pixar character style portrait of a person, smooth and rounded features, bright and cheerful colors, expressive eyes, fantasy background, cinematic composition, trending on artstation",
			"negative_prompt": "blurry, distorted, low resolution, sharp edges, disfigured hands, deformed fingers, missing fingers, extra fingers, poorly drawn hands, mutated hands, bad anatomy, malformed limbs, disfigured eyes, uneven eyes, mismatched eyes, distorted pupils, extra eyes, missing eyes, asymmetrical face, disfigured face, deformed face, distorted facial features, wrong proportions, misplaced facial features, poorly drawn face, low quality, bad quality, poorly drawn eyes, blurry face",
			"strength":        0.5,
			"steps":           50,
			"image":           encodeImg,
			"guidance":        8,
			"seed":            1,
			"scheduler":       "euler",
			"output_format":   "jpeg",
			"response_format": "url",
		}, nil
	case "Minecraft":
		return map[string]any{
			"model":           "stable-diffusion-xl-v1-0",
			"prompt":          "Minecraft style portrait of a person, blocky and pixelated design, bright colors, simple shapes, voxel art, cinematic composition, trending on artstation",
			"negative_prompt": "blurry, distorted, low resolution, realistic textures, detailed features, smooth edges, disfigured face, deformed face, malformed facial features, bad anatomy, uneven eyes, mismatched eyes, distorted pupils, asymmetrical face, poorly drawn face, missing eyes, extra eyes, blurry face, facial distortion, wrong proportions, misplaced facial features, poorly drawn eyes, poorly drawn facial features, mutated face, face glitch",
			"strength":        0.5,
			"steps":           50,
			"image":           encodeImg,
			"guidance":        8, "seed": 1,
			"scheduler": "euler", "output_format": "jpeg",
			"response_format": "url",
		}, nil
	case "Rick_Morty":
		return map[string]any{
			"model":           "stable-diffusion-xl-v1-0",
			"prompt":          "Rick and Morty style portrait of a person, simple lines, vibrant colors, exaggerated features, cartoonish design, sci-fi elements, fantasy background, cinematic composition, trending on artstation",
			"negative_prompt": "blurry, distorted, low resolution, realistic textures, detailed features, smooth edges, disfigured face, deformed face, malformed facial features, bad anatomy, uneven eyes, mismatched eyes, distorted pupils, asymmetrical face, poorly drawn face, missing eyes, extra eyes, blurry face, facial distortion, wrong proportions, misplaced facial features, poorly drawn eyes, poorly drawn facial features, mutated face, face glitch",
			"strength":        0.5,
			"steps":           50,
			"guidance":        8,
			"image":           encodeImg,
			"seed":            1,
			"scheduler":       "euler",
			"output_format":   "jpeg",
			"response_format": "url",
		}, nil
	default:
		return map[string]any{}, errors.New("invalid format")
	}
}
