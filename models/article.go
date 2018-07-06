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
	Body      string
	Author    string
	Image     string
	CreatedAt time.Time
	Lang      string
	HTML      template.HTML
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

func AllArticles(lang, limit, offset string) ([]Article, int, error) {
	var models []Article
	var err error
	var cnt uint64
	var offset_int, limit_int int

	offset_int, _ = strconv.Atoi(offset)

	limit_int, _ = strconv.Atoi(limit)
	if limit_int == 0 {
		limit_int = 5
	}

	fAids := fmt.Sprintf(dbAids, lang)
	//if err = sp.Set(fAids, id32, []byte(a.Author)); err != nil {
	keys, err := sp.Keys(fAids, nil, uint32(limit_int), uint32(offset_int), false)
	//log.Println("no params", len(keys), limit_int)
	if err != nil {
		return models, 0, err
	}
	for _, key := range keys {
		var model Article

		uidb, err := sp.Get(fAids, key)
		if err != nil {
			break
		}
		fAUser := fmt.Sprintf(dbAUser, lang, string(uidb))
		if err = sp.GetGob(fAUser, key, &model); err != nil {
			fmt.Println("kerr", err)
			break
		}

		models = append(models, model)
	}
	cnt, _ = sp.Count(fAids)

	//log.Println("models", err, models)
	return models, int(cnt), err

}
