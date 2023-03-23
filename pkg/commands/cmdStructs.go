package commands

import (
	"time"
)

type steamGames struct {
	Applist struct {
		Apps []struct {
			Appid int    `json:"appid"`
			Name  string `json:"name"`
		} `json:"apps"`
	} `json:"applist"`
}

type affirmation struct {
	Affirmation string `json:"affirmation"`
}

type kanye struct {
	Quote string `json:"quote"`
}

type advice struct {
	Slip struct {
		ID     int    `json:"id"`
		Advice string `json:"advice"`
	} `json:"slip"`
}

type doggo []struct {
	Breeds []struct {
		Weight struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"weight"`
		Height struct {
			Imperial string `json:"imperial"`
			Metric   string `json:"metric"`
		} `json:"height"`
		ID               int    `json:"id"`
		Name             string `json:"name"`
		BredFor          string `json:"bred_for"`
		BreedGroup       string `json:"breed_group"`
		LifeSpan         string `json:"life_span"`
		Temperament      string `json:"temperament"`
		Origin           string `json:"origin"`
		ReferenceImageID string `json:"reference_image_id"`
	} `json:"breeds"`
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
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
	Data struct {
		Type      []string `json:"Type"`
		Abilities []string `json:"abilities"`
		ASCII     string   `json:"ascii"`
		Height    float64  `json:"height"`
		ID        int      `json:"id"`
		Link      string   `json:"link"`
		Name      string   `json:"name"`
		Weight    int      `json:"weight"`
	} `json:"Data"`
	Answer   string `json:"answer"`
	Question string `json:"question"`
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

//region FakePerson Structs
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
