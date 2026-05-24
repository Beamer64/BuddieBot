package slash

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/config"
)

func TestCallKatzAPI(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		t.Fatal(err)
	}

	// apiURL := "https://api.api-ninjas.com/v1/cats?offset=2"
	apiURL := "https://api.api-ninjas.com/v1/cats?family_friendly=" + strconv.Itoa(rand.Intn(6))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("X-Api-Key", cfg.Keys.NinjaAPIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	var katzObj katz

	err = json.NewDecoder(res.Body).Decode(&katzObj)
	if err != nil {
		t.Fatal(err)
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	fmt.Println(katzObj)
}
