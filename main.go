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
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

type config struct {
	FeedPath       string
	MarkdownFolder string
	PubKeys        []string
	Relay          string
}

type frontMatter struct {
	Title   string `json:"title"`
	Updated string `json:"updated"`
}

const kindLongForm = 30023

var pubKeyRE = regexp.MustCompile("^[a-f0-9]{64}$")

func main() {
	conf, err := loadConfig()
	panicIfErr(err)

	ctx := context.Background()

	sm, err := loadOrInitFeed(conf.FeedPath)
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

	// TODO: keep listening - jraedisch
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

		mdPath, err := markdownPath(conf, d)
		panicIfErr(err)
		if sm.Add(mdPath, event.CreatedAt) {
			panicIfErr(persist(event, mdPath))
			panicIfErr(sm.persist())
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

	fM, ok := extractFrontMatter(event)
	if ok {
		bytes, err := json.MarshalIndent(fM, "", "  ")
		if err != nil {
			log.Printf("error marshalling frontmatter from event: %s", event.ID)
		} else {
			_, err = f.Write(append(bytes, []byte("\n")...))
			panicIfErr(err)
		}
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

func (ol *orderedList) marshall() ([]byte, error) {
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
	XMLName xml.Name `xml:"feed"`
	XMLNS   string   `xml:"xmlns,attr"`
	path    string   `xml:"-"`
	Entries []*entry `xml:"entry"`
}

const atomNS = "http://www.w3.org/2005/Atom"

func loadOrInitFeed(path string) (*feed, error) {
	fd := &feed{path: path, XMLNS: atomNS}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("initialized feed: %s", path)
			return fd, nil
		}
		return nil, err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	log.Printf("loaded feed: %s", path)
	return fd, dec.Decode(fd)
}

func (fd *feed) Add(href string, updated time.Time) (changed bool) {
	found := false
	for _, en := range fd.Entries {
		if en.Link.Href == href {
			found = true
			if en.Updated.Before(updated) {
				en.Updated = updated
				changed = true
			}
		}
	}
	if found {
		return
	}

	fd.Entries = append(fd.Entries, &entry{Link: &link{Href: href}, Updated: updated})
	changed = true
	return
}

var declaration = []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")

func (fd *feed) marshall() ([]byte, error) {
	bytes, err := xml.MarshalIndent(fd, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(declaration, bytes...), nil
}

func (fd *feed) persist() error {
	f, err := os.Create(fd.path)
	if err != nil {
		return err
	}
	defer f.Close()
	bytes, err := fd.marshall()
	if err != nil {
		return err
	}
	_, err = f.Write(bytes)
	return err
}

type entry struct {
	Link    *link     `xml:"link"`
	Updated time.Time `xml:"updated"`
}

type link struct {
	Href string `xml:"href,attr"`
}

func extractFrontMatter(event *nostr.Event) (*frontMatter, bool) {
	title, ok := extractTagValue(event, "title")
	return &frontMatter{Title: title, Updated: event.CreatedAt.Format("2006-01-02")}, ok
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

func markdownPath(conf *config, id string) (string, error) {
	return url.JoinPath(conf.MarkdownFolder, fmt.Sprintf("%s.md", id))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
