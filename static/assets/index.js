var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
let functions;
let conn_id = undefined;
let base_url = undefined;
let verbose = false;
class FunSocketAPI {
    constructor(addr) {
        this.ws = null;
        this.addr = undefined;
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
        try {
            this.ws = new WebSocket(addr);
        }
        catch (_a) {
            throw new Error("ws: failed to connect to server...");
        }
        this.ws.onopen = function () { };
        this.ws.onclose = function () { };
        this.ws.onerror = function (e) {
            throw new Error("ws: " + e);
        };
        this.ws.onmessage = function (event) {
            const fr = JSON.parse(event.data);
            if (!fr) {
                throw new Error("ws: no data received in request...");
            }
            // Initialize connection with fc_config.json
            if (fr.function == "_connect") {
                if (conn_id) {
                    return;
                }
                const conn = fr;
                conn_id = conn.data.conn_id;
                base_url = conn.data.base_url;
                verbose = conn.data.verbose;
                this.send(JSON.stringify({
                    error: false,
                    message: "connected",
                }));
                return;
            }
            if (!conn_id) {
                return;
            }
            // Parse function execution
            const func = funs[fr.function];
            const res = func(fr.data);
            this.send(JSON.stringify(res));
        };
    }
}
const api = new FunSocketAPI("ws://localhost:8080/funs");
// todo: make all of these functions private and configure a key to access them
const funs = {
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
        const fcs = elem.querySelectorAll("[fc]");
        fcs.forEach((elem) => {
            if (!elem) {
                return;
            }
            const _fc = elem.getAttribute("fc");
            // Parse attribute
            const fc = JSON.parse(_fc);
            elem.addEventListener(fc.on, (ev) => {
                ev.preventDefault();
                console.log("event: " + fc.on);
                let data;
                if (fc.on == "submit") {
                    const form = ev.target;
                    const formData = new FormData(form);
                    data = Object.fromEntries(formData.entries());
                }
                // Send event to server
                funs.event({
                    conn_id: conn_id,
                    target_id: elem.id,
                    event: fc,
                    data: data,
                });
            });
        });
        return {
            error: false,
            message: "success",
        };
    },
    event: (e) => __awaiter(this, void 0, void 0, function* () {
        let headers = {};
        console.log("event: " + e.toString());
        const result = yield fetcher(e.event.action, e.event.method, e)
            // This parses html from socket message
            .then((res) => {
            headers = res.headers;
            return res.text();
        })
            .then((text) => text)
            .catch((err) => {
            return {
                error: true,
                message: err,
            };
        });
        funs._render({
            id: e.target_id,
            inner: false,
            html: result.toString(),
        });
    }),
    render: (e) => __awaiter(this, void 0, void 0, function* () {
        // make sure this is right
        const result = yield fetcher(e.action, e.method, e)
            .then((res) => res.text().then((text) => text))
            .catch((err) => {
            return {
                error: true,
                message: err,
            };
        });
        funs._render({
            id: e.target_id,
            inner: e.inner,
            html: result,
        });
    }),
};
// Utility functions
const fetcher = (url, method, data, headers) => {
    console.log("base_url: " + base_url);
    return fetch(base_url || "http://localhost:8080" + url, {
        method: method,
        headers: Object.assign(Object.assign({}, headers), { 
            // "Content-Type": "*",
            // "Access-Control-Allow-Origin": "*",
            // "Access-Control-Allow-Headers": "*",
            Conn: conn_id || "" }),
        body: JSON.stringify(data),
    });
};
