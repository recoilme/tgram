# Tgram (medium clone on pure go)

Based on RealWorld Example


> ### Tgram codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


This codebase was created to demonstrate a fully fledged fullstack application built with **Golang/Gin/Slowpoke** including CRUD operations, authentication, routing, pagination, and more.


[Based on this backend implementation](https://github.com/recoilme/golang-gin-realworld-example-app) 

This app use [slowpoke](https://github.com/recoilme/slowpoke) as database. Package slowpoke implements a low-level key/value store in pure Go.

![slowpoke](http://tggram.com/media/recoilme/photos/file_488344.jpg)


# How it works
```
.
├── hello.go
├── common
│   ├── utils.go        //small tools function
└── users
    ├── models.go       //data models define & DB operation
    ├── serializers.go  //response computing & format
    ├── routers.go      //business logic & router binding
    ├── middlewares.go  //put the before & after logic of handle request
    └── validators.go   //form/json checker
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
## Install Govendor & Fresh
I used Govendor manage the package, and Fresh can help build without reload

https://github.com/kardianos/govendor

https://github.com/pilu/fresh
```
cd 
go get -u github.com/kardianos/govendor
go get -u github.com/pilu/fresh
```

## Start
```
➜  govendor sync
➜  govendor add +external
➜  fresh
```


## Testing
```
➜  go test -v ./... -cover
```

## Todo
- More elegance config
