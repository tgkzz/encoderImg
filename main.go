package main

import (
	"encoding/base64"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	url = "https://base64.guru/converter/encode/url"
)

type Req struct {
	Link string `json:"link"`
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

	encodeImg := base64.StdEncoding.EncodeToString(d)

	return c.JSON(http.StatusOK, encodeImg)
}
