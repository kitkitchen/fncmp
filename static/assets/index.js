let functions;
let _funs;
let conn_id = undefined;
let base_url = undefined;
let verbose = false;
class FnSocketAPI {
    constructor(addr) {
        this.ws = null;
        this.addr = undefined;
        this.funs = {
            _render: (dispatch) => {
                let elem = null;
                let body = false;
                // Parse event listeners
                if (dispatch.target_id != "") {
                    elem = document.getElementById(dispatch.target_id);
                }
                else {
                    elem = document.body;
                    body = true;
                }
                if (!elem) {
                    dispatch.function = "error";
                    dispatch.message = "element not found: " + dispatch.target_id;
                    this.ws.send(JSON.stringify(dispatch));
                }
                if (dispatch.inner) {
                    elem.innerHTML = dispatch.html;
                }
                else if (dispatch.outer && !body) {
                    elem.outerHTML = dispatch.html;
                }
                else {
                    console.log(elem);
                    // create element from html
                    const div = document.createElement("div");
                    div.innerHTML = dispatch.html;
                    elem.appendChild(div);
                }
                console.log("ELEMENT:");
                console.log(elem);
                console.log("HTML:");
                console.log(dispatch.html);
                if (!dispatch.render.event_listeners)
                    return;
                dispatch.render.event_listeners.forEach((fc) => {
                    const elem = document.getElementById(fc.fn_id);
                    if (!elem) {
                        return;
                    }
                    elem.addEventListener(fc.on, (ev) => {
                        ev.preventDefault();
                        console.log("event: " + fc.on);
                        let data;
                        if (fc.on == "submit") {
                            const form = ev.target;
                            const formData = new FormData(form);
                            data = Object.fromEntries(formData.entries());
                            console.log(data);
                        }
                        let event = Object.assign(Object.assign({}, dispatch), { function: "event", target_id: fc.fn_id, inner: false, action: fc.action, method: fc.method, event: fc, event_listeners: [], form_data: JSON.stringify(data), html: "", message: "event dispatched" });
                        console.log("EVENT:");
                        console.log(event);
                        // Send event to server
                        this.ws.send(JSON.stringify(event));
                    });
                });
            },
            redirect: (e) => {
                // redirect to url:
                if (e.action == "") {
                    return;
                }
                window.location.href = e.action;
                // let req: Dispatch = {
                //     function: "redirect",
                //     ...e
                // };
                // // Send event to server
                // this.ws.send(JSON.stringify(req));
            },
            render: (data) => {
                console.log(data);
                // call websocket for render instructions
                this.funs._render(data);
            },
            replace_element: (data) => {
                const elem = document.getElementById(data.target_id);
                if (!elem) {
                    return;
                }
                elem.outerHTML = data.html;
            },
            render_tag: (data) => {
                const elem = document.getElementsByTagName("body")[0];
                if (!elem) {
                    return;
                }
                elem.innerHTML = data.html;
            },
        };
        if (this.addr) {
            throw new Error("ws: already connected to server...");
        }
        if (!addr) {
            throw new Error("ws: no address provided...");
        }
        this.addr = addr;
        _funs = this.funs;
        this.connect(addr);
    }
    connect(addr) {
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
            console.log(d);
            if (!d) {
                throw new Error("ws: no data received in request...");
            }
            if (d.function == "_connect") {
                // Reject multiple connections
                if (conn_id) {
                    return;
                }
                // Initialize connection
                conn_id = d.conn_id;
            }
            else {
                if (!d.function) {
                    console.log("ws: error no function provided in request...");
                    return;
                }
                // Parse function execution
                _funs[d.function](d);
            }
        };
    }
}
// utils for dom manipulation
const utils = {
    // Element selectors
    getFormData: (form) => {
        const formData = new FormData(form);
        return Object.fromEntries(formData.entries());
    },
    getElementById: (id) => document.getElementById(id),
    getElementByTagName: (tag) => document.getElementsByTagName(tag),
    getElementByClassName: (className) => document.getElementsByClassName(className),
    getElementByAttribute: (attribute) => document.querySelectorAll(`[${attribute}]`),
    // Replace elements
    replaceElementOuter: (elem, html) => {
        elem.outerHTML = html;
    },
    replaceElementInner: (elem, html) => {
        elem.innerHTML = html;
    },
    // Event listeners
    addEventListeners: (data) => {
        data.listeners.forEach((fc) => {
            const elem = document.getElementById(fc.fn_id);
            if (!elem) {
                return;
            }
            elem.addEventListener(fc.on, (ev) => {
                ev.preventDefault();
                console.log("event: " + fc.on);
                let data;
                if (fc.on == "submit") {
                    const form = ev.target;
                    const formData = new FormData(form);
                    data = Object.fromEntries(formData.entries());
                    console.log(data);
                }
                let event = Object.assign(Object.assign({}, data.dispatch), { function: "event", target_id: fc.fn_id, inner: false, action: fc.action, method: fc.method, event: fc, event_listeners: [], form_data: JSON.stringify(data), html: "", message: "event dispatched" });
                console.log("EVENT:");
                console.log(event);
                // Send event to server
                this.ws.send(JSON.stringify(event));
            });
        });
    },
};
// Easy access to API
new FnSocketAPI("ws://localhost%s%s");
// const api = new FnSocketAPI("ws://localhost%s%s");
