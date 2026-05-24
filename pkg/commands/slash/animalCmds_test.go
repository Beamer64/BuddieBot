package slash

import "testing"

// TestAnimalsKatz checks the Ninja Cats API (api.api-ninjas.com).
// Requires NinjaAPIKey in config.yaml.
func TestAnimalsKatz(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Keys.NinjaAPIKey, "NinjaAPIKey")
	defer rateLimit("ninja-api")()

	cats, err := callKatzAPI(cfg)
	if err != nil {
		t.Fatalf("callKatzAPI: %v", err)
	}
	if len(cats) == 0 {
		t.Fatal("expected at least one cat in response — API likely changed shape")
	}
	if cats[0].Name == "" {
		t.Fatal("cat name is empty — API likely changed shape")
	}
}

// TestAnimalsDoggo checks TheDogAPI (api.thedogapi.com).
// Requires DoggoAPIkey in config.yaml. Calls a specific breed ID (1)
// to avoid the random-retry loop that the live command uses.
func TestAnimalsDoggo(t *testing.T) {
	requireINTEGRATION(t)
	cfg := loadTestConfig(t)
	requireKey(t, cfg.Keys.DoggoAPIkey, "DoggoAPIkey")
	defer rateLimit("thedogapi")()

	dog, err := callDoggoAPI(cfg, 1)
	if err != nil {
		t.Fatalf("callDoggoAPI: %v", err)
	}
	if dog.Name == "" {
		t.Fatal("dog name is empty — API likely changed shape")
	}
	if dog.Image.URL == "" {
		t.Fatal("dog image URL is empty — API likely changed shape")
	}
}
