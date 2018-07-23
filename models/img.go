package models

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"regexp"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	_ "golang.org/x/image/webp"

	_ "golang.org/x/image/bmp"

	"github.com/MaxHalford/halfgone"
	"github.com/nfnt/resize"
	sp "github.com/recoilme/slowpoke"
	"github.com/recoilme/tgram/utils"
)

const (
	// lang/username/img/ext
	fileImg = "img/%s/%s/%d%s"

	// lang/username/counter
	dbImgID = "img/%s/%s/id"

	// max size
	maxSize = 10240000
)

// ImgProcess extract img links from markdown, download, and replace with local copy
func ImgProcess(s, lang, username, host string) (res string, err error) {
	//fmt.Printf("s:'%s'", s)
	r, err := regexp.Compile(`!\[(.*?)\]\((.*?)\)`)
	if err != nil {
		return s, err
	}
	//res = r.ReplaceAllStringFunc(s, isImg)
	//r.FindAllString(s,)
	var arrayFrom = []string{}
	var arrayTo = []string{}

	submatchall := r.FindAllString(s, -1)
	for _, element := range submatchall {
		if strings.Contains(element, host) {
			continue
		}
		b, href := isImg(element)
		if b != nil {
			file, orig := store(href, lang, username, b)
			if file == "" || orig == "" {
				continue
			}
			arrayFrom = append(arrayFrom, element)

			newElement := "[" + strings.Replace(element, href, (host+file), 1) +
				"](" + host + orig + ")"
			//log.Println(element, newElement, href, file, orig)
			arrayTo = append(arrayTo, newElement)
		}
	}
	if len(arrayFrom) > 0 {
		zip := func(a1, a2 []string) []string {
			r := make([]string, 2*len(a1))
			for i, e := range a1 {
				r[i*2] = e
				r[i*2+1] = a2[i]
			}
			return r
		}
		s = strings.NewReplacer(zip(arrayFrom, arrayTo)...).Replace(s)
	}

	res = s
	return res, err
}

func store(href, lang, username string, b []byte) (file, orig string) {
	//image processing
	if img, _, err := image.Decode(bytes.NewReader(b)); err == nil {
		thumb := resize.Thumbnail(800, 800, img, resize.MitchellNetravali)
		thumbb := new(bytes.Buffer)
		if err := png.Encode(thumbb, thumb); err == nil {
			atk := halfgone.AtkinsonDitherer{}.Apply(halfgone.ImageToGray(thumb))
			//store
			if imgid, err := sp.Counter(fmt.Sprintf(dbImgID, lang, username), []byte("id")); err == nil {
				path := fmt.Sprintf(fileImg, lang, username, imgid, ".png")
				if _, err := utils.СheckAndCreate(path); err == nil {
					// save Atkinson
					f, err := os.Create(path)
					defer f.Close()
					if err == nil {
						if err := png.Encode(f, atk); err == nil {
							file = "i" + path[3:]
						}
					}
					// save orig
					pathOrig := fmt.Sprintf(fileImg, lang, username, imgid, "_.png")
					fo, err := os.Create(pathOrig)
					defer fo.Close()
					if err == nil {
						if err := png.Encode(fo, thumb); err == nil {
							orig = "i" + pathOrig[3:]
						}
					}

				}

			}

		}
	}
	return file, orig
}

func isImg(s string) ([]byte, string) {
	//fmt.Println("img:", s)
	var href = ""
	//var err error
	first := strings.IndexByte(s, '(') + 1
	last := strings.IndexByte(s, ')')
	if first > 0 && last > 0 && last > first {
		// extract link
		href = s[first:last]
		len := utils.HTTPImgLen(href)
		//log.Println("href", href, "len", len)
		if len > 0 && len < maxSize {
			//try download
			b := utils.HTTPGetBody(href)
			if b != nil {
				return b, href

			} else {
				//log.Println("b is nil")
			}
		}
		//fmt.Println(href)
		//href = "http://ya.ru"
		//s = s[:first] + href + s[last:]
		//fmt.Println(s)
	}

	return nil, ""
}

// convertToPNG converts from any recognized format to PNG.
func convertToPNG(w io.Writer, r io.Reader) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}
