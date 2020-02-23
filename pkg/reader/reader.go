package reader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

	hub, err := registry.NewInsecure(registryURL, username, password)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v2/%s/manifests/%s", registryURL, repository, tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json")

	resp, err := hub.Client.Do(req)
	if err != nil {
		return err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//fmt.Println(string(body))

	//for k, v := range resp.Header {
	//	fmt.Print(k)
	//	fmt.Print(" : ")
	//	fmt.Println(v)
	//}

	if len(resp.Header["Content-Type"]) < 1 {
		return errors.New("Content-type header not found")
	}

	switch contentTypeHeader := resp.Header["Content-Type"][0]; contentTypeHeader {
	case "application/vnd.docker.distribution.manifest.list.v2+json":
		digest, err := parseFatManifest(body)
		if err != nil {
			return err
		}
		fmt.Println(digest)
	case "application/vnd.docker.distribution.manifest.v2+json":
		digest, err := parseManifestV2Headers(resp.Header)
		if err != nil {
			return err
		}
		fmt.Println(digest)
	default:
		return fmt.Errorf("Content-type not known: %s", contentTypeHeader)
	}

	return nil
}

func parseFatManifest(body []byte) (string, error) {
	//{
	//	"manifests": [
	//	{
	//		"digest": "sha256:de3ab628b403dc5eed986a7f392c34687bddafee7bdfccfd65cecf137ade3dfd",
	//		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	//		"platform": {
	//			"architecture": "amd64",
	//			"os": "linux"
	//		},
	//		"size": 1160
	//	},
	//	{
	//		"digest": "sha256:53215627779289c622f99744e546ec8a73b334bf63a19aa1a1db0f36a2e87d13",
	//		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	//		"platform": {
	//			"architecture": "ppc64le",
	//			"os": "linux"
	//		},
	//		"size": 1160
	//	},
	//	{
	//		"digest": "sha256:b14915ab40ba42f7cc353da51f49490675d6920e0c1e86d6ec99bf5873ca2df3",
	//		"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	//		"platform": {
	//			"architecture": "s390x",
	//			"os": "linux"
	//		},
	//		"size": 1160
	//	}
	//	],
	//	"mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
	//	"schemaVersion": 2
	//}
	//

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	manifests := result["manifests"].([]interface{})

	digest := ""
	for _, item := range manifests {
		manifest := item.(map[string]interface{})
		platform := manifest["platform"].(map[string]interface{})
		architecture := platform["architecture"].(string)
		if architecture == "amd64" {
			digest = manifest["digest"].(string)
		}
	}
	if digest == "" {
		return "", errors.New("Digest not found for amd64 architecture")
	}

	return digest, nil
}

func parseManifestV2Headers(headers http.Header) (string, error) {
	if len(headers["Docker-Content-Digest"]) < 1 {
		return "", errors.New("Expected Docker-Content-Digest header not found")
	}

	return headers["Docker-Content-Digest"][0], nil
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
