package articles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sp "github.com/recoilme/slowpoke"
	"github.com/recoilme/tgram/common"
	"github.com/recoilme/tgram/users"
	"github.com/stretchr/testify/assert"
)

func TestRandString(t *testing.T) {
	asserts := assert.New(t)

	str := "RandString"
	asserts.Equal(len(str), 10, "length should be 10")
}

func TestNewGob(t *testing.T) {
	asserts := assert.New(t)

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(42))
	err := sp.SetGob("db/test", b, 13)
	asserts.Nil(err)
	keys, err := sp.Keys("db/test", nil, uint32(0), uint32(0), false)
	asserts.Nil(err)
	//fmt.Println(keys)

	vals := sp.Gets("db/test", keys)
	//fmt.Println(vals)
	var val int
	sp.GetGob("db/test", vals[0], &val)
	//fmt.Println(val)
	asserts.Equal(13, val)
	sp.DeleteFile("db/test")
}

func HeaderTokenMock(req *http.Request, u uint32) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %v", common.GenToken(u)))
}

func TestWithoutAuth(t *testing.T) {
	asserts := assert.New(t)
	//You could write the reset database code here if you want to create a database for this block
	//resetDB()

	r := gin.New()

	users.UsersRegister(r.Group("/users"))
	r.Use(users.AuthMiddleware(false))
	ArticlesAnonymousRegister(r.Group("/articles"))
	TagsAnonymousRegister(r.Group("/tags"))

	r.Use(users.AuthMiddleware(true))
	users.UserRegister(r.Group("/user"))
	users.ProfileRegister(r.Group("/profiles"))

	ArticlesRegister(r.Group("/articles"))

	stop := 19
	for num, testData := range unauthRequestTests {
		if num >= stop {
			break
		}
		bodyData := testData.bodyData
		req, err := http.NewRequest(testData.method, testData.url, bytes.NewBufferString(bodyData))
		req.Header.Set("Content-Type", "application/json")
		asserts.NoError(err)

		testData.init(req)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		res := asserts.Equal(testData.expectedCode, w.Code, "Response Status - "+testData.msg)
		fmt.Println("\n\nTest:", num, testData.msg, " Response Content - ", w.Body.String())
		fmt.Println()

		asserts.Regexp(testData.responseRegexg, w.Body.String(), "Response Content - "+testData.msg)
		if !res {
			fmt.Println("num", num)
		}

	}
}

var unauthRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	//Testing will run one by one, so you can combine it to a user story till another init().
	//And you can modified the header or body in the func(req *http.Request) {}

	//---------------------   Testing for user register   ---------------------
	{
		func(req *http.Request) {
			common.ResetUsersDBWithMock()
		},
		"/users",
		"POST",
		`{"user":{"username": "user1","email": "e@mail.ru","password": "password","image":"http://tggram.com/media/natus_vincere_official/profile_photos/file_655143.jpg"}}`,
		http.StatusCreated,
		``,
		"valid data and should return StatusCreated",
	},

	{
		func(req *http.Request) {
		},
		"/users",
		"POST",
		`{"user":{"username": "user2","email": "e2@mail.ru","password": "password","image":"http://tggram.com/media/recoilmeblog/profile_photos/file_654897.jpg"}}`,
		http.StatusCreated,
		``,
		"valid data and should return StatusCreated",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/profiles/user1",
		"GET",
		``,
		http.StatusOK,
		`{"profile":{"username":"user1","bio":"","image":"http://tggram.com/media/natus_vincere_official/profile_photos/file_655143.jpg","following":false}}`,
		"request should return self profile",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles",
		"POST",
		`{
			"article": {
				"title": "How to train your dragon",
				"description": "Ever wonder how?",
				"body": "You have to believe"
			}
		}`,
		http.StatusCreated,
		``,
		"request should create article",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles/how-to-train-your-dragon",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return article",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/articles",
		"POST",
		`{
			"article": {
				"title": "How are you?",
				"description": "how?",
				"body": "I'm fine, thank you"
			}
		}`,
		http.StatusCreated,
		``,
		"request should create second article",
	},

	{
		func(req *http.Request) {

		},
		"/articles",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return 2 articles",
	},

	{
		func(req *http.Request) {

		},
		"/articles?author=user2",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return article by uid2",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/articles/how-to-train-your-dragon/comments",
		"POST",
		`{
			"comment": {
				"body": "His name was my name too."
			}
		}`,
		http.StatusCreated,
		``,
		"request should create comment from 2",
	},
	//second comment
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles/how-to-train-your-dragon/comments",
		"POST",
		`{
			"comment": {
				"body": "Second comment from user1."
			}
		}`,
		http.StatusCreated,
		``,
		"request should create comment from 1",
	},
	//get comments anonim
	{
		func(req *http.Request) {

		},
		"/articles/how-to-train-your-dragon/comments",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return 2 comments",
	},

	//delete comment 2
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles/how-to-train-your-dragon/comments/2",
		"DELETE",
		``,
		http.StatusOK,
		``,
		"request should delete second comment",
	},

	//get comments anonim
	{
		func(req *http.Request) {

		},
		"/articles/how-to-train-your-dragon/comments",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return 1 comments",
	},

	//delete article 2

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 1)
		},
		"/articles/how-are-you",
		"DELETE",
		``,
		http.StatusOK,
		``,
		"request should delete article 2",
	},

	{
		func(req *http.Request) {

		},
		"/articles",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return 1 articles",
	},

	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/articles",
		"POST",
		`{
			"article": {
				"title": "How are you2?",
				"description": "how?2",
				"body": "I'm fine, thank you2"
			}
		}`,
		http.StatusCreated,
		``,
		"request should create 3 article",
	},

	//follow
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/profiles/user1/follow",
		"POST",
		``,
		http.StatusOK,
		``,
		"user follow another should work",
	},
	//get follow feed
	{
		func(req *http.Request) {
			HeaderTokenMock(req, 2)
		},
		"/articles/feed",
		"GET",
		``,
		http.StatusOK,
		``,
		"request should return 1 article",
	},
}
