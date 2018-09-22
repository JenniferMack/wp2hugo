package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type RSS struct {
	Channel struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

type Item struct {
	Title    string     `xml:"title"`
	Link     string     `xml:"link"`
	Pubdate  string     `xml:"pubDate"`
	Body     Body       `xml:"bodyText"`
	PostName string     `xml:"post_name"`
	Category []Category `xml:"category"`
}
type Body struct {
	Html string `xml:",cdata"`
}
type Category struct {
	XMLName xml.Name `xml:"category"`
	Type    string   `xml:"domain,attr"`
	Html    string   `xml:",cdata"`
}

func (i *Item) DateTime() time.Time {
	time, _ := time.Parse(time.RFC1123Z, i.Pubdate)
	return time
}

func main() {
	export, err := os.Open("export.xml")
	if err != nil {
		log.Fatal(err)
	}

	rss := RSS{}
	err = xml.NewDecoder(export).Decode(&rss)
	if err != nil {
		log.Fatal(err)
	}

	for _, i := range rss.Channel.Items {
		if len(i.Body.Html) < 1 {
			continue
		}
		fp, err := os.Create("content/blog/" + i.PostName + ".md")
		if err != nil {
			log.Fatal(err)
		}
		_, err = fp.WriteString(fmt.Sprintf(
			`---
date: %s
slug: %s
title: "%s"
%s
---
%s
`, i.DateTime().Format(time.RFC3339), i.PostName, i.Title, formatTags(i.Category), fixBody(i.Body.Html)))
	}
}

func fixBody(s string) string {
	re := regexp.MustCompile(`<!--more-->`)
	s = re.ReplaceAllString(s, ``)

	re = regexp.MustCompile(`(?:<a .*?>)(<img.*?>)(?:</a>)`)
	s = re.ReplaceAllString(s, `<figure class="figstyle">$1<figcaption class="figcapstyle">$2</figcaption></figure>`+"\n\n")

	re = regexp.MustCompile(`\[caption.*?\].*(<img.*>)(.*)\[/caption\]`)
	s = re.ReplaceAllString(s, `<figure class="figstyle">$1<figcaption class="figcapstyle">$2</figcaption></figure>`+"\n\n")

	re = regexp.MustCompile(`<img.*src="(.*?)" .*?/?>`)
	s = re.ReplaceAllString(s, `<img src="$1" />`)

	re = regexp.MustCompile(`"https?://jennifermackdotnet.files.wordpress.com/\d{4}/\d{2}/(.+\..{3,4}).*"`)
	s = re.ReplaceAllString(s, `"/images/$1"`)

	re = regexp.MustCompile(`\[display-posts.*?\]`)
	s = re.ReplaceAllString(s, ``)

	re = regexp.MustCompile(`More MVW travel reports:`)
	s = re.ReplaceAllString(s, `Use the [MVW Travel tag](/tag/mvw-travel) to see all the posts in this series.`)
	return s
}

func formatTags(c []Category) string {
	var cats, tags []string
	for _, n := range c {
		if n.Type == "category" {
			cats = append(cats, n.Html)
		}
		if n.Type == "post_tag" {
			tags = append(tags, n.Html)
		}
	}
	if len(cats) == 0 {
		cats = append(cats, "blog")
	}
	return fmt.Sprintf(`
tag:
  - %s
category:
  - %s`, strings.Join(tags, "\n  - "), strings.Join(cats, "\n  - "))
}
