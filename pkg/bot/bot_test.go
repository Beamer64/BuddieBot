package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCoinFlip(t *testing.T) {
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

func TestMemberHasRole(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	/*cfg, err := config.ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}*/

	roleName := "test"
	s := discordgo.NewState()

	member, err := s.Member("", "289217573004902400")
	if err != nil {
		t.Fatal(err)
	}

	// memberRoles := make([]string, len(member.Roles))

	for _, role := range member.Roles {
		if role == "@everyone" {
			continue
		}

		if strings.ToLower(role) == roleName {
			fmt.Println("Role not found")
		}
	}
}
