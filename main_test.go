package main

import (
	"testing"
	"time"
)

func TestValidatePubKeysWithValidHex(t *testing.T) {
	err := validatePubKeys([]string{"b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe", "b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe"})
	if err != nil {
		t.Errorf("unexpected pubKey validation error: %v", err)
	}
}

func TestValidatePubKeysWithNpub(t *testing.T) {
	err := validatePubKeys([]string{"b8aafafe72f7cd06ae8c337f93147f65fe2d34c0065b52696123982438cf06fe", "npub1hz404lnj7lxsdt5vxdlex9rlvhlz6dxqqed4y6tpywvzgwx0qmlqfpl6sm"})
	if err == nil {
		t.Error("expected npub not supported error")
	}
}

func TestOrderedList(t *testing.T) {
	ol := &orderedList{
		Anchors: []*anchor{
			{Text: "a1", Href: "h1"},
			{Text: "a2", Href: "h2"},
		},
	}

	expected := "<ol>\n" +
		"  <li>\n" +
		"    <a href=\"h1\">a1</a>\n" +
		"    <a href=\"h2\">a2</a>\n" +
		"  </li>\n" +
		"</ol>"

	bytes, err := ol.marshall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actual := string(bytes)
	if actual != expected {
		t.Errorf("expected %v, received %v", expected, actual)
	}
}

func TestSiteMapMarshalling(t *testing.T) {
	sm := &siteMap{XMLNS: sitemapNS}
	t3 := time.Date(2019, 12, 23, 22, 30, 0, 0, time.UTC)
	t2 := t3.Add(-1 * time.Minute)
	t1 := t2.Add(-1 * time.Minute)
	sm.Add("locA", t1)
	sm.Add("locB", t2)
	sm.Add("locA", t3)
	sm.Add("locA", t2)
	expected := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n" +
		"  <url>\n" +
		"    <loc>locA</loc>\n" +
		"    <lastmod>2019-12-23T22:30:00Z</lastmod>\n" +
		"  </url>\n" +
		"  <url>\n" +
		"    <loc>locB</loc>\n" +
		"    <lastmod>2019-12-23T22:29:00Z</lastmod>\n" +
		"  </url>\n" +
		"</urlset>"

	bytes, err := sm.marshall()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actual := string(bytes)
	if actual != expected {
		t.Errorf("expected %v, received %v", expected, actual)
	}
}
