package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

type config struct {
	Folder  string
	PubKeys []string
	Relay   string
}

type frontMatter struct {
	Title string `json:"title"`
}

const kindLongForm = 30023

var pubKeyRE = regexp.MustCompile("^[a-f0-9]{64}$")

func main() {
	conf, err := loadConfig()
	panicIfErr(err)

	ctx := context.Background()

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
		panicIfErr(persist(event, conf.Folder))
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

func persist(event *nostr.Event, folder string) error {
	d, ok := extractTagValue(event, "d")
	if !ok {
		return errors.New("missing d tag")
	}

	path, err := filePath(folder, d)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

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

func extractFrontMatter(event *nostr.Event) (*frontMatter, bool) {
	title, ok := extractTagValue(event, "title")
	return &frontMatter{Title: title}, ok
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

func filePath(folder string, id string) (string, error) {
	return url.JoinPath(folder, fmt.Sprintf("%s.md", id))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
