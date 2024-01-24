let functions: Functions;
let conn_id: string | undefined = undefined;
let base_url: string | undefined = undefined;
let verbose = false;

type Functions = {
    [key: string]: (data: any) => any;
};

type FunRequest<T> = {
    function: string;
    data: T;
};

type EventRequest = {
    conn_id: string;
    target_id: string;
    event: {
        on: string;
        action: string;
        method: string;
        id: string;
    };
    data: Object;
};

type RenderTargetRequest = {
    conn_id: string;
    target_id: string;
    inner: boolean;
    action: string;
    method: string;
    data: Object;
};

class FunSocketAPI {
    private ws: WebSocket | null = null;
    private addr: string | undefined = undefined;

    constructor(addr: string) {
        if (this.addr) {
            throw new Error("ws: already connected to server...");
        }
        if (!addr) {
            throw new Error("ws: no address provided...");
        }
        this.addr = addr;
        this.connect(addr);
    }

    private connect(addr: string) {
        try {
            this.ws = new WebSocket(addr);
        } catch {
            throw new Error("ws: failed to connect to server...");
        }
        this.ws.onopen = function () {};
        this.ws.onclose = function () {};
        this.ws.onerror = function (e) {
            throw new Error("ws: " + e);
        };

        this.ws.onmessage = function (event) {
            const fr = JSON.parse(event.data) as FunRequest<any>;
            if (!fr) {
                throw new Error("ws: no data received in request...");
            }
            // Initialize connection with fc_config.json
            if (fr.function == "_connect") {
                if (conn_id) {
                    return;
                }
                const conn = fr as FunRequest<{
                    conn_id: string;
                    base_url: string;
                    verbose: boolean;
                }>;
                conn_id = conn.data.conn_id;
                base_url = conn.data.base_url;
                verbose = conn.data.verbose;
                this.send(
                    JSON.stringify({
                        error: false,
                        message: "connected",
                    })
                );
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
    _render: (data: { id: string; inner: boolean, html: string }) => {
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
        console.log(data.html)
        const fcs = elem.querySelectorAll("[fc]");
        fcs.forEach((elem: Element) => {
            if (!elem) {
                return;
            }
            const _fc = elem.getAttribute("fc");
            // Parse attribute
            const fc = JSON.parse(_fc) as {
                on: string;
                action: string;
                method: string;
                id: string;
            };
            elem.addEventListener(fc.on, (ev) => {
                ev.preventDefault();
                console.log("event: " + fc.on)
                let data: any;
                if (fc.on == "submit") {
                    const form = ev.target as HTMLFormElement;
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
    event: async (e: EventRequest) => {
        let headers: {} = {};
        console.log("event: " + e.toString());
        const result = await fetcher(e.event.action, e.event.method, e)
            // This parses html from socket message
            .then((res) => {
                headers = res.headers;
                return res.text()
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
    },
    render: async (e: RenderTargetRequest) => {
        // make sure this is right
        const result = await fetcher(e.action, e.method, e)
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
            html: result as string,
        });
    },
};
// Utility functions
const fetcher = (url: string, method: string, data: Object, headers?: {}) => {
    console.log("base_url: " + base_url)
    return fetch(base_url || "http://localhost:8080" + url, {
        method: method,
        headers: {
            ...headers,
            // "Content-Type": "*",
            // "Access-Control-Allow-Origin": "*",
            // "Access-Control-Allow-Headers": "*",
            Conn: conn_id || "",
        },
        body: JSON.stringify(data),
    });
};
