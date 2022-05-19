package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"os"
)

type User struct {
	gorm.Model
	Id       uint    `gorm:"unique;not null"`
	Name     string  `gorm:"not null"`
	Location float32 `gorm:"not null"`
	Gender   string
	Email    string
}

type Likes struct {
	gorm.Model
	Id1 uint
	Id2 uint
}

type UserJ struct {
	Id       uint    `json:"id"`
	Name     string  `json:"name"`
	Location float32 `json:"location"`
	Gender   string  `json:"gender"`
	Email    string  `json:"email"`
}

type UsersJ []UserJ

type LikesJ struct {
	Id           uint `json:"id"`
	Who_likes    uint `json:"who_likes"`
	Who_is_liked uint `json:"who_is_liked"`
}

type LikesJs []LikesJ

var gdb *gorm.DB

func parseOne(path string) (id string) {
	var lastSlash = strings.LastIndex(path, "/")
	id = path[(lastSlash + 1):]
	return id
}

func createAndPopulate(db *gorm.DB) {
	// users.json
	userJsonFile, err := os.Open("./users.json")
	if err != nil {
		panic(err)
	}

	defer userJsonFile.Close()

	userData, _ := ioutil.ReadAll(userJsonFile)

	var uresult UsersJ

	err = json.Unmarshal(userData, &uresult)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(uresult); i++ {
		db.Create(&User{Id: uresult[i].Id,
			Name:     uresult[i].Name,
			Location: uresult[i].Location,
			Gender:   uresult[i].Gender,
			Email:    uresult[i].Email})
	}

	// likes.json

	likesJsonFile, err := os.Open("./likes.json")
	if err != nil {
		panic(err)
	}

	defer likesJsonFile.Close()

	likesData, _ := ioutil.ReadAll(likesJsonFile)

	var lresult LikesJs

	err = json.Unmarshal(likesData, &lresult)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(lresult); i++ {
		db.Create(&Likes{
			Id1: lresult[i].Who_likes,
			Id2: lresult[i].Who_is_liked,
		})
	}
}

func contains(list []Likes, l Likes) bool {
	for i := 0; i < len(list); i++ {
		if list[i].Id1 == l.Id1 && list[i].Id2 == l.Id2 {
			return true
		}
	}

	return false
}

func getAllMatches(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Invalid request"}`))

		return
	}

	var l []Likes
	var m []Likes

	gdb.Find(&l)

	for i := 0; i < len(l); i++ {
		for j := 0; j < len(l); j++ {

			if i == j {
				continue
			}

			if contains(m, l[i]) {
				continue
			}

			if l[i].Id1 == l[j].Id2 && l[i].Id2 == l[j].Id1 {
				m = append(m, l[i])
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(m)
}

func getUserName(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Invalid request"}`))

		return
	}

	var q string = parseOne(r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var u []User
	var res []User

	gdb.Find(&u)

	for i := 0; i < len(u); i++ {
		var name string = u[i].Name
		if strings.Contains(name, q) {
			res = append(res, u[i])
		}

	}

	json.NewEncoder(w).Encode(res)
}

func getNearbyUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Invalid request"}`))

		return
	}

	var q1 string = parseOne(r.URL.Path)
	var lastSlash int = strings.LastIndex(r.URL.Path, "/")
	var q2 string = parseOne(r.URL.Path[:lastSlash])

	var name string = q2
	f, _ := strconv.ParseFloat(q1, 32)
	var dist float32 = float32(f)

	var u []User
	var res []User

	gdb.Find(&u)
	var pos float32

	for i := 0; i < len(u); i++ {
		if u[i].Name == name {
			pos = u[i].Location
			break
		}
	}

	for i := 0; i < len(u); i++ {
		if u[i].Name == name {
			continue
		}

		var d = math.Abs(float64(u[i].Location - pos))

		if d < float64(dist) {
			res = append(res, u[i])
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func main() {
	db, err := gorm.Open(sqlite.Open("userData.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// db.AutoMigrate(&User{})
	// db.AutoMigrate(&Likes{})

	// createAndPopulate(db)

	gdb = db

	http.HandleFunc("/matches/", getAllMatches)
	http.HandleFunc("/usernamequery/", getUserName)
	http.HandleFunc("/nearbyusers/", getNearbyUsers)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
