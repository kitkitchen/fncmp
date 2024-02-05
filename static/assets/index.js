// let functions: DispatchFunctions;
let conn_id = undefined;
let base_url = undefined;
let verbose = false;
class Socket {
    constructor(addr) {
        this.ws = null;
        this.addr = undefined;
        this.key = undefined;
        if (this.addr) {
            throw new Error("ws: already connected to server...");
        }
        if (!addr) {
            throw new Error("ws: no address provided...");
        }
        this.addr = addr;
        this.connect(addr);
    }
    connect(addr) {
        // let key = "";
        try {
            this.ws = new WebSocket(addr);
        }
        catch (_a) {
            throw new Error("ws: failed to connect to server...");
        }
        this.ws.onopen = function () { };
        this.ws.onclose = function () { };
        this.ws.onerror = function (e) { };
        this.ws.onmessage = function (event) {
            let d = JSON.parse(event.data);
            // if(d.key != key) {
            //     throw new Error("ws: invalid key...");
            // }
            api.Process(this, d);
        };
    }
}
class API {
    constructor() {
        this.ws = null;
        this.Dispatch = (data) => {
            if (!data)
                return;
            if (!this.ws) {
                throw new Error("ws: not connected to server...");
            }
            this.ws.send(JSON.stringify(data));
        };
        this.funs = {
            render: (d) => {
                console.log("RENDER:");
                console.log(d);
                let elem = null;
                const parsed = new DOMParser().parseFromString(d.render.html, "text/html").firstChild;
                const html = parsed.getElementsByTagName("body")[0].innerHTML;
                // Select element to render to
                if (d.render.tag != "") {
                    elem = this.utils.getElementsByTagName(d.render.tag)[0];
                    if (!elem) {
                        return this.Error(d, "element with tag not found: " + d.render.tag);
                    }
                }
                else if (d.render.target_id != "") {
                    elem = this.utils.getElementById(d.render.target_id);
                    if (!elem) {
                        return this.Error(d, "element with target_id not found: " +
                            d.render.target_id);
                    }
                }
                else {
                    return this.Error(d, "no target or tag specified");
                }
                // Render element
                if (d.render.inner) {
                    elem.innerHTML = html;
                }
                if (d.render.outer) {
                    elem.outerHTML = html;
                }
                if (d.render.append) {
                    const div = document.createElement("div");
                    div.innerHTML = html;
                    elem.appendChild(div);
                }
                if (d.render.prepend) {
                    elem.prepend(html);
                }
                this.Dispatch(this.utils.addEventListeners(d));
                return;
            },
        };
        this.utils = {
            // Element selectors
            parseFormData: (ev, d) => {
                const form = ev.target;
                const formData = new FormData(form);
                d.event.data = Object.fromEntries(formData.entries());
                return d;
            },
            getElementById: (id) => document.getElementById(id),
            getElementsByTagName: (tag) => document.getElementsByTagName(tag),
            getElementByClassName: (className) => document.getElementsByClassName(className),
            getElementByAttribute: (attribute) => document.querySelectorAll(`[${attribute}]`),
            trackTouch: (elem) => {
                elem.addEventListener("touchstart", (ev) => {
                    //TODO: event object comes back as touch specific
                    ev.preventDefault();
                    elem.classList.add("touch");
                    // send data to api
                });
                elem.addEventListener("touchend", (ev) => {
                    elem.classList.remove("touch");
                });
            },
            addEventListeners: (d) => {
                if (!d.render.event_listeners)
                    return;
                // Event listeners
                d.render.event_listeners.forEach((listener) => {
                    console.log("listener: " + listener);
                    let elem = document.getElementById(listener.target_id);
                    if (!elem) {
                        console.log("elem not found");
                        this.Error(d, "element not found");
                        return;
                    }
                    if (elem.firstChild) {
                        console.log("elem has children");
                        elem = elem.firstChild;
                    }
                    else {
                        console.log("elem has no children");
                    }
                    console.log("elem with listener: " + elem);
                    elem.addEventListener(listener.on, (ev) => {
                        ev.preventDefault();
                        console.log("event: " + listener.on);
                        d.function = "event";
                        d.event = listener;
                        switch (listener.on) {
                            case "submit":
                                d = this.utils.parseFormData(ev, d);
                                break;
                            default:
                                break;
                        }
                        console.log("EVENT:");
                        console.log(d);
                        this.Dispatch(d);
                    });
                });
            },
        };
        this.Error = (d, message) => {
            d.function = "error";
            d.error.message = message;
            this.Dispatch(d);
        };
        console.log("API: initialized...");
    }
    // Process is the entry point for all api calls via the websocket
    Process(ws, d) {
        if (!this.ws) {
            this.ws = ws;
        }
        switch (d.function) {
            case "redirect":
                window.location.href = d.redirect.url;
                return;
            case "custom":
                this.Dispatch(window[d.custom.function](d.custom.data));
                return;
            case "render":
                this.Dispatch(this.funs.render(d));
            default:
                // this.Error(d, "invalid function: " + d.function);
                break;
        }
    }
}
new Socket("ws://localhost%s%s");
const api = new API();
