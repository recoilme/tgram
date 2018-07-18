# Tgram

[tgr.am] (http://tgr.am) - Zen service for writers and readers with minimalistic design, user-friendly interface and high speed of work.


Main features:
 - publish posts
 - comment on
 - add to favorites
 - subscribe to authors


The service is available for [Russian-speaking] (http://ru.tgr.am/), and [English-speaking] (http://en.tgr.am/) users. You can test the service on a special [test site] (http://tst.tgr.am/).


Authors - please respect the readers. Do not publish spam, advertising, illegal and / or copyrighted content. Posts with similar content will be deleted, and their authors - blocked.


This app use [slowpoke](https://github.com/recoilme/slowpoke) as database. Package slowpoke implements a low-level key/value store in pure Go.

![slowpoke](http://tggram.com/media/recoilme/photos/file_488344.jpg)


# How it works
```
.
├── main.go
├── routers
│   ├── routers.go      //routers
└── models.go
    ├── article.go      //data models for article
    └── user.go         //data models for user

```

# Getting started

## Install the Golang
https://golang.org/doc/install
## Environment Config
make sure your ~/.*shrc have those varible:
```
➜  echo $GOPATH
/Users/zitwang/test/
➜  echo $GOROOT
/usr/local/go/
➜  echo $PATH
...:/usr/local/go/bin:/Users/zitwang/test//bin:/usr/local/go//bin
```
## Install 
Fresh can help build without reload

https://github.com/pilu/fresh
```
cd 
go get -u github.com/pilu/fresh


➜  go get ./...

```

### Env

You may create file tgram.env with params, sample:
```
TGRAMPWD=SOM2324&E*&Ff!!EDjweljf
TGRAMPORT=:8081
RECAPTCHA=somekey
```


## Start
```
➜  ./tgram
```

## Demo

[tgr.am](http://tgr.am)

## Thanks

[awsm.css](https://github.com/igoradamenko/awsm.css)


[realworld.io](https://realworld.io)


[inscryb-markdown-editor](https://github.com/Inscryb/inscryb-markdown-editor)


## Design

[egorabaturov](https://egorabaturov.com)