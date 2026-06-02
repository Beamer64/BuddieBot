package bot

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestCoinFlip(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	fmt.Println("Flipping...")

	time.Sleep(3 * time.Second)
	fmt.Println("...")

	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		x1 := rand.NewSource(time.Now().UnixNano())
		y1 := rand.New(x1)
		randNum := y1.Intn(200)

		if randNum%2 == 0 {
			fmt.Println("It landed heads")

		} else {
			fmt.Println("It landed tails")
		}
	}
}
