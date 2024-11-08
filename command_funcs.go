package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MatthewTully/gator/internal/database"
	"github.com/google/uuid"
)

func login_handler(s *state, cmd command) error {
	args := cmd.args
	if len(args) != 1 {
		return fmt.Errorf("login command expects a single argument, username")
	}
	username := args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("cannot login as %v, has the user been registered?", username)
	}

	err = s.config.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("%v has been set as the current user.\n", s.config.Current_user_name)
	return nil
}

func register_handler(s *state, cmd command) error {
	args := cmd.args
	if len(args) != 1 {
		return fmt.Errorf("register command expects a single argument, username")
	}
	username := args[0]
	db_args := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}

	dbUser, err := s.db.CreateUser(context.Background(), db_args)
	if err != nil {
		return fmt.Errorf("db error occurred creating user %v: %v", username, err)
	}
	s.config.SetUser(username)
	fmt.Printf("New user created successfully: %v\n", dbUser)
	return nil

}

func reset_handler(s *state, cmd command) error {
	return s.db.DeleteAllUsers(context.Background())
}

func users_handler(s *state, cmd command) error {
	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range userList {
		if user == s.config.Current_user_name {
			fmt.Printf(" * %v (current)\n", user)
			continue
		}
		fmt.Printf(" * %v\n", user)
	}
	return nil
}

func aggregator_handler(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected 1 argument for follow. Fetch Duration")
	}
	timeBetweenReq, err := time.ParseDuration(cmd.args[0])

	if err != nil {
		return err
	}
	fmt.Printf("Collecting Feeds every %v\n", timeBetweenReq)
	ticker := time.NewTicker(timeBetweenReq)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func addfeed_handler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("expected 2 arguments for add feed. name, url")
	}
	feedName := cmd.args[0]
	url := cmd.args[1]

	dbArgs := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       url,
		UserID:    user.ID,
	}

	dbRec, err := s.db.CreateFeed(context.Background(), dbArgs)
	if err != nil {
		return fmt.Errorf("an error occurred adding new RSS feed: %v", err)
	}
	fmt.Printf("New feed added: %v\n", dbRec)

	cmd.args = []string{url}
	err = follow_handler(s, cmd, user)
	if err != nil {
		return fmt.Errorf("an error occurred adding new RSS feed: %v", err)
	}

	return nil

}

func listfeed_handler(s *state, cmd command) error {
	feedList, err := s.db.ListAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error listing all feeds: %v", err)
	}

	dbUserCache := map[string]database.User{}
	fmt.Printf("--------------------------\n")
	for _, feed := range feedList {
		dbUser, fetched := dbUserCache[feed.UserID.String()]
		if !fetched {
			dbUser, err = s.db.GetUserById(context.Background(), feed.UserID)
			if err != nil {
				return fmt.Errorf("cannot find matching user for feed item %v. error: %v", feed.Name, err)
			}
			dbUserCache[feed.UserID.String()] = dbUser
		}
		fmt.Printf("Name:\t%v\n", feed.Name)
		fmt.Printf("URL:\t%v\n", feed.Url)
		fmt.Printf("Added:\t%v\n", dbUser.Name)
		fmt.Printf("--------------------------\n")
	}
	return nil
}

func follow_handler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected 1 argument for follow. Url")
	}

	url := cmd.args[0]
	feedDb, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}
	followDBArgs := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feedDb.ID,
		UserID:    user.ID,
	}
	res, err := s.db.CreateFeedFollow(context.Background(), followDBArgs)
	if err != nil {
		return err
	}
	fmt.Printf("Feed Name\t%v\nUser\t%v\n", res.FeedName, res.UserName)
	return nil
}

func following_handler(s *state, cmd command, user database.User) error {
	res, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	if len(res) > 0 {
		fmt.Printf("%v is following:\n", user.Name)
		for _, feed := range res {
			fmt.Printf(" * %v\n", feed.Name)

		}
	} else {
		fmt.Printf("%v is not currently following any feeds.\n", user.Name)
	}

	return nil
}

func unfollow_handler(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected 1 argument for follow. Url")
	}

	url := cmd.args[0]
	feedDb, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	dbArgs := database.DeleteFeedFollowParams{
		FeedID: feedDb.ID,
		UserID: user.ID,
	}

	err = s.db.DeleteFeedFollow(context.Background(), dbArgs)
	if err != nil {
		return err
	}
	return nil
}

func browse_handler(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		i, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("cannot convert %v to an int", cmd.args[0])
		}
		limit = i
	} else if len(cmd.args) > 1 {
		return fmt.Errorf("too many arguments for browse. Expected 1, Limit")
	}

	postsArgs := database.GetPostsForUserParams{
		ID:    user.ID,
		Limit: int32(limit),
	}

	posts, err := s.db.GetPostsForUser(context.Background(), postsArgs)
	if err != nil {
		return err
	}
	fmt.Printf("Posts for user:\t%v\n", user.Name)
	fmt.Printf("--------------------------\n")
	for _, p := range posts {
		fmt.Printf("Title:\t%v\n", p.Title.String)
		fmt.Printf("Desc:\t%v\n", p.Description.String)
		fmt.Printf("Feed:\t%v\n", p.FeedName)
		fmt.Printf("--------------------------\n")

	}
	return nil

}
