package fncmp

import (
	"github.com/kitkitchen/mnemo"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"-"`
}

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func init() {
	var _, _ = mnemo.NewStore("chat")
	mnemo.NewCache[User]("chat", "user")
	mnemo.NewCache[Message]("chat", "message")
}

// func HandleChat(ctx context.Context) FnComponent {

// 	users, err := mnemo.UseCache[User]("chat", "user")
// 	if err != nil {
// 		log.Println(err)
// 		return NewFn(HTML(`<div>error: failed to get users</div>`))
// 	}

// 	username := "guest_" + fmt.Sprint(len(users.GetAll()))

// 	users.Cache(username, &User{
// 		Username: username,
// 		Password: "password",
// 	})

// 	// Set template
// 	NewFn(ChatPage()).SwapTagInner("main").Now(ctx)

// 	return NewFn(Form(
// 		Input("username", "username", true),
// 		Input("content", "content", true),
// 		ButtonSolid("submit", "Send"),
// 	)).WithEvents(func(ctx context.Context) FnComponent {
// 		event, ok := ctx.Value(EventKey).(EventListener)
// 		//TODO: Include original component with event so it can be returned or appended to
// 		if !ok {
// 			return NewFn(HTML(`
// 					<div>Event not found</div>
// 				`))
// 		}
// 		message, err := UnmarshalEventData[Message](event)
// 		if err != nil {
// 			log.Println(err)
// 			return NewFn(ErrorMessage(err)).SwapTargetInner("error_message")
// 		}
// 		return NewFn(HTML(`
// 		<div class="w-full"><p>` + username + ": " + message.Content + `</p></div>
// 		`)).AppendTarget(messages)
// 	}, OnSubmit).AppendTarget(userInput)
// }
