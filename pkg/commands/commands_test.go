package commands

import (
	"math/rand"
	"testing"
	"time"
)

func TestPickChoices(t *testing.T) {
	/*if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}*/

	/*cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
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
