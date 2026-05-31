package slash

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// flexString accepts a JSON value that may arrive as either a string or a
// number, storing it as a string either way.
// Null and empty-string inputs decode to "". Anything other than a JSON
// string or number returns a clear error rather than silently storing the
// literal text.
type flexString string

func (s *flexString) UnmarshalJSON(b []byte) error {
	raw := strings.TrimSpace(string(b))
	if raw == "" || raw == "null" {
		*s = ""
		return nil
	}
	if raw[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		*s = flexString(str)
		return nil
	}
	// json.Number validates the input is a well-formed JSON number and
	// preserves the literal so large values don't lose precision through a
	// float64 round-trip.
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*s = flexString(n.String())
		return nil
	}
	return fmt.Errorf("flexString: expected JSON string or number, got %s", raw)
}

type steamGames struct {
	Applist steamAppList `json:"applist"`
}
type steamApps struct {
	Appid int    `json:"appid"`
	Name  string `json:"name"`
}
type steamAppList struct {
	Apps []steamApps `json:"apps"`
}

type affirmation struct {
	Affirmation string `json:"affirmation"`
}

type advice struct {
	Slip adviceSlip `json:"slip"`
}
type adviceSlip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
}

type doggo struct {
	Weight           animalWeight `json:"weight"`
	Height           doggoHeight  `json:"height"`
	ID               int          `json:"id"`
	Name             string       `json:"name"`
	BredFor          string       `json:"bred_for"`
	BreedGroup       string       `json:"breed_group"`
	LifeSpan         string       `json:"life_span"`
	Temperament      string       `json:"temperament"`
	Origin           string       `json:"origin"`
	ReferenceImageID string       `json:"reference_image_id"`
	Image            doggoImage   `json:"image"`
}
type doggoImage struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}
type animalWeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}
type doggoHeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}

type katz struct {
	Length            string  `json:"length"`
	Origin            string  `json:"origin"`
	ImageLink         string  `json:"image_link"`
	FamilyFriendly    int     `json:"family_friendly"`
	Shedding          int     `json:"shedding"`
	GeneralHealth     int     `json:"general_health"`
	Playfulness       int     `json:"playfulness"`
	Meowing           int     `json:"meowing"`
	ChildrenFriendly  int     `json:"children_friendly"`
	StrangerFriendly  int     `json:"stranger_friendly"`
	Grooming          int     `json:"grooming"`
	Intelligence      int     `json:"intelligence"`
	OtherPetsFriendly int     `json:"other_pets_friendly"`
	MinWeight         float64 `json:"min_weight"`
	MaxWeight         float64 `json:"max_weight"`
	MinLifeExpectancy float64 `json:"min_life_expectancy"`
	MaxLifeExpectancy float64 `json:"max_life_expectancy"`
	Name              string  `json:"name"`
}

type fakePerson struct {
	Results []fakePersonResults `json:"results"`
	Info    fakePersonInfo      `json:"info"`
}
type fakePersonName struct {
	Title string `json:"title"`
	First string `json:"first"`
	Last  string `json:"last"`
}
type fakePersonStreet struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}
type fakePersonCoordinates struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}
type fakePersonTimezone struct {
	Offset      string `json:"offset"`
	Description string `json:"description"`
}
type fakePersonLocation struct {
	Street   fakePersonStreet `json:"street"`
	City     string           `json:"city"`
	State    string           `json:"state"`
	Country  string           `json:"country"`
	Postcode flexString       `json:"postcode"` // string|number — see flexString

	Coordinates fakePersonCoordinates `json:"coordinates"`
	Timezone    fakePersonTimezone    `json:"timezone"`
}
type fakePersonLogin struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Md5      string `json:"md5"`
	Sha1     string `json:"sha1"`
	Sha256   string `json:"sha256"`
}
type fakePersonDob struct {
	Date string `json:"date"`
	Age  int    `json:"age"`
}
type fakePersonRegistered struct {
	Date time.Time `json:"date"`
	Age  int       `json:"age"`
}
type fakePersonID struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type fakePersonPicture struct {
	Large     string `json:"large"`
	Medium    string `json:"medium"`
	Thumbnail string `json:"thumbnail"`
}
type fakePersonResults struct {
	Gender     string               `json:"gender"`
	Name       fakePersonName       `json:"name"`
	Location   fakePersonLocation   `json:"location"`
	Email      string               `json:"email"`
	Login      fakePersonLogin      `json:"login"`
	Dob        fakePersonDob        `json:"dob"`
	Registered fakePersonRegistered `json:"registered"`
	Phone      string               `json:"phone"`
	Cell       string               `json:"cell"`
	ID         fakePersonID         `json:"id"`
	Picture    fakePersonPicture    `json:"picture"`
	Nat        string               `json:"nat"`
}
type fakePersonInfo struct {
	Seed    string `json:"seed"`
	Results int    `json:"results"`
	Page    int    `json:"page"`
	Version string `json:"version"`
}
