package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"time"

	"github.com/MatthewTully/gator/internal/database"
	"github.com/MatthewTully/gator/internal/outbound"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func cleanRSSItemStrings(item RSSItem) RSSItem {
	item.Title = html.UnescapeString(item.Title)
	item.Link = html.UnescapeString(item.Link)
	item.Description = html.UnescapeString(item.Description)
	item.PubDate = html.UnescapeString(item.PubDate)
	return item
}

func cleanRSSFeedStrings(rss RSSFeed) RSSFeed {
	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Link = html.UnescapeString(rss.Channel.Link)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)

	for i, v := range rss.Channel.Item {
		rss.Channel.Item[i] = cleanRSSItemStrings(v)
	}
	return rss
}

func fetchFeed(ctx context.Context, feedURL string) (RSSFeed, error) {
	res, err := outbound.Get(ctx, feedURL)
	if err != nil {
		return RSSFeed{}, fmt.Errorf("error fetching RSS feed %v, Error: %v", feedURL, err)
	}

	rssFeed := RSSFeed{}
	defer res.Body.Close()

	xmlBody, err := io.ReadAll(res.Body)
	if err != nil {
		return RSSFeed{}, err
	}
	err = xml.Unmarshal(xmlBody, &rssFeed)
	if err != nil {
		return RSSFeed{}, err
	}
	rssFeed = cleanRSSFeedStrings(rssFeed)
	return rssFeed, nil

}

func scrapeFeeds(s *state) {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("An error occurred getting next feed from DB: %v\n", err)
		return
	}
	markFetchArgs := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feedToFetch.ID,
	}

	_, err = s.db.MarkFeedFetched(context.Background(), markFetchArgs)
	if err != nil {
		fmt.Printf("An error occurred updating feed as fetched in DB: %v\n", err)
		return
	}
	rssFeed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		fmt.Printf("An error occurred fetching feed for (%v): %v\n", feedToFetch.Url, err)
		return
	}
	for _, feed := range rssFeed.Channel.Item {
		parsedPubTime, err := time.Parse(time.RFC1123Z, feed.PubDate)
		if err != nil {
			fmt.Printf("Error occurred parsing date: %v\n", err)
			continue
		}

		postArgs := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       sql.NullString{String: feed.Title, Valid: true},
			Url:         feed.Link,
			Description: sql.NullString{String: feed.Description, Valid: true},
			PublishedAt: parsedPubTime,
			FeedID:      feedToFetch.ID,
		}

		_, err = s.db.CreatePost(context.Background(), postArgs)
		if err != nil {
			if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
				continue
			}
			fmt.Printf("An error occurred adding post %v for feed %v to DB: %v\n", feed.Title, feedToFetch.Url, err)
			continue
		}
	}
}
