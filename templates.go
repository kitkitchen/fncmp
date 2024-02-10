package main

import "context"

const (
	Content = "content"
)

func HandlePortfolioFn(ctx context.Context) FnComponent {
	return NewFn(SideBarTemplate(
		//TODO: make content dynamic
		HTML(`<div id=`+Content+` class="flex flex-col items-center justify-center h-screen bg-gray-100"></div>`),
		MenuButton("About", "/main"),
		MenuButton("Projects", "/projects"),
		MenuButton("Contact", "/contact"),
	))
}

func MenuButton(label string, url string) FnComponent {
	return NewFn(MenuLi(label)).WithEvents(func(ctx context.Context) FnComponent {
		switch url {
		case "/main":
			return NewFn(HTML(`<div>main</div>`)).SwapTargetInner(Content)
		case "/projects":
			return NewFn(HTML(`<div>projects</div>`)).SwapTargetInner(Content)
		case "/contact":
			return NewFn(HTML(`<div>contact</div>`)).SwapTargetInner(Content)
		}
		return NewFn(HTML(`<div>404</div>`))
	}, OnClick)
}
