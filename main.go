package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/MatthewTully/gator/internal/config"
	"github.com/MatthewTully/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Read()

	s := state{
		config: &cfg,
	}
	db, err := sql.Open("postgres", s.config.DB_url)
	if err != nil {
		fmt.Printf("Error connecting to DB: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

	cmds := commands{
		commands: map[string]func(*state, command) error{},
	}

	cmds.register("login", login_handler)
	cmds.register("register", register_handler)
	cmds.register("reset", reset_handler)
	cmds.register("users", users_handler)
	cmds.register("agg", aggregator_handler)
	cmds.register("addfeed", middlewareLoggedIn(addfeed_handler))
	cmds.register("feeds", listfeed_handler)
	cmds.register("follow", middlewareLoggedIn(follow_handler))
	cmds.register("following", middlewareLoggedIn(following_handler))
	cmds.register("unfollow", middlewareLoggedIn(unfollow_handler))
	cmds.register("browse", middlewareLoggedIn(browse_handler))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}
	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)

}
