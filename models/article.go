package models

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"strconv"
	"time"

	sp "github.com/recoilme/slowpoke"
)

const (
	dbAid   = "db/%s/aid"
	dbAids  = "db/%s/aids"
	dbAUser = "db/%s/a/%s"
)

type Article struct {
	ID        uint32
	Title     string `form:"title" json:"title" binding:"max=255"`
	Body      string `form:"body" json:"body" binding:"max=16384"`
	Author    string
	Image     string
	CreatedAt time.Time
	Lang      string
	HTML      template.HTML
	Comments  []Article
}

func Uint32toBin(id uint32) []byte {
	id32 := make([]byte, 4)
	binary.BigEndian.PutUint32(id32, id)
	return id32
}

func BintoUint32(b []byte) uint32 {

	return binary.BigEndian.Uint32(b)
}

func ArticleNew(a *Article) (id uint32, err error) {
	a.CreatedAt = time.Now()
	fAid := fmt.Sprintf(dbAid, a.Lang)

	aid, err := sp.Counter(fAid, []byte("aid"))
	if err != nil {
		return 0, err
	}
	a.ID = uint32(aid)
	id32 := Uint32toBin(a.ID)

	fAids := fmt.Sprintf(dbAids, a.Lang)
	if err = sp.Set(fAids, id32, []byte(a.Author)); err != nil {
		return 0, err
	}

	// uid
	fAUser := fmt.Sprintf(dbAUser, a.Lang, a.Author)
	// store
	return a.ID, sp.SetGob(fAUser, id32, a)
}

func ArticleUpd(a *Article) (err error) {
	fAUser := fmt.Sprintf(dbAUser, a.Lang, a.Author)
	return sp.SetGob(fAUser, Uint32toBin(a.ID), a)
}

