package commands

import (
	"time"
)

type tuuckCmdInfo struct {
	Name    string
	Desc    string
	Example string
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

type kanye struct {
	Quote string `json:"quote"`
}

type advice struct {
	Slip adviceSlip `json:"slip"`
}
type adviceSlip struct {
	ID     int    `json:"id"`
	Advice string `json:"advice"`
}

type doggo struct {
	Breeds []doggoBreeds `json:"breeds"`
	ID     string        `json:"id"`
	URL    string        `json:"url"`
	Width  int           `json:"width"`
	Height int           `json:"height"`
}
type animalWeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}
type doggoHeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}
type doggoBreeds struct {
	Weight           animalWeight `json:"weight"`
	Height           doggoHeight  `json:"height"`
	ID               int          `json:"id"`
	Name             string       `json:"name"`
	CountryCode      string       `json:"country_code"`
	BredFor          string       `json:"bred_for"`
	BreedGroup       string       `json:"breed_group"`
	LifeSpan         string       `json:"life_span"`
	Temperament      string       `json:"temperament"`
	Origin           string       `json:"origin"`
	ReferenceImageID string       `json:"reference_image_id"`
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

type joke struct {
	ID   string `json:"id"`
	Joke string `json:"joke"`
}

type pickupLine struct {
	Category string `json:"category"`
	Joke     string `json:"joke"`
}

type wyr struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

type gtl struct {
	Answer   string `json:"answer"`
	Brand    string `json:"brand"`
	Clue     string `json:"clue"`
	Easy     bool   `json:"easy"`
	Hint     string `json:"hint"`
	Question string `json:"question"`
	WikiURL  string `json:"wiki_url"`
}

type wtp struct {
	Data     wtpData `json:"Data"`
	Answer   string  `json:"answer"`
	Question string  `json:"question"`
}
type wtpData struct {
	Type      []string `json:"Type"`
	Abilities []string `json:"abilities"`
	ASCII     string   `json:"ascii"`
	Height    float64  `json:"height"`
	ID        int      `json:"id"`
	Link      string   `json:"link"`
	Name      string   `json:"name"`
	Weight    float64  `json:"weight"`
}

type albumPicker struct {
	Genres      string `json:"Genres"`
	Descriptors string `json:"Descriptors"`
	Artist      string `json:"Artist"`
	AlbumName   string `json:"Album_Name"`
	SecGenres   string `json:"Sec_Genres"`
	URL         string `json:"url"`
	//SpotifyAlbumURL string `json:"spotify_album_url"`
}

// region FakePerson Structs
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
	Street      fakePersonStreet      `json:"street"`
	City        string                `json:"city"`
	State       string                `json:"state"`
	Country     string                `json:"country"`
	Postcode    int                   `json:"postcode"`
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

//endregion
