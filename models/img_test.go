package models_test

import (
	"testing"

	"github.com/recoilme/tgram/models"
)

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
	![descr descr](wrongurl)
	`
	s, err := models.ImgProcess(s, "ru", "recoilme")
	if err != nil {
		t.Error(err)
	} else {
		//fmt.Printf("s:'%s'", s)
	}

}
