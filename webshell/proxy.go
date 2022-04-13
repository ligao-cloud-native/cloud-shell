package webshell

import (
	"crypto/tls"
	"github.com/ligao-cloud-native/cloud-shell/pkg/reverseproxy"
	"github.com/ligao-cloud-native/cloud-shell/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func ServeReverseProxy(target string, w http.ResponseWriter, r *http.Request) (err error) {
	if strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
		err = ServeWSReversedProxy(target, w, r)
	} else {
		err = ServeHttpReverseProxy(target, w, r)
	}

	return err
}

// ServeHttpReverseProxy takes target host and creates a reverse proxy
func ServeHttpReverseProxy(targetHost string, w http.ResponseWriter, r *http.Request) error {
	oUrl, err := url.Parse(targetHost)
	if err != nil {
		return err
	}
	httpProxy := httputil.NewSingleHostReverseProxy(oUrl)

	httpProxy.Director = func(r *http.Request) {
		r.Header.Add("X-Forwarder-Host", r.Host)
		r.Header.Add("X-Origin-Host", oUrl.Host)
		r.URL.Scheme = oUrl.Scheme
		r.URL.Host = oUrl.Host
		r.Host = oUrl.Host
		r.URL.Path = utils.SingleJoiningSlash(oUrl.Path, r.URL.Path)
		if oUrl.RawQuery == "" || r.URL.RawQuery == "" {
			r.URL.RawQuery = oUrl.RawQuery + r.URL.RawQuery
		} else {
			r.URL.RawQuery = oUrl.RawQuery + "&" + r.URL.RawQuery
		}
	}

	if oUrl.Scheme == "https" {
		httpProxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				ServerName:         oUrl.Host,
			},
		}
	}

	httpProxy.ServeHTTP(w, r)
	return nil
}

func ServeWSReversedProxy(targetHost string, w http.ResponseWriter, r *http.Request) error {
	oUrl, err := url.Parse(targetHost)
	if err != nil {
		return err
	}
	wsScheme := "ws" + strings.TrimPrefix(oUrl.Scheme, "http")
	wsUrl := &url.URL{Scheme: wsScheme, Host: oUrl.Host}

	wsRProxy := reverseproxy.NewSingleHostWsReverseProxy(wsUrl)
	if wsScheme == "wss" {
		wsRProxy.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         oUrl.Host,
		}
	}

	wsRProxy.ServeHTTP(w, r)
	return nil
}
