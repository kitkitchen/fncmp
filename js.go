package main

import "fmt"

func js(port string, path string) string {
	return fmt.Sprintf(`(()=>{var a,f;var c=class{constructor(s){if(this.ws=null,this.addr=void 0,this.funs={_render:e=>{let n=null,r=!1;if(e.target_id!=""?n=document.getElementById(e.target_id):(n=document.body,r=!0),n||(e.function="error",e.message="element not found: "+e.target_id,this.ws.send(JSON.stringify(e))),e.inner)n.innerHTML=e.html;else if(e.outer&&!r)n.outerHTML=e.html;else{console.log(n);let t=document.createElement("div");t.innerHTML=e.html,n.appendChild(t)}console.log("ELEMENT:"),console.log(n),console.log("HTML:"),console.log(e.html),e.event_listeners&&e.event_listeners.forEach(t=>{let o=document.getElementById(t.fn_id);o&&o.addEventListener(t.on,i=>{i.preventDefault(),console.log("event: "+t.on);let l;if(t.on=="submit"){let u=i.target,m=new FormData(u);l=Object.fromEntries(m.entries()),console.log(l)}let d=Object.assign(Object.assign({},e),{function:"event",target_id:t.fn_id,inner:!1,action:t.action,method:t.method,event:t,event_listeners:[],form_data:JSON.stringify(l),html:"",message:"event dispatched"});console.log("EVENT:"),console.log(d),this.ws.send(JSON.stringify(d))})})},redirect:e=>{e.action!=""&&(window.location.href=e.action)},render:e=>{console.log(e),this.funs._render(e)},replaceElement:e=>{let n=document.getElementById(e.target_id);n&&(n.outerHTML=e.html)},addEventListeners:e=>{e.listeners.forEach(n=>{let r=document.getElementById(n.fn_id);r&&r.addEventListener(n.on,t=>{t.preventDefault(),console.log("event: "+n.on);let o;if(n.on=="submit"){let l=t.target,d=new FormData(l);o=Object.fromEntries(d.entries()),console.log(o)}let i=Object.assign(Object.assign({},o.dispatch),{function:"event",target_id:n.fn_id,inner:!1,action:n.action,method:n.method,event:n,event_listeners:[],form_data:JSON.stringify(o),html:"",message:"event dispatched"});console.log("EVENT:"),console.log(i),this.ws.send(JSON.stringify(i))})})}},this.addr)throw new Error("ws: already connected to server...");if(!s)throw new Error("ws: no address provided...");this.addr=s,a=this.funs,this.connect(s)}connect(s){try{this.ws=new WebSocket(s)}catch(e){throw new Error("ws: failed to connect to server...")}this.ws.onopen=function(){},this.ws.onclose=function(){},this.ws.onerror=function(e){},this.ws.onmessage=function(e){let n=JSON.parse(e.data);if(console.log(n),!n)throw new Error("ws: no data received in request...");if(n.function=="_connect"){if(f)return;f=n.conn_id}else{if(!n.function){console.log("ws: error no function provided in request...");return}a[n.function](n)}}}},w=new c("ws://localhost%s%s");})();`, port, path)
}
