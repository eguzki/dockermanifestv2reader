package reader

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
)

func Read(imageURL string) error {
	username := "" // anonymous
	password := "" // anonymous
	if userToken := strings.Split(os.Getenv("USER_TOKEN"), ":"); len(userToken) > 1 {
		username = userToken[0]
		password = userToken[1]
	}

	registryURL, repository, tag := parseImageURL(imageURL)

	hub, err := registry.New(registryURL, username, password)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repository, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// do not follow redirects - this is critical so we can get the registry digest from Location in redirect response
	hub.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := hub.Client.Do(req)
	if err != nil {
		return err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != 302 && resp.StatusCode != 301 {
		return fmt.Errorf("returned statusCode: %v", resp.StatusCode)
	}

	digestURL, err := resp.Location()
	if err != nil {
		return err
	}
	if digestURL == nil {
		return errors.New("digestURL is nil")
	}
	digestURLParts := strings.Split(digestURL.EscapedPath(), "/")
	if len(digestURLParts) < 1 {
		return fmt.Errorf("digestURL can not be splitted: %s", digestURL)
	}

	fmt.Println(digestURLParts[len(digestURLParts)-1])

	return nil
}

//
// Returns (registryURL, repository, tag)
// Example: registry.redhat.io/rhscl/mysql-57-rhel7:5.7
//   Returns registry.redhat.io, rhscl/mysql-57-rhel7, 5.7
func parseImageURL(imageURL string) (string, string, string) {
	urlParts := strings.Split(imageURL, "/")
	if len(urlParts) < 3 {
		panic(fmt.Sprintf("imageURL format not expected: %s", imageURL))
	}

	registryURL := urlParts[0]
	imageContext := urlParts[len(urlParts)-2]
	imageAndTag := urlParts[len(urlParts)-1]
	imageParts := strings.Split(imageAndTag, ":")
	image := imageParts[0]
	var imageTag string
	if len(imageParts) < 2 {
		imageTag = "latest"
	} else {
		imageTag = imageParts[1]
	}
	return fmt.Sprintf("https://%s", registryURL), fmt.Sprintf("%s/%s", imageContext, image), imageTag
}
