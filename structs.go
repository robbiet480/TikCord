package main

import (
	"bytes"
	"strconv"
	"time"
)

// Time defines a timestamp encoded as epoch seconds in JSON
type Time time.Time

// MarshalJSON is used to convert the timestamp to JSON
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(t).Unix(), 10)), nil
}

// UnmarshalJSON is used to convert the timestamp from JSON
func (t *Time) UnmarshalJSON(s []byte) (err error) {
	s = bytes.Trim(s, `"`)
	r := string(s)
	q, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return err
	}
	*(*time.Time)(t) = time.Unix(q, 0)
	return nil
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t Time) Unix() int64 {
	return time.Time(t).Unix()
}

// Time returns the JSON time as a time.Time instance in UTC
func (t Time) Time() time.Time {
	return time.Time(t).UTC()
}

// String returns t as a formatted string
func (t Time) String() string {
	return t.Time().String()
}

type PageData struct {
	Props struct {
		PageProps struct {
			VideoData struct {
				ItemInfos struct {
					ID    string `json:"id"`
					Video struct {
						Urls      []string `json:"urls"`
						VideoMeta struct {
							Width    int `json:"width"`
							Height   int `json:"height"`
							Ratio    int `json:"ratio"`
							Duration int `json:"duration"`
						} `json:"videoMeta"`
					} `json:"video"`
					Covers         []string      `json:"covers"`
					AuthorID       string        `json:"authorId"`
					CoversOrigin   []string      `json:"coversOrigin"`
					ShareCover     []string      `json:"shareCover"`
					Text           string        `json:"text"`
					CommentCount   int           `json:"commentCount"`
					DiggCount      int           `json:"diggCount"`
					PlayCount      int           `json:"playCount"`
					ShareCount     int           `json:"shareCount"`
					CreateTime     Time          `json:"createTime"`
					IsActivityItem bool          `json:"isActivityItem"`
					WarnInfo       []interface{} `json:"warnInfo"`
					Liked          bool          `json:"liked"`
					CommentStatus  int           `json:"commentStatus"`
					ShowNotPass    bool          `json:"showNotPass"`
				} `json:"itemInfos"`
				AuthorInfos struct {
					Verified bool     `json:"verified"`
					SecUID   string   `json:"secUid"`
					UniqueID string   `json:"uniqueId"`
					UserID   string   `json:"userId"`
					NickName string   `json:"nickName"`
					Covers   []string `json:"covers"`
					Relation int      `json:"relation"`
				} `json:"authorInfos"`
				MusicInfos struct {
					MusicID    string   `json:"musicId"`
					MusicName  string   `json:"musicName"`
					AuthorName string   `json:"authorName"`
					Covers     []string `json:"covers"`
				} `json:"musicInfos"`
				AuthorStats struct {
					FollowerCount int    `json:"followerCount"`
					HeartCount    string `json:"heartCount"`
				} `json:"authorStats"`
				ChallengeInfoList []struct {
					ChallengeID   string `json:"challengeId"`
					ChallengeName string `json:"challengeName"`
				} `json:"challengeInfoList"`
				DuetInfo  string `json:"duetInfo"`
				TextExtra []struct {
					AwemeID     string `json:"AwemeId"`
					Start       int    `json:"Start"`
					End         int    `json:"End"`
					HashtagName string `json:"HashtagName"`
					HashtagID   string `json:"HashtagId"`
					Type        int    `json:"Type"`
					UserID      string `json:"UserId"`
					IsCommerce  bool   `json:"IsCommerce"`
				} `json:"textExtra"`
				StickerTextList []interface{} `json:"stickerTextList"`
			} `json:"videoData"`
		} `json:"pageProps"`
	} `json:"props"`
}
