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
                // Parse event listeners
                const elem = document.getElementById(dispatch.target_id);
                if (!elem) {
                    dispatch.function = "error";
                    dispatch.message = "element not found: " + dispatch.target_id;
                    this.ws.send(JSON.stringify(dispatch));
                }
                if (dispatch.inner) {
                    elem.innerHTML = dispatch.data;
                }
                else {
                    elem.outerHTML = dispatch.data;
                }
                console.log("ELEMENT:");
                console.log(elem);
                console.log("HTML:");
                console.log(dispatch.data);
                dispatch.event_listeners.forEach((fc) => {
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
                        let event = {
                            function: "event",
                            conn_id: conn_id,
                            target_id: fc.fn_id,
                            inner: false,
                            action: fc.action,
                            method: fc.method,
                            event: fc,
                            event_listeners: [],
                            data: JSON.stringify(elem),
                            message: "event dispatched",
                        };
                        console.log("EVENT:");
                        console.log(event);
                        // Send event to server
                        this.ws.send(JSON.stringify(event));
                    });
                });
            },
            redirect: (e) => {
                let req = {
                    function: "redirect",
                    conn_id: conn_id,
                    target_id: e.target_id,
                    inner: e.inner,
                    action: e.action,
                    method: e.method,
                    event: e.event,
                    data: e.data,
                    event_listeners: e.event_listeners,
                    message: e.message,
                };
                // Send event to server
                this.ws.send(JSON.stringify(req));
            },
            render: (data) => {
                console.log(data);
                // call websocket for render instructions
                this.funs._render(data);
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
        const retry = () => {
            setTimeout(() => {
                if (this.ws.CLOSED) {
                    location.reload();
                    retry();
                }
            }, 1000);
        };
        this.ws.onopen = function () { };
        this.ws.onclose = function () {
            retry();
        };
        this.ws.onerror = function (e) {
            retry();
        };
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
const api = new FnSocketAPI("ws://localhost%s/%s/%s");
