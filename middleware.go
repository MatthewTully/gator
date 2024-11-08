package main

import (
	"context"
	"fmt"

	"github.com/MatthewTully/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		current_user := s.config.Current_user_name

		dbUser, err := s.db.GetUser(context.Background(), current_user)
		if err != nil {
			return fmt.Errorf("error fetching users details")
		}
		return handler(s, cmd, dbUser)
	}
}
