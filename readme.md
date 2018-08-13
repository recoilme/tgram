# Typegram

[typegram](http://tgr.am) - zen platform for writers and their subscribers with a minimalistic design.

**Basic Capabilities**

 - publications
 - comments
 - favorites
 - subscriptions
 - mentions
 - and so on

**Playground**

You can try the service on a special [test site](http://tst.tgr.am/). Please! Use this [playground](http://tst.tgr.am/) for play with engine!

**Localization**

The service is available, for example, for [Russian-speaking](http://ru.tgr.am/), or [English-speaking](http://en.tgr.am/) users. In process of development, platforms for other languages are opened. On each subdomain, users and publications are separate. Please, help me to translate welcome post for your language
[Add my country](https://github.com/recoilme/tgram/issues/43)

**Optimization**

The first thing that catches your eye is the high speed of loading pages and aggressive optimization.
![](https://tst.tgr.am/i/tst/recoilme/17_.png)


You will not find third-party scripts that monitor user behavior or heavy-weight styles / images. The site works with javascript turned off, it remains fast and convenient on any platform.

**Subscriptions**

On the main page, the author you are subscribed to is displayed, and the number of new publications. The link leads to the first unread message, in chronological order, as in telegram. Typegram does not impose on you whom and when to read.
![](https://tst.tgr.am/i/tst/recoilme/23_.png)

**Mentions**

When someone mention you in comments you will see it on the main page
![](https://tst.tgr.am/i/tst/recoilme/22_.png)

**Editor**

The editor supports both typing in markdown markup, with rich features and visual formatting. With the ability to deploy the post to the full screen, preview, autosave and other convenient "chips"
![](https://en.tgr.am/i/en/recoilme/2_.png)

**Openness**

The project is free and open source. I always welcome comments and suggestions on [github](https://github.com/recoilme/tgram)

**Weekly digest**

[subscribe on weekly digest](https://www.producthunt.com/upcoming/typegram)

**App**

This app use [slowpoke](https://github.com/recoilme/slowpoke) as database. Package slowpoke implements a low-level key/value store in pure Go. This database engine was developed specially for typegram

![slowpoke](https://en.tgr.am/i/en/recoilme/3_.png)


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


You need only golang for start typegram

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
## Fresh 

Fresh can help build without reload
```
go get -u github.com/pilu/fresh
```


## Env

You may create file tgram.env with params, sample:
```
TGRAMPWD=SOM2324&E*&Ff!!EDjweljf
TGRAMPORT=:8081
TGRAMTITLE=typegram
TGRAMNAME=Typegram
TGRAMDESC=zen platform for writers
TGRAMADMIN=recoilme
TGRAMABOUT=/@recoilme/1
TGRAMDOMAIN=tgr.am
```


## Start
```
➜  go get ./...
➜  go build
➜  ./tgram
```

## Thanks

[awsm.css](https://github.com/igoradamenko/awsm.css)


[realworld.io](https://realworld.io)


[inscryb-markdown-editor](https://github.com/Inscryb/inscryb-markdown-editor)


## Design

[egorabaturov](https://egorabaturov.com)

## Dev branch

- master

## Contributors

You are welcome!

## Plans


I try to build the new big thing for blogging) Mutch more than just text version of medium. But i started from basics.
Mobile version will be an incredible mix of messenger for writers and readers. Subscribe to be first!

https://www.producthunt.com/upcoming/typegram