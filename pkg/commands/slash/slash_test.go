package slash

import (
	"encoding/json"
	"fmt"
	"github.com/beamer64/buddieBot/pkg/config"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestPickChoices(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	/*cfg, err := config_files.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}*/

	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		seed := rand.NewSource(time.Now().UnixNano())
		randSource := rand.New(seed)

		testData := []string{
			"test 1",
			"test 2",
			"test 3",
			"test 4",
			"test 5",
		}
		randomIndex := randSource.Intn(len(testData))
		choice := testData[randomIndex]

		println(choice)
	}
}

func TestCallKatzAPI(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := config.ReadConfig("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		t.Fatal(err)
	}

	// apiURL := "https://api.api-ninjas.com/v1/cats?offset=2"
	apiURL := "https://api.api-ninjas.com/v1/cats?family_friendly=" + strconv.Itoa(rand.Intn(6))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("X-Api-Key", cfg.Configs.Keys.NinjaAPIKey)

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
