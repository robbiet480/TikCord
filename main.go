package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/message"
	xurls "mvdan.cc/xurls/v2"
)

var (
	token          string
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:80.0) Gecko/20100101 Firefox/80.0"
	httpClient     = &http.Client{}
	collyCollector = colly.NewCollector(colly.AllowURLRevisit())
	textPrinter    = message.NewPrinter(message.MatchLanguage("en"))
	rxStrict       = xurls.Strict()
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	if token == "" {
		log.Fatalln("No token provided. Please run: tikcord -t <bot token>")
		return
	}

	collyCollector.UserAgent = userAgent
	collyCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
	})

	dg, discordClientErr := discordgo.New("Bot " + token)
	if discordClientErr != nil {
		log.Fatalln("Error creating Discord session:", discordClientErr)
		return
	}

	dg.AddHandler(messageCreate)
	if dgOpenErr := dg.Open(); dgOpenErr != nil {
		log.Fatalln("Error opening Discord session:", dgOpenErr)
	}

	log.Infoln("tikcord is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	urls := rxStrict.FindAllString(m.Content, -1)

	if len(urls) > 0 {
		isTyping := false

		for _, url := range urls {
			if !strings.Contains(url, "tiktok.com") {
				continue
			}

			if !isTyping {
				typingErr := s.ChannelTyping(m.ChannelID)
				if typingErr != nil {
					log.Errorln("Error setting typing status", typingErr)
				}
				isTyping = typingErr == nil
			}

			log.Infof("Found TikTok URL in %s from %s: %s", m.ChannelID, m.Author, url)

			videoData, videoDataErr := getVideoData(url)
			if videoDataErr != nil {
				log.Errorln("Error getting video data", videoDataErr)
				continue
			}

			if len(videoData.Props.PageProps.VideoData.ItemInfos.Video.Urls) == 0 {
				log.Warnln("No video URL found in JSON, exiting!")
				continue
			}

			videoFile, closer, videoFileErr := downloadVideo(url, videoData.Props.PageProps.VideoData.ItemInfos.ID, videoData.Props.PageProps.VideoData.ItemInfos.Video.Urls[0])
			if videoFileErr != nil {
				log.Errorln("Error downloading video", videoFileErr)
				continue
			}

			authorIconURL := ""
			if len(videoData.Props.PageProps.VideoData.AuthorInfos.Covers) > 0 {
				authorIconURL = videoData.Props.PageProps.VideoData.AuthorInfos.Covers[0]
			}

			_, messageSendErr := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				File: videoFile,
				Embed: &discordgo.MessageEmbed{
					Description: videoData.Props.PageProps.VideoData.ItemInfos.Text,
					Timestamp:   videoData.Props.PageProps.VideoData.ItemInfos.CreateTime.Time().Format(time.RFC3339),
					URL:         url,
					Author: &discordgo.MessageEmbedAuthor{
						Name:    videoData.Props.PageProps.VideoData.AuthorInfos.NickName,
						URL:     fmt.Sprintf("https://www.tiktok.com/@%s", videoData.Props.PageProps.VideoData.AuthorInfos.UniqueID),
						IconURL: authorIconURL,
					},
					Fields: []*discordgo.MessageEmbedField{
						/*{
							Name:   "Comments",
							Value:  textPrinter.Sprint(videoData.Props.PageProps.VideoData.ItemInfos.CommentCount),
							Inline: true,
						},
						{
							Name:   "Likes",
							Value:  textPrinter.Sprint(videoData.Props.PageProps.VideoData.ItemInfos.DiggCount),
							Inline: true,
						},
						{
							Name:   "Plays",
							Value:  textPrinter.Sprint(videoData.Props.PageProps.VideoData.ItemInfos.PlayCount),
							Inline: true,
						},
						{
							Name:   "Shares",
							Value:  textPrinter.Sprint(videoData.Props.PageProps.VideoData.ItemInfos.ShareCount),
							Inline: true,
						},*/
						{
							Name:  "Link",
							Value: url,
						},
					},
				},
			})
			if messageSendErr != nil {
				log.Errorln("Error sending message", messageSendErr)
			}
			closer.Close()
		}
	}
}

func getVideoData(tikTokURL string) (*PageData, error) {
	var data PageData
	collyCollector.OnHTML("script[id=__NEXT_DATA__]", func(e *colly.HTMLElement) {
		if unmarshalErr := json.Unmarshal([]byte(e.Text), &data); unmarshalErr != nil {
			log.Errorln("error unmarshalling video data", unmarshalErr)
		}
	})
	if visitErr := collyCollector.Visit(tikTokURL); visitErr != nil {
		return nil, fmt.Errorf("error visiting tiktok url: %w", visitErr)
	}

	return &data, nil
}

func downloadVideo(pageURL, videoID, videoURL string) (*discordgo.File, io.ReadCloser, error) {

	req, reqErr := http.NewRequest("GET", videoURL, nil)
	if reqErr != nil {
		return nil, nil, reqErr
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "video/webm,video/ogg,video/*;q=0.9,application/ogg;q=0.7,audio/*;q=0.6,*/*;q=0.5")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", pageURL)
	resp, resErr := httpClient.Do(req)
	if resErr != nil {
		return nil, nil, resErr
	}

	return &discordgo.File{
		Name:        fmt.Sprintf("%s.mp4", videoID),
		ContentType: "video/mp4",
		Reader:      resp.Body,
	}, resp.Body, nil
}
