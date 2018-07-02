package models

import (
	"encoding/binary"
	"fmt"
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
	CreatedAt time.Time
	Lang      string
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
