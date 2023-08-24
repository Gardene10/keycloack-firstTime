//pegando dados de autozizaçao

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	clientID     = "myclient"
	clientSecret = "XYQntLp9YXSkBDyqmyBch1wQy6CyNil4"
)

func main() {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, "http://localhost:8080/auth/realms/master")
	if err != nil {
		log.Fatal(err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:8081/auth/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "roles"},
	}

	state := "123"

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/callback", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("state") != state {
			http.Error(writer, " State Invalido", http.StatusBadRequest)
			return
		}

		code := request.URL.Query().Get("code")
		token, err := config.Exchange(ctx, code)
		if err != nil {
			http.Error(writer, "Falha ao trocar o token", http.StatusInternalServerError)
			return
		}

		idToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(writer, "Falha ao gerar o IDToken", http.StatusInternalServerError)
			return
		}
		userInfo, err:= provider.UserInfo(ctx,oauth2.StaticTokenSource(token))
		if err != nil {
			http.Error(writer, "Erro ao pegar UserInfo", http.StatusInternalServerError)
			return
		}


		resp := struct {
			AccessToken *oauth2.Token
			IDToken string
			userInfo *oidc.UserInfo

		}{
			AccessToken: token,
			IDToken: idToken,
			userInfo: userInfo,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}



		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(data)
	})

	addr := ":8081"
	log.Fatal(http.ListenAndServe(addr, nil))
}
