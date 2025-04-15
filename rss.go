package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/mmcdole/gofeed"
)

type RssItem struct {
	title       string
	link        url.URL
	content     []string
	categories  []string
	feedName    string
	published   *time.Time
	description *string
}

type FeedDatabaseItem struct {
	FeedLink     *url.URL
	Name         string
	Created      time.Time
	LastModified time.Time
}

// GetRssContent from a channel of RSS urls, parses new RSS items (that are from the lastNHours),
// and sends them to an output channel.
func GetRssContent(feedDatabaseItems chan *FeedDatabaseItem, since time.Time) chan RssItem {
	result := make(chan RssItem)

	go func(feeds chan *FeedDatabaseItem, since time.Time, rssContent chan RssItem) {
		defer close(result)

		for f := range feeds {
			for _, item := range GetRssContentFrom(f, since) {
				rssContent <- *item
			}
		}
	}(feedDatabaseItems, since, result)

	return result
}

// GetRssContentFrom since afterTime from the RSS feed found at url.
func GetRssContentFrom(feed *FeedDatabaseItem, afterTime time.Time) []*RssItem {
	feedUrl := feed.FeedLink
	feedContent, err := gofeed.NewParser().ParseURL(feed.FeedLink.String())
	if err != nil {
		fmt.Println(fmt.Errorf("could not get content from rss url: %s. Error occurred %w", feedUrl, err).Error())
		return []*RssItem{}
	}

	// If Feed entry is new, publish all the content from it.
	publishAllItems := feed.Created.After(afterTime)
	return ExtractRssContentFeed(feedContent, afterTime, publishAllItems, feed.Name)
}

func ExtractRssContentFeed(f *gofeed.Feed, afterTime time.Time, publishAllItems bool, databaseFeedName string) []*RssItem {
    if f == nil || f.Items == nil {
        return nil
    }
    
    result := make([]*RssItem, len(f.Items))
    count := 0
    for _, item := range f.Items {
        // Check if item and PublishedParsed are not nil
        if item != nil && item.PublishedParsed != nil {
            if publishAllItems || item.PublishedParsed.After(afterTime) {
                result[count] = convert(item, databaseFeedName)
                count++
            }
        }
    }
    
    if f.Title != "" {
        fmt.Printf("Feed %s has %d items. %d are eligible to be uploaded\n", f.Title, len(f.Items), count)
    }
    
    if count == 0 {
        return nil
    }
    return result[:count]
}


// convert gofeed.Item into an internal RSSItem model.
func convert(item *gofeed.Item, itemFeedName string) *RssItem {
	link, _ := url.Parse(item.Link)
	return &RssItem{
		title:       item.Title,
		link:        *link,
		content:     []string{item.Content},
		categories:  item.Categories,
		feedName:    itemFeedName,
		published:   item.PublishedParsed,
		description: &item.Description,
	}
}
