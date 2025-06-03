package main

import (
	"larkbot/cmd/internal/adapter/inbound"
	"larkbot/cmd/internal/adapter/outbound"
	"larkbot/cmd/internal/app"
	"log"
	"net/http"
)

func main() {
	lark := outbound.NewLark(
		"https://open.larksuite.com/open-apis/bot/v2/hook/88fccfea-8fad-47d9-99a9-44d214785fff",
		"cli_a8b2c70af7389029",
		"QUbHQALAU0xrxWid9QU8Hb50wpY1wtwv",
	)

	usecase := app.NewWebhookUsecase(lark)
	handler := inbound.NewWebhookHandler(usecase)

	http.HandleFunc("/git-webhook", handler.Handle)

	log.Printf("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
>>>>>>> ed45973deec9451b6001e1f8bc0a548e813b2153
