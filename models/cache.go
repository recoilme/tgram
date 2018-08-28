package models

import (
	"errors"
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var (
	cc *cache.Cache
)

const (
	RateIP      = 10 * time.Minute
	RatePost    = 10 * time.Minute
	RateComment = 30 * time.Second

	VoteComStore = 24 * time.Hour
	VoteArtStore = 24 * time.Hour

	VoteComMax = 10
	VoteArtMax = 10
)

func init() {
	cc = cache.New(24*time.Hour, 10*time.Minute)
}

func RegisterIPSet(ip string) {
	cc.Set(ip, time.Now().Unix(), cache.DefaultExpiration)
}

func RegisterIPGet(ip string) int {
	return ratelimit(ip, RateIP)
}

func PostLimitGet(lang, username string) int {
	postRate := lang + ":p:" + username
	return ratelimit(postRate, RatePost)
}

func PostLimitSet(lang, username string) {
	postRate := lang + ":p:" + username
	cc.Set(postRate, time.Now().Unix(), cache.DefaultExpiration)
}

func PostLimitDel(lang, username string) {
	postRate := lang + ":p:" + username
	cc.Delete(postRate)
}

func ComLimitSet(lang, username string) {
	rateComKey := lang + ":c:" + username
	cc.Set(rateComKey, time.Now().Unix(), cache.DefaultExpiration)
}

func ComLimitGet(lang, username string) int {
	rateComKey := lang + ":c:" + username
	return ratelimit(rateComKey, RateComment)
}

func UserBanGet(username string) bool {
	_, bannedAuthor := cc.Get("ban:uid:" + username)
	return bannedAuthor
}

func UserBanSet(author string) {
	cc.Set("ban:uid:"+author, time.Now().Unix(), cache.DefaultExpiration)
}

func ratelimit(key string, dur time.Duration) (wait int) {
	if key == "" {
		return 0
	}
	if x, found := cc.Get(key); found {
		// if found
		t := time.Now()
		elapsed := t.Sub(time.Unix(x.(int64), 0))

		if elapsed < dur {
			wait = int((dur - elapsed).Seconds())
		}
	}
	return wait
}

func ArticleViewGet(lang, ip string, aid uint32) (view int) {
	unic := fmt.Sprintf("%s%d%s", lang, aid, ip)
	unicCnt := fmt.Sprintf("%s%d", lang, aid)

	if _, found := cc.Get(unic); !found {
		// new unique view for last 24 h - increment
		cc.Set(unic, 0, cache.DefaultExpiration) //store on 24 h
		v, notfounderr := cc.IncrementInt(unicCnt, 1)
		if notfounderr != nil {
			stored := ViewGet(lang, aid)
			//log.Println("stored", stored)
			cc.Add(unicCnt, stored, cache.NoExpiration)
			view = stored
		} else {
			view = v
			//if v%5 == 0 {
			ViewSet(lang, aid, v)
			//}
		}
	} else {
		if x, f := cc.Get(unicCnt); f {
			view = x.(int)
		} else {
			view = ViewGet(lang, aid)
		}
	}
	return view
}

func ComUpSet(lang, username, cid string) error {
	unicCnt := fmt.Sprintf("%s:cuidcnt:%s", lang, username)

	if val, found := cc.Get(unicCnt); !found {
		//no votes
		cc.Add(unicCnt, 1, VoteComStore)
	} else {
		// found
		votes := val.(int)
		//log.Println(votes)
		if votes >= VoteComMax {
			// limit
			return errors.New("Oh: today comment vote limit exceeded(")
		}
		// add vote
		cc.IncrementInt(unicCnt, 1)
	}
	// one comment - one vote
	uniq := fmt.Sprintf("%s:ciduid:%s:%s", lang, cid, username)
	if _, found := cc.Get(uniq); !found {
		// uniq
		cc.Set(uniq, 1, 24*30*time.Hour) // 30 days
	} else {
		return errors.New("Oh: only one vote for each comment allowed(")
	}

	return nil
}

func VoteSet(lang, username string) error {
	unicCnt := fmt.Sprintf("%s:auidcnt:%s", lang, username)

	if val, found := cc.Get(unicCnt); !found {
		//no votes
		cc.Add(unicCnt, 1, VoteArtStore)
	} else {
		// found
		votes := val.(int)
		//log.Println(votes)
		if votes >= VoteArtMax {
			// limit
			return errors.New("Oh: today article vote limit exceeded(")
		}
		// add vote
		cc.IncrementInt(unicCnt, 1)
	}
	return nil
}

func Type2TeleSet(aid uint32, mid int) {
	cc.Set(fmt.Sprintf("type2tele:%d", aid), mid, cache.DefaultExpiration)
}

func Type2TeleGet(aid uint32) int {
	if x, found := cc.Get(fmt.Sprintf("type2tele:%d", aid)); found {
		// if found
		return x.(int)
	}
	return 0
}
