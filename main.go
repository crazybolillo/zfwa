package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type tokenInfo struct {
	Active    bool              `json:"active"`
	Audience  []string          `json:"aud,omitempty"`
	ClientID  string            `json:"client_id,omitempty"`
	Expires   int64             `json:"exp,omitempty"`
	Issuer    string            `json:"iss,omitempty"`
	IssuedAt  int64             `json:"iat,omitempty"`
	ID        string            `json:"jti,omitempty"`
	NotBefore int64             `json:"nbf,omitempty"`
	Scope     string            `json:"scope,omitempty"`
	Type      string            `json:"token_type,omitempty"`
	Username  string            `json:"username,omitempty"`
	Metadata  map[string]string `json:"urn:zitadel:iam:user:metadata"`
}

type zitadelOpts struct {
	Host         string `env:"ZITADEL_HOST,required"`
	ProjectID    string `env:"PROJECT_ID,required"`
	ClientID     string `env:"CLIENT_ID,required"`
	ClientSecret string `env:"CLIENT_SECRET,required"`
	VerifyTenant bool   `env:"VERIFY_TENANT" envDefault:"true"`
}

type zitadel struct {
	opts   zitadelOpts
	client *http.Client
}

func (z *zitadel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-Method") == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	info, err := z.introspectToken(token)
	if err != nil {
		slog.Error("Failed to introspect ZITADEL token", slog.String("reason", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = z.validateToken(r, info)
	if err != nil {
		slog.Info("Rejected token", slog.String("reason", err.Error()))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	w.Header().Add("X-Auth-User", info.Username)
}

func (z *zitadel) validateToken(r *http.Request, info tokenInfo) error {
	if !info.Active {
		return errors.New("token is inactive")
	}

	properAudience := false
	for _, aud := range info.Audience {
		if aud == z.opts.ProjectID {
			properAudience = true
			break
		}
	}
	if !properAudience {
		return errors.New("token is missing proper audience")
	}

	if z.opts.VerifyTenant {
		tenant := r.Header.Get("X-Tenant-Id")
		if tenant == "" {
			return errors.New("request is missing X-Tenant-Id header")
		}

		tokenTenant := info.Metadata["tenant"]
		if tokenTenant != tenant {
			return fmt.Errorf("request's tenant (%s) does not match token's (%s)", tenant, tokenTenant)
		}
	}

	return nil
}

func (z *zitadel) introspectToken(token string) (tokenInfo, error) {
	form := url.Values{}
	form.Add("token", token)

	path, err := url.JoinPath(z.opts.Host, "/oauth/v2/introspect")
	if err != nil {
		return tokenInfo{}, err
	}

	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	if err != nil {
		return tokenInfo{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(
		"Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(z.opts.ClientID+":"+z.opts.ClientSecret)),
	)

	res, err := z.client.Do(req)
	if err != nil {
		return tokenInfo{}, err
	}
	defer res.Body.Close()

	var info tokenInfo
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&info)
	if err != nil {
		return tokenInfo{}, err
	}

	return info, nil
}

func Zitadel(opts zitadelOpts) http.Handler {
	z := &zitadel{
		opts:   opts,
		client: &http.Client{Timeout: 10 * time.Second},
	}

	return z
}

func main() {
	listen := os.Getenv("LISTEN_ADDR")
	if listen == "" {
		listen = ":7800"
	}

	var opts zitadelOpts
	err := env.Parse(&opts)
	if err != nil {
		slog.Error("Failed to parse options", slog.String("reason", err.Error()))
		os.Exit(2)
	}

	handler := Zitadel(opts)
	http.Handle("/", handler)

	err = http.ListenAndServe(listen, nil)
	if err != nil {
		slog.Error("Failed to start", slog.String("reason", err.Error()))
		os.Exit(1)
	}
}
