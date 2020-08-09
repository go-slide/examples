package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-slide/slide"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Input struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Age      int    `json:"age" db:"age"`
}

type User struct {
	Email string         `json:"email" db:"email"`
	Input types.JSONText `json:"input" db:"input"`
	Tags  pq.Int64Array  `json:"tags" db:"tags"`
}

func (user *User) AddInput(input Input) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	user.Input = data
	return nil
}

func main() {
	db, err := sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/rust_postgres_server?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	app := slide.InitServer(&slide.Config{})
	app.Get("/users", func(ctx *slide.Ctx) error {
		users := []User{}
		if err := db.Select(&users, "SELECT email, input, tags from json_table"); err != nil {
			return ctx.Send(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"payload": users,
		})
	})
	app.Get("/adduser", func(ctx *slide.Ctx) error {
		user := new(User)
		user.Email = "someemail@gmail.com"
		input := Input{
			Username: "Slide",
			Age:      1,
			Password: "somePassword",
		}
		if err := user.AddInput(input); err != nil {
			return ctx.Send(http.StatusInternalServerError, err.Error())
		}
		_, err := db.NamedExec("INSERT INTO json_table (email, input) VALUES (:email, :input)", user)
		if err != nil {
			return ctx.Send(http.StatusInternalServerError, err.Error())
		}
		return ctx.SendStatusCode(http.StatusOK)
	})
	app.Get("/filter/:age", func(ctx *slide.Ctx) error {
		users := []User{}
		if err := db.Select(&users, "SELECT email, input from json_table"); err != nil {
			return ctx.Send(http.StatusInternalServerError, err.Error())
		}
		age, _ := strconv.Atoi(ctx.GetParam("age"))
		filteredUsers := []string{}
		for _, v := range users {
			input := new(Input)
			err := json.Unmarshal(v.Input, input)
			if err != nil {
				return ctx.Send(http.StatusInternalServerError, err.Error())
			}
			if input.Age == age {
				filteredUsers = append(filteredUsers, input.Username)
			}
		}
		return ctx.JSON(http.StatusOK, map[string][]string{
			"payload": filteredUsers,
		})
	})
	app.Get("/filter/tags/:number", func(ctx *slide.Ctx) error {
		users := new([]User)
		if err := db.Select(users, "SELECT email, input, tags from json_table"); err != nil {
			return ctx.Send(http.StatusInternalServerError, err.Error())
		}
		tagNumber, _ := strconv.ParseInt(ctx.GetParam("number"), 10, 64)
		filteredUsers := new([]string)
		for _, v := range *users {
			for _, t := range v.Tags {
				if t == tagNumber {
					*filteredUsers = append(*filteredUsers, v.Email)
				}
			}
		}
		return ctx.JSON(http.StatusOK, filteredUsers)
	})
	log.Fatal(app.Listen("localhost:3000"))
}
