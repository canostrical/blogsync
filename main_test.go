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

	bytes, err := ol.marshal()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actual := string(bytes)
	if actual != expected {
		t.Errorf("expected \n%v, received \n%v", expected, actual)
	}
}

func TestFeedMarshalling(t *testing.T) {
	fd := &feed{XMLNS: atomNS}
	t3 := time.Date(2019, 12, 23, 22, 30, 0, 0, time.UTC)
	t2 := t3.Add(-1 * time.Minute)
	t1 := t2.Add(-1 * time.Minute)
	fd.add("titelA v1", "locA", t1)
	fd.add("titelB", "locB", t2)
	fd.add("titelA v3", "locA", t3)
	fd.add("titelA v2", "locA", t2)
	expected := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<feed xmlns=\"http://www.w3.org/2005/Atom\">\n" +
		"  <entry>\n" +
		"    <title>titelA v3</title>\n" +
		"    <link href=\"locA\"></link>\n" +
		"    <published>2019-12-23T22:28:00Z</published>\n" +
		"    <updated>2019-12-23T22:30:00Z</updated>\n" +
		"  </entry>\n" +
		"  <entry>\n" +
		"    <title>titelB</title>\n" +
		"    <link href=\"locB\"></link>\n" +
		"    <published>2019-12-23T22:29:00Z</published>\n" +
		"    <updated>2019-12-23T22:29:00Z</updated>\n" +
		"  </entry>\n" +
		"</feed>"

	bytes, err := fd.marshal()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	actual := string(bytes)
	if actual != expected {
		t.Errorf("expected \n%v, received \n%v", expected, actual)
	}
}
