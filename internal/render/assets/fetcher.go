package assets

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/arayofcode/footprint/internal/domain"
)

// FetchAssets retrieves all necessary images (avatars) and returns them as a map of Data URLs.
func FetchAssets(user domain.User, repos []domain.RepoContribution, projects []domain.OwnedProject) map[domain.AssetKey]string {
	assets := make(map[domain.AssetKey]string)

	// Fetch User Avatar
	if user.AvatarURL != "" {
		key := domain.UserAvatarKey(user.Username)
		assets[key] = fetchAsDataURL(user.AvatarURL)
	}

	// Fetch Repo Avatars
	for _, r := range repos {
		if r.AvatarURL != "" {
			key := domain.RepoAvatarKey(r.Repo)
			if _, ok := assets[key]; !ok {
				assets[key] = fetchAsDataURL(r.AvatarURL)
			}
		}
	}

	// Fetch Project Avatars
	for _, p := range projects {
		if p.AvatarURL != "" {
			key := domain.RepoAvatarKey(p.Repo)
			if _, ok := assets[key]; !ok {
				assets[key] = fetchAsDataURL(p.AvatarURL)
			}
		}
	}

	return assets
}

func fetchAsDataURL(url string) string {
	if url == "" {
		return ""
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return html.EscapeString(url)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return html.EscapeString(url)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return html.EscapeString(url)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
}