func ArticleGet(lang, username string, aid uint32) (a *Article, err error) {
	fAUser := fmt.Sprintf(dbAUser, lang, username)

	err = sp.GetGob(fAUser, Uint32toBin(aid), &a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func ArticleDelete(lang, username string, aid uint32) (err error) {
	fAUser := fmt.Sprintf(dbAUser, lang, username)

	_, err = sp.Delete(fAUser, Uint32toBin(aid))
	if err != nil {
		return err
	}
	fAids := fmt.Sprintf(dbAids, lang)
	sp.Delete(fAids, Uint32toBin(aid))
	return nil
}

func AllArticles(lang, from_str string) (models []Article, page string, prev, next, last uint32, err error) {

	from_int, _ := strconv.Atoi(from_str)
	var limit_int uint32
	limit_int = 5
	fAids := fmt.Sprintf(dbAids, lang)
	var from []byte
	if from_int > 0 {
		from = Uint32toBin(uint32(from_int))
	} else {
		from = nil
	}
	keys, err := sp.Keys(fAids, from, limit_int, uint32(0), false)
	if err != nil {
		return models, page, prev, next, last, err
	}
	var firstkey uint32
	for _, key := range keys {
		var model Article

		uidb, err := sp.Get(fAids, key)
		if err != nil {
			break
		}
		fAUser := fmt.Sprintf(dbAUser, lang, string(uidb))
		if err = sp.GetGob(fAUser, key, &model); err != nil {
			fmt.Println("kerr", err)
			break //todo continue?
		}
		if firstkey == 0 {
			firstkey = BintoUint32(key)
		}
		next = BintoUint32(key)
		models = append(models, model)
	}
	//all, _ := sp.Count(fAids)
	page = fmt.Sprintf("%d..%d", firstkey, next)

	// last article is prev to first article
	lastkeys, _ := sp.Keys(fAids, nil, uint32(1), uint32(1), true)
	if len(lastkeys) > 0 {
		last = BintoUint32(lastkeys[0])
	}
	// prev article
	prevkeys, _ := sp.Keys(fAids, Uint32toBin(firstkey), uint32(1), uint32(5), true)
	if len(prevkeys) > 0 {
		prev = BintoUint32(prevkeys[0])
	}
	if next < last {
		next = last
	}
	return models, page, prev, next, last, err

}

func ArticlesAuthor(lang, username, author, from_str string) (models []Article, page string, prev, next, last uint32, err error) {

	from_int, _ := strconv.Atoi(from_str)
	var limit_int, firstkey uint32
	limit_int = 5

	fAUser := fmt.Sprintf(dbAUser, lang, author)
	var from []byte
	if from_int > 0 {
		from = Uint32toBin(uint32(from_int))
	} else {
		from = nil
	}
	keys, err := sp.Keys(fAUser, from, limit_int, uint32(0), true)
	if err != nil {
		return models, page, prev, next, last, err
	}
	for _, key := range keys {
		var model Article

		if err = sp.GetGob(fAUser, key, &model); err != nil {
			fmt.Println("kerr", err)
			break
		}
		if firstkey == 0 {
			firstkey = BintoUint32(key)
		}
		next = BintoUint32(key)
		models = append(models, model)
	}
	//all, _ := sp.Count(fAUser)
	page = fmt.Sprintf("%d..%d", firstkey, next) //, all)
	// last article is prev to last article
	lastkeys, _ := sp.Keys(fAUser, nil, uint32(1), uint32(1), false)
	if len(lastkeys) > 0 {
		last = BintoUint32(lastkeys[0])
	}
	// prev article
	prevkeys, _ := sp.Keys(fAUser, Uint32toBin(firstkey), uint32(1), uint32(5), false)
	if len(prevkeys) > 0 {
		prev = BintoUint32(prevkeys[0])
	}

	// update last seen if follow
	if username != "" {
		_, slavemaster := GetMasterSlave(author, username)
		smf := fmt.Sprintf(dbSlaveMaster, lang, "fol")

		has, err := sp.Has(smf, slavemaster)
		if err == nil && has {
			b, _ := sp.Get(smf, slavemaster)
			//log.Println("smf", next)
			if len(b) == 4 {
				lastSeen := BintoUint32(b)
				if next > lastSeen {
					sp.Set(smf, slavemaster, Uint32toBin(next))
				}
			} else {
				sp.Set(smf, slavemaster, Uint32toBin(next))
			}
		}
	}
	if next > last {
		next = last
	}
	return models, page, prev, next, last, err
}

func CommentNew(a *Article, user string, mainaid uint32) (id uint32, err error) {
	a.CreatedAt = time.Now()
	fAid := fmt.Sprintf(dbAid, a.Lang)

	aid, err := sp.Counter(fAid, []byte("cid"))
	if err != nil {
		return 0, err
	}
	a.ID = uint32(aid)
	// uid
	fAUser := fmt.Sprintf(dbAUser, a.Lang, user)
	var maina Article
	err = sp.GetGob(fAUser, Uint32toBin(mainaid), &maina)
	if err != nil {
		return 0, err
	}
	maina.Comments = append(maina.Comments, *a)
	//var comments []Article
	// store
	return a.ID, sp.SetGob(fAUser, Uint32toBin(mainaid), maina)
}

func Favorites(lang, u string) (articles []Article) {
	cat := "fav"
	//var err error
	master32 := []byte(u)
	var masterstar = make([]byte, 0)
	masterstar = append(masterstar, master32...)
	masterstar = append(masterstar, '*')
	smf := fmt.Sprintf(dbSlaveMaster, lang, cat)

	keys, _ := sp.Keys(smf, masterstar, uint32(100), 0, false)
	//log.Println("keys", keys)
	lenU := len(u) + 1

	fAids := fmt.Sprintf(dbAids, lang)
	for _, k := range keys {
		aid32 := k[lenU:]
		//log.Println((aid32))
		auser32, err := sp.Get(fAids, aid32)
		//log.Println(string(auser32), err)
		if err == nil {
			var a Article
			fAUser := fmt.Sprintf(dbAUser, lang, string(auser32))
			if err := sp.GetGob(fAUser, aid32, &a); err == nil {
				articles = append(articles, a)
				//log.Println(a)
			}
		}
	}

	return articles
}
