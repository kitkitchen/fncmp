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
            _render: (data) => {
                const elem = document.getElementById(data.id);
                if (!elem) {
                    return {
                        error: true,
                        message: "element not found: " + data,
                    };
                }
                elem.innerHTML = data.html;
                // if (data.inner) {
                //     elem.innerHTML = data.html;
                // } else {
                //     elem.outerHTML = data.html;
                // }
                console.log(data.html);
                const elems = elem.querySelectorAll("[fc]");
                console.log(elems);
                elems.forEach((elem) => {
                    console.log("ELEM");
                    console.log(elem);
                    if (!elem) {
                        return;
                    }
                    const _fc = elem.getAttribute("fc");
                    console.log("_FC ATTRIBUTE");
                    console.log(_fc);
                    const array = JSON.parse(_fc);
                    // Parse attributes
                    array.forEach((thisFc) => {
                        const fc = thisFc;
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
                            let req = {
                                function: "event",
                                conn_id: conn_id,
                                target_id: elem.id,
                                inner: false,
                                action: fc.action,
                                method: fc.method,
                                event: fc,
                                data: JSON.stringify(data),
                                message: "event dispatched",
                            };
                            console.log(req);
                            // Send event to server
                            this.ws.send(JSON.stringify(req));
                        });
                    });
                });
            },
            redirect: (e) => {
                let req = {
                    function: "redirect",
                    conn_id: conn_id,
                    target_id: e.target_id,
                    inner: false,
                    action: e.action,
                    method: e.method,
                    event: e.event,
                    data: e.data,
                    message: e.message,
                };
                // Send event to server
                this.ws.send(JSON.stringify(req));
            },
            render: (e) => {
                console.log(e);
                // call websocket for render instructions
                this.funs._render({
                    id: e.target_id,
                    inner: e.inner,
                    html: e.data,
                });
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
