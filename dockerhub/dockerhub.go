package dockerhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type TokenResponse struct {
	Token string     `json:"token"`
	Err   []ApiError `json:"errors"`
}

type ManifestResponse struct {
	Config ManifestResponseConfig `json:"config"`
	Err    []ApiError             `json:"errors"`
}

type ManifestResponseConfig struct {
	Digest string `json:"digest"`
}

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ParseContainerIdentifier(identifier string) (string, string) {
	var tag string
	var repository string

	if strings.Contains(identifier, ":") {
		parts := strings.SplitN(identifier, ":", 2)
		identifier = parts[0]
		tag = parts[1]
	} else {
		tag = "latest"
	}

	if strings.Contains(identifier, "/") {
		repository = identifier
	} else {
		repository = "library/" + identifier
	}

	return repository, tag
}

func FormatContainerIdentifier(identifier string) string {
	repo, tag := ParseContainerIdentifier(identifier)

	if strings.HasPrefix(repo, "library/") {
		repo = strings.SplitN(repo, "/", 2)[1]
	}

	return fmt.Sprintf("%s:%s", repo, tag)
}

func GetContainerDigest(repository string, tag string) (string, error) {

	token, err := GetBearerToken(repository)
	if err != nil {
		return "", err
	}
	manifest, err := GetContainerManifest(repository, tag, token)
	if err != nil {
		return "", fmt.Errorf("failed to fetch container manifest json %s: %w", manifest, err)
	}

	var manifestRes ManifestResponse
	err = json.Unmarshal([]byte(manifest), &manifestRes)

	if err != nil {
		return "", fmt.Errorf("failed to parse manifest json %s: %w", manifest, err)
	}

	if len(manifestRes.Err) != 0 {
		errs := make([]string, len(manifestRes.Err))
		for i, err := range manifestRes.Err {
			errs[i] = err.Code
		}
		return "", fmt.Errorf("%s", strings.Join(errs, ","))
	}

	return manifestRes.Config.Digest, nil
}

func GetBearerToken(repository string) (string, error) {
	req, err := http.NewRequest("GET", "https://auth.docker.io/token", nil)

	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	q.Set("scope", fmt.Sprintf("repository:%s:pull", repository))
	q.Set("service", "registry.docker.io")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("failed to send request for bearer token: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read bearer token response body: %w", err)
	}

	var tokenres TokenResponse
	err = json.Unmarshal(body, &tokenres)

	if err != nil {
		return "", fmt.Errorf("failed to parse token json %s: %w", body, err)
	}

	if len(tokenres.Err) != 0 {
		errs := make([]string, len(tokenres.Err))
		for i, err := range tokenres.Err {
			errs[i] = err.Code
		}
		return "", fmt.Errorf("%s", strings.Join(errs, ","))
	}

	return tokenres.Token, nil
}

func GetContainerManifest(repository string, tag string, token string) (string, error) {

	manifesturl := fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", repository, url.PathEscape(tag))
	req, err := http.NewRequest("GET", manifesturl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("failed to send request for manifest: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read manifest response body: %w", err)
	}

	return string(body), nil

}
