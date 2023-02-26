package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

const kindLongForm = 30023

// TODO: support multiple, configurable relays - jraedisch
const relayURL = "wss://nostr-pub.wellorder.net"

// TODO: make folder configurable - jraedisch
const folder = "articles"

var pubKeyRE = regexp.MustCompile("^[a-f0-9]{64}$")

func main() {
	ctx := context.Background()
	relay, err := nostr.RelayConnect(ctx, relayURL)
	panicIfErr(err)

	pubKeys := os.Args[1:]
	panicIfErr(validatePubKeys(pubKeys))

	filters := nostr.Filters{{
		Authors: pubKeys,
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
		panicIfErr(persist(event))
	}
	log.Println("EOS")
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

func persist(event *nostr.Event) error {
	d, err := extractDTagValue(event)
	if err != nil {
		return err
	}

	path, err := filePath(d)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.WriteString(event.Content)
	if err != nil {
		return err
	}
	log.Printf("persisted event: %s; article: %s", event.ID, path)
	return nil
}

func extractDTagValue(event *nostr.Event) (string, error) {
	dTag := event.Tags.GetFirst([]string{"d"})
	if dTag == nil {
		return "", fmt.Errorf("no d tag for event: %s", event.ID)
	}
	dValue := dTag.Value()
	if dValue == "" {
		return "", fmt.Errorf("empty d tag for event: %s", event.ID)
	}
	return dValue, nil
}

func filePath(id string) (string, error) {
	return url.JoinPath(folder, fmt.Sprintf("%s.md", id))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
