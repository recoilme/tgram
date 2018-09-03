package models

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strconv"
	"time"

	sp "github.com/recoilme/slowpoke"
)

const (
	dbAid   = "db/%s/aid"
	dbAids  = "db/%s/aids"
	dbAUser = "db/%s/a/%s"
	dbView  = "db/%s/view"
	dbATag  = "db/%s/t/%s"
	dbDau   = "db/%s/dau/%d/%d/%d"
)

// Article model
type Article struct {
	ID          uint32
	Title       string `form:"title" json:"title" binding:"max=255"`
	Body        string `form:"body" json:"body" binding:"exists,min=10,max=65536"`
	Author      string
	Image       string
	OgImage     string `form:"ogimage" json:"ogimage" binding:"omitempty,url"`
	CreatedAt   time.Time
	Lang        string
	HTML        template.HTML
	Plus        uint32
	Minus       uint32
	Comments    []Article
	ReadingTime int
	WordCount   int
	Tag         string `form:"tag" json:"tag" binding:"omitempty,alphanum,max=20"`
}

// Uint32toBin convert to binary
func Uint32toBin(id uint32) []byte {
	id32 := make([]byte, 4)
	binary.BigEndian.PutUint32(id32, id)
	return id32
}

// BintoUint32 convert to uint32
func BintoUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

// ArticleNew create article
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

	// tag
	if a.Tag != "" {
		fATag := fmt.Sprintf(dbATag, a.Lang, a.Tag)
		sp.Set(fATag, id32, []byte(a.Author))
	}

	// uid
	fAUser := fmt.Sprintf(dbAUser, a.Lang, a.Author)
	// store
	return a.ID, sp.SetGob(fAUser, id32, a)
}

// ArticleUpd update article
func ArticleUpd(a *Article, oldTag string) (err error) {
	// tag check
	if a.Tag != oldTag {
		if oldTag != "" {
			//remove old tag
			sp.Delete(fmt.Sprintf(dbATag, a.Lang, oldTag), Uint32toBin(a.ID))
		}
		//set new tag
		sp.Set(fmt.Sprintf(dbATag, a.Lang, a.Tag), Uint32toBin(a.ID), []byte(a.Author))
	}
	fAUser := fmt.Sprintf(dbAUser, a.Lang, a.Author)
	return sp.SetGob(fAUser, Uint32toBin(a.ID), a)
}

// ArticleGet get article
func ArticleGet(lang, username string, aid uint32) (a *Article, err error) {
	fAUser := fmt.Sprintf(dbAUser, lang, username)

	err = sp.GetGob(fAUser, Uint32toBin(aid), &a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ArticleDelete delete article
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

func ArticlesSelect(lang, fAids string, from []byte, limit, offset uint32, asc bool) (models []Article, first, last uint32, err error) {
	keys, err := sp.Keys(fAids, from, limit, offset, asc)
	//log.Println(fAids, keys)
	if err != nil {
		return models, first, last, err
	}
	for _, key := range keys {
		var model Article
		uidb, err := sp.Get(fAids, key)
		if err != nil {
			break
			//continue
		}
		fAUser := fmt.Sprintf(dbAUser, lang, string(uidb))
		if err = sp.GetGob(fAUser, key, &model); err != nil {
			break
			//continue
		}
		if first == 0 {
			first = BintoUint32(key)
		}
		last = BintoUint32(key)
		models = append(models, model)
	}
	return models, first, last, err
}

// AllArticles return page from list of articles
func AllArticles(lang, from_str, tag string) (models []Article, page string, prev, next, last uint32, err error) {
	//log.Println("tag:", tag)
	from_int, _ := strconv.Atoi(from_str)
	var limit_int uint32
	limit_int = 5

	var from []byte
	if from_int > 0 {
		from = Uint32toBin(uint32(from_int))
	} else {
		from = nil
	}
	var fAids = fmt.Sprintf(dbAids, lang)
	if tag != "" {
		fAids = fmt.Sprintf(dbATag, lang, tag)
	}
	models, firstkey, next, err := ArticlesSelect(lang, fAids, from, limit_int, uint32(0), false)
	//all, _ := sp.Count(fAids)
	page = fmt.Sprintf("%d..%d", firstkey, next)

	// last article is prev to first article
	//fAids := fmt.Sprintf(dbAids, lang)
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

func TopArticles(lang string, cnt uint32, by string) (models []Article, err error) {

	models, _, _, err = ArticlesSelect(lang, fmt.Sprintf(dbAids, lang), nil, cnt*5, uint32(0), false)
	if err != nil {
		return models, err
	}
	sorted, err := ArticlesSort(models, by, cnt)
	return sorted, err
}

func ArticlesSort(models []Article, by string, cnt uint32) (sorted []Article, err error) {
	type ArticleSort struct {
		Article Article
		Score   float64
	}
	now := time.Now()
	var tmp []ArticleSort
	for _, a := range models {
		diff := now.Sub(a.CreatedAt)
		min := diff.Minutes()
		var vote float64
		switch by {
		case "minus":
			vote = float64(a.Minus)
		default:
			vote = float64(a.Plus)
		}
		//https://medium.com/hacking-and-gonzo/how-hacker-news-ranking-algorithm-works-1d9b0cf2c08d
		score := vote / (math.Pow(float64(min+120), float64(1.8)))
		//log.Println(a.Title, a.Plus, int(min), score)
		var as ArticleSort
		as.Article = a
		as.Score = score
		tmp = append(tmp, as)
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Score > tmp[j].Score
	})

	cur := 0
	for _, nm := range tmp {
		cur++
		sorted = append(sorted, nm.Article)
		if cur >= int(cnt) {
			break
		}
	}
	return sorted, err
}

// ArticlesAuthor return articles by author
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

// CommentNew create comment
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

// Favorites return 100 last Favorites
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

// ViewSet counter view by aid
func ViewSet(lang string, aid uint32, v int) {
	go sp.Set(fmt.Sprintf(dbView, lang), Uint32toBin(aid), Uint32toBin(uint32(v)))
}

// ViewGet return stored counter
func ViewGet(lang string, aid uint32) (v int) {
	v = 1
	b, err := sp.Get(fmt.Sprintf(dbView, lang), Uint32toBin(aid))
	if err == nil {
		v = int(BintoUint32(b))
	}
	return v
}

// DauSet log of dau
func DauSet(lang, ip string) {
	t := time.Now()
	year, month, day := t.Date()
	stat := fmt.Sprintf(dbDau, lang, year, int(month), day)
	has, _ := sp.Has(stat, []byte(ip))
	if !has {
		go sp.Set(stat, []byte(ip), nil)
	}
}

// DauGet return dau
func DauGet(lang string) int {
	t := time.Now()
	year, month, day := t.Date()
	stat := fmt.Sprintf(dbDau, lang, year, int(month), day)
	cnt, _ := sp.Count(stat)
	return int(cnt)
}

// WauGet return wau
func WauGet(lang string) int {
	t := time.Now()
	ips := make(map[string]int)
	for i := 1; i <= 7; i++ {
		year, month, day := t.Date()
		stat := fmt.Sprintf(dbDau, lang, year, int(month), day)
		//log.Println(stat)
		keys, err := sp.Keys(stat, nil, uint32(0), uint32(0), true)
		if err != nil {
			continue
		}
		for _, key := range keys {
			ips[string(key)]++
			//log.Println(string(key))
		}
		t = t.AddDate(0, 0, -1)
	}
	return len(ips)
}
