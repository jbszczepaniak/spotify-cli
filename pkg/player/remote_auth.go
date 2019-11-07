package player

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/jedruniu/spotify-cli/pkg/web"
)

// StartRemoteAuthentication redirects to spotify's API in order to authenticate user
func StartRemoteAuthentication(authenticator web.SpotifyAuthenticatorInterface, state string) error {
	authUrl := authenticator.AuthURL(state)
	err := openBrowserWith(authUrl)
	if err != nil {
		return fmt.Errorf("could not open browser with url: %s, err: %v", authUrl, err)
	}
	return nil
}

func openBrowserWith(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", "-a", "/Applications/Google Chrome.app", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default:
		return fmt.Errorf("OS: %v is not supported", runtime.GOOS)
	}
}
