package models_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"regexp"
	"strings"
	"testing"

	sp "github.com/recoilme/slowpoke"
	"github.com/recoilme/tgram/models"
)

func TestPngCreate(t *testing.T) {
	const width, height = 256, 256

	// Create a colored image of the given width and height.
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8((x + y) & 255),
				G: uint8((x + y) << 1 & 255),
				B: uint8((x + y) << 2 & 255),
				A: 255,
			})
		}
	}
	smallb := new(bytes.Buffer)
	e := png.Encode(smallb, img)
	if e != nil {
		fmt.Println("e", e)
	} else {
		//fmt.Println(smallb.Bytes())
	}
	fo := "img/image.png"
	sp.Set(fo, []byte("1"), smallb.Bytes())
	sp.Close(fo)

}
func TestImgProcess(t *testing.T) {
	s := `Очень большая картинка без оптимизаций
	a peach

	![](https://cdn-images-1.medium.com/max/2000/1*dT8VX9g8ig6lxmobTRmCiA.jpeg)
	[tgr.am](http://tgr.am) - дзэн сервис для писателей и читателей с минималистичным дизайном, удобным интерфейсом и высокой скоростью работы.

	Тут можно:
	 - публиковать посты
	 - комментировать
	 - добавлять в избранное
	 - подписываться на авторов
	
	Сервис доступен для [русскоязычных](http://ru.tgr.am/), и  [англоязычных](http://en.tgr.am/) пользователей. Потестировать  сервис можно на специальной [тестовой площадке](http://tst.tgr.am/).
	
	Авторы - пожалуйста, уважайте читателей. Не публикуйте спам, рекламу, запрещенный и/или защищенный авторским правом контент. Посты с подобным содержанием будут удалены, а их авторы - заблокированы.
	
	Будьте хорошим пользователем!
	
	Проект бесплатен и с открытым исходным кодом. Буду рад замечаниям и предложениям на [github](https://github.com/recoilme/tgram) проекта.
	
	С уважением, [@recoilme](http://ru.tgr.am/@recoilme)
	![descr descr](https://image.freepik.com/free-vector/industrial-machine-vector_23-2147498405.jpg)
	![descr descr](http://tggram.com/media/daokedao/photos/file_826207.jpg)
	![descr descr](http://tst.tgr.am/m/img/logo_big.png)
	`
	s, err := models.ImgProcess(s, "ru", "recoilme", "http://sub.localhost:8081/")
	if err != nil {
		t.Error(err)
	} else {
		//fmt.Printf("s:'%s'", s)
	}

}

func TestFindUser(t *testing.T) {
	r, e := regexp.Compile(`@[a-z0-9]*`)
	if e != nil {
		return
	}
	s := "@ee2  asdsd\n@wqw @ \n@re3 @4re @32 @6ffg& git commit -m '@abc'"
	var arrayFrom = []string{}
	var arrayTo = []string{}
	submatchall := r.FindAllString(s, -1)
	for _, element := range submatchall {
		if len(element) < 2 {
			continue
		}
		fmt.Println(element)
		arrayFrom = append(arrayFrom, element)
		newElement := "[" + element +
			"](/" + element + ")"
		arrayTo = append(arrayTo, newElement)
	}
	if len(arrayFrom) > 0 {

		ss := strings.NewReplacer(models.Zip(arrayFrom, arrayTo)...).Replace(s)
		log.Println(ss)
	}
}

func TestTgImgProcess(t *testing.T) {
	models.TgClickableImage()
}
