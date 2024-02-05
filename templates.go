package main

func Modal(content ...Component) Component {
	return HTML(`
	<div id="myModal" class="modal" style="
		position:absolute;
		display:flex;
		flex:1;
		height:100vh;
		width:100vw;
		background-color:rgba(0,0,0,0.1);
		justify-content:center;
		align-items:center;
	">
		<div class="modal-content" style="
			background-color:white;
			padding:20px;
			border-radius:5px;
			box-shadow: 0px 0px 10px 0px rgba(0,0,0,0.75);
		">
			` + RenderHTML(content...) + `
		</div>
	</div>
	`)
}

type Opt func() string

func Type(v string) Opt {
	return func() string {
		return "type='" + v + "'"
	}
}

func Name(v string) Opt {
	return func() string {
		return "name='" + v + "'"
	}
}

func Placeholder(v string) Opt {
	return func() string {
		return "placeholder='" + v + "'"
	}
}

func Value(v string) Opt {
	return func() string {
		return "value='" + v + "'"
	}
}

func Style(v string) Opt {
	return func() string {
		return "style='" + v + "'"
	}
}

func Class(v string) Opt {
	return func() string {
		return "class='" + v + "'"
	}
}

func For(v string) Opt {
	return func() string {
		return "for='" + v + "'"
	}
}
