package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

type config struct {
	ArticleLinkPrefix string
	FeedID            string
	FeedPath          string
	FeedTitle         string
	MarkdownFolder    string
	OrderedListPath   string
	PubKeys           []string
	Relay             string
}

type frontMatter struct {
	Title   string `json:"title,omitempty"`
	Updated string `json:"updated"`
}

const kindLongForm = 30023

var pubKeyRE = regexp.MustCompile("^[a-f0-9]{64}$")

func main() {
	conf, err := loadConfig()
	panicIfErr(err)

	ctx := context.Background()

	fd, err := loadOrInitFeed(conf)
	panicIfErr(err)

	// TODO: support multiple, configurable relays - jraedisch
	relay, err := nostr.RelayConnect(ctx, conf.Relay)
	panicIfErr(err)

	panicIfErr(validatePubKeys(conf.PubKeys))

	filters := nostr.Filters{{
		Authors: conf.PubKeys,
		Kinds:   []int{kindLongForm},
	}}
	subscription := relay.Subscribe(ctx, filters)

	go func() {
		<-subscription.EndOfStoredEvents
		subscription.Unsub()
		panicIfErr(relay.Close())
	}()

	for event := range subscription.Events {
		d, ok := extractTagValue(event, "d")
		if !ok {
			log.Fatalf("event missing d tag: %v", event)
		}

		arPath, err := articlePath(conf, d)
		panicIfErr(err)
		summary, _ := extractTagValue(event, "summary")
		title, ok := extractTagValue(event, "title")
		if ok && fd.add(title, arPath, event.CreatedAt, summary, extractTime(event, "published_at")) {
			mdPath, err := markdownPath(conf, d)
			panicIfErr(err)
			panicIfErr(persist(event, mdPath))
			panicIfErr(fd.persist())
			panicIfErr(fd.persistOrderedList())
		}
	}
	log.Println("EOS")
}

func loadConfig() (*config, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("no config provided")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}
	defer f.Close()
	config := &config{}
	return config, json.NewDecoder(f).Decode(config)
}

func validatePubKeys(pubKeys []string) error {
	if len(pubKeys) == 0 {
		return errors.New("pubKey missing")
	}
	for _, pubKey := range pubKeys {
		if strings.HasPrefix(pubKey, "npub") {
			return fmt.Errorf("bech encoded pubKeys are not supported: %s", pubKey)
		}
		if !pubKeyRE.MatchString(pubKey) {
			return fmt.Errorf("invalid pubKey: %s", pubKey)
		}
	}

	return nil
}

func persist(event *nostr.Event, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fM := extractFrontMatter(event)
	bytes, err := json.MarshalIndent(fM, "", "  ")
	if err != nil {
		log.Printf("error marshalling frontmatter from event: %s", event.ID)
	} else {
		_, err = f.Write(append(bytes, []byte("\n")...))
		panicIfErr(err)
	}

	_, err = f.WriteString(event.Content)
	if err != nil {
		return err
	}
	log.Printf("persisted event: %s; article: %s", event.ID, path)
	return nil
}

type orderedList struct {
	XMLName xml.Name  `xml:"ol"`
	Anchors []*anchor `xml:"li>a"`
}

func (ol *orderedList) marshal() ([]byte, error) {
	bytes, err := xml.MarshalIndent(ol, "", "  ")
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type anchor struct {
	Text string `xml:",innerxml"`
	Href string `xml:"href,attr"`
}

type feed struct {
	conf    *config   `xml:"-"`
	XMLName xml.Name  `xml:"feed"`
	XMLNS   string    `xml:"xmlns,attr"`
	ID      string    `xml:"id"`
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`
	Entries []*entry  `xml:"entry"`
}

const atomNS = "http://www.w3.org/2005/Atom"

func loadOrInitFeed(conf *config) (*feed, error) {
	fd := &feed{conf: conf, XMLNS: atomNS, Title: conf.FeedTitle, ID: conf.FeedID}
	f, err := os.Open(conf.FeedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("initialized feed: %s", conf.FeedPath)
			return fd, nil
		}
		return nil, err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	log.Printf("loaded feed: %s", conf.FeedPath)
	return fd, dec.Decode(fd)
}

func (fd *feed) add(title string, href string, date time.Time, summary string, published *time.Time) (changed bool) {
	found := false
	for _, en := range fd.Entries {
		if en.Link.Href == href {
			found = true
			if en.Updated.Before(date) {
				fd.Updated = date
				en.Updated = date
				en.Summary = summary
				en.Title = title
				changed = true
			}
		}
	}
	if found {
		return
	}

	fd.Updated = date
	en := &entry{Title: title, Link: &link{Href: href}, Published: date, Updated: date}
	if published != nil {
		en.Published = *published
	}
	fd.Entries = append(fd.Entries, en)
	changed = true
	return
}

func (fd *feed) persistOrderedList() error {
	ol := &orderedList{Anchors: make([]*anchor, len(fd.Entries))}
	for i, entry := range fd.Entries {
		ol.Anchors[i] = &anchor{Href: entry.Link.Href, Text: entry.Title}
	}
	bytes, err := ol.marshal()
	if err != nil {
		return err
	}
	f, err := os.Create(fd.conf.OrderedListPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(bytes)
	return err
}

var declaration = []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

func (fd *feed) marshal() ([]byte, error) {
	bytes, err := xml.MarshalIndent(fd, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(declaration, bytes...), nil
}

func (fd *feed) persist() error {
	f, err := os.Create(fd.conf.FeedPath)
	if err != nil {
		return err
	}
	defer f.Close()
	bytes, err := fd.marshal()
	if err != nil {
		return err
	}
	_, err = f.Write(bytes)
	return err
}

type entry struct {
	Title     string    `xml:"title"`
	Link      *link     `xml:"link"`
	Published time.Time `xml:"published"`
	Updated   time.Time `xml:"updated"`
	Summary   string    `xml:"summary,omitempty"`
}

type link struct {
	Href string `xml:"href,attr"`
}

func extractFrontMatter(event *nostr.Event) *frontMatter {
	title, _ := extractTagValue(event, "title")
	return &frontMatter{
		Title:   title,
		Updated: event.CreatedAt.Format("2006-01-02"),
	}
}

func extractTagValue(event *nostr.Event, tagName string) (string, bool) {
	tag := event.Tags.GetFirst([]string{tagName})
	if tag == nil {
		return "", false
	}
	value := strings.TrimSpace(tag.Value())
	if value == "" {
		return "", false
	}
	return value, true
}

func extractTime(event *nostr.Event, tagName string) *time.Time {
	tag := event.Tags.GetFirst([]string{tagName})
	if tag == nil {
		return nil
	}
	sec, err := strconv.Atoi(tag.Value())
	if err != nil {
		return nil
	}
	t := time.Unix(int64(sec), 0)
	return &t
}

func articlePath(conf *config, id string) (string, error) {
	return url.JoinPath(conf.ArticleLinkPrefix, id)
}

func markdownPath(conf *config, id string) (string, error) {
	return url.JoinPath(conf.MarkdownFolder, fmt.Sprintf("%s.md", id))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
