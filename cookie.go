package main

import "net/http"

func NewCookie(name string, value string, path string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	}
}
