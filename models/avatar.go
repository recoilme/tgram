package models

// backport from https://github.com/recoilme/govatar

import (
	"hash/fnv"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Gender represents gender type
type Gender int

// stupid
type store struct {
	Background []string
	Monster    person
}

// it's stupid too and may be array of array
type person struct {
	Clothes []string
	Eye     []string
	Face    []string
	Hair    []string
	Mouth   []string
}

//type naturalSort []string

const (
	// MONSTER default person type
	MONSTER Gender = iota
)

var (
	//r           = regexp.MustCompile(`[^0-9]+|[0-9]+`)
	assetsStore *store
)

func init() {
	monster := getPerson(MONSTER)
	assetsStore = &store{Background: readAssetsFrom("data/background"), Monster: monster}
	//rand.Seed(time.Now().UTC().UnixNano())
}

func getPerson(gender Gender) person {
	var genderPath string

	switch gender {
	case MONSTER:
		genderPath = "monster"
	}

	return person{
		Clothes: readAssetsFrom("data/" + genderPath + "/clothes"),
		Eye:     readAssetsFrom("data/" + genderPath + "/eye"),
		Face:    readAssetsFrom("data/" + genderPath + "/face"),
		Hair:    readAssetsFrom("data/" + genderPath + "/hair"),
		Mouth:   readAssetsFrom("data/" + genderPath + "/mouth"),
	}
}

func readAssetsFrom(dir string) (assets []string) {

	files, err := ioutil.ReadDir("./" + dir)
	if err != nil {
		log.Println(err)
		return assets
	}

	for _, asset := range files {
		if asset.Name() == ".DS_Store" {
			continue
		}
		// TODO append only if image? see http.DetectContentType(buf)
		assets = append(assets, filepath.Join(dir, asset.Name()))
	}
	// TODO for what?
	//sort.Sort(naturalSort(assets))
	sort.Strings(assets)
	return assets
}

// GenerateFromUsername generates avatar from string
func GenerateMonster(username string) (image.Image, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(username))
	if err != nil {
		return nil, err
	}

	rnd := rand.New(rand.NewSource(int64(h.Sum32())))
	avatar := image.NewRGBA(image.Rect(0, 0, 200, 200))
	p := assetsStore.Monster
	//var err error
	err = drawImg(avatar, randSliceString(rnd, assetsStore.Background), err)
	//log.Println("assetsStore.Background", assetsStore.Background, err)
	err = drawImg(avatar, randSliceString(rnd, p.Face), err)
	//log.Println("p.Face", p.Face, err)
	err = drawImg(avatar, randSliceString(rnd, p.Clothes), err)
	//log.Println("p.Clothes", p.Clothes, err)
	err = drawImg(avatar, randSliceString(rnd, p.Eye), err)
	//log.Println("p.Eye", p.Eye, err)
	err = drawImg(avatar, randSliceString(rnd, p.Mouth), err)
	//log.Println("p.Mouth", p.Mouth, err)
	err = drawImg(avatar, randSliceString(rnd, p.Hair), err)
	//log.Println("p.Hair", p.Hair, err)
	return avatar, err
}

func drawImg(dst draw.Image, asset string, err error) error {
	if err != nil {
		return err
	}
	infile, err := os.Open(asset)
	if err != nil {
		return err
	}
	defer infile.Close()
	src, _, err := image.Decode(infile) //bindata.MustAsset(asset)))
	if err != nil {
		return err
	}
	draw.Draw(dst, dst.Bounds(), src, image.Point{0, 0}, draw.Over)
	return nil
}

// randSliceString returns random element from slice of string
func randSliceString(rnd *rand.Rand, slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[randInt(rnd, 0, len(slice))]
}

// randInt returns random integer
func randInt(rnd *rand.Rand, min int, max int) int {
	return min + rnd.Intn(max-min)
}

func SaveToFile(img image.Image, filePath string) error {
	outFile, err := os.Create(filePath)
	defer outFile.Close()
	if err != nil {
		return err
	}
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".jpeg", ".jpg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 80})
	case ".gif":
		err = gif.Encode(outFile, img, nil)
	default:
		err = png.Encode(outFile, img)
	}
	return err
}

/*
func (s naturalSort) Len() int {
	return len(s)
}

func (s naturalSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// TODO refactor this
// why natural sort? it may be not ordered
func (s naturalSort) Less(i, j int) bool {

	spliti := r.FindAllString(strings.Replace(s[i], " ", "", -1), -1)
	splitj := r.FindAllString(strings.Replace(s[j], " ", "", -1), -1)

	for index := 0; index < len(spliti) && index < len(splitj); index++ {
		if spliti[index] != splitj[index] {
			// Both slices are numbers
			if isNumber(spliti[index][0]) && isNumber(splitj[index][0]) {
				// Remove Leading Zeroes
				stringi := strings.TrimLeft(spliti[index], "0")
				stringj := strings.TrimLeft(splitj[index], "0")
				if len(stringi) == len(stringj) {
					for indexchar := 0; indexchar < len(stringi); indexchar++ {
						if stringi[indexchar] != stringj[indexchar] {
							return stringi[indexchar] < stringj[indexchar]
						}
					}
					return len(spliti[index]) < len(splitj[index])
				}
				return len(stringi) < len(stringj)
			}
			// One of the slices is a number (we give precedence to numbers regardless of ASCII table position)
			if isNumber(spliti[index][0]) || isNumber(splitj[index][0]) {
				return isNumber(spliti[index][0])
			}
			// Both slices are not numbers
			return spliti[index] < splitj[index]
		}

	}
	// Fall back for cases where space characters have been annihilated by the replacement call
	// Here we iterate over the unmolested string and prioritize numbers
	for index := 0; index < len(s[i]) && index < len(s[j]); index++ {
		if isNumber(s[i][index]) || isNumber(s[j][index]) {
			return isNumber(s[i][index])
		}
	}
	return s[i] < s[j]
}

func isNumber(input uint8) bool {
	return input >= '0' && input <= '9'
}
*/
