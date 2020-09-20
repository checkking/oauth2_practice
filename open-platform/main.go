package main

import (
	"github.com/checkking/oauth2_practice/open-platform/handlers"
	"github.com/golang/glog"
	"net/http"
)

func main() {
	s := &http.Server{
		Addr:    ":8001",
		Handler: handlers.New(),
	}
	glog.Infof("http sever start, listen at %q", s.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("%v", err)
	} else {
		glog.Errorf("server closed!")
	}
}
