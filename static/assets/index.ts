let functions: Functions;
let _funs: Functions;
let conn_id: string | undefined = undefined;
let base_url: string | undefined = undefined;
let verbose = false;

type Functions = {
    [key: string]: (data: any) => any;
};

type FnRequest<T> = {
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

type Dispatch = {
    function: "render" | "redirect" | "event" | "error" | "_connect";
    conn_id: string;
    target_id: string;
    inner: boolean;
    action: string;
    method: string;
    event: {
        on: string;
        action: string;
        method: string;
        id: string;
    };
    data: string;
    message: string;
};

class FnSocketAPI {
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
        _funs = this.funs;
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
            let d = JSON.parse(event.data) as Dispatch;
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
            } else {
                if (!d.function) {
                    console.log("ws: error no function provided in request...");
                    return;
                }
                // Parse function execution
                _funs[d.function](d);
            }
        };
    }

    funs: Functions = {
        _render: (data: { id: string; inner: boolean; html: string }) => {
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
            elems.forEach((elem: Element) => {
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
                array.forEach((thisFc: any) => {
                    const fc = thisFc as {
                        on: string;
                        action: string;
                        method: "POST";
                        id: string;
                    };
                    elem.addEventListener(fc.on, (ev) => {
                        ev.preventDefault();
                        console.log("event: " + fc.on);
                        let data: any;
                        if (fc.on == "submit") {
                            const form = ev.target as HTMLFormElement;
                            const formData = new FormData(form);
                            data = Object.fromEntries(formData.entries());
                            console.log(data)
                        }

                        let req: Dispatch = {
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
                        console.log(req)
                        // Send event to server
                        this.ws.send(JSON.stringify(req));
                    });
                });
            });
        },
        redirect: (e: Dispatch) => {
            let req: Dispatch = {
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
        render: (e: Dispatch) => {
            console.log(e);
            // call websocket for render instructions
            this.funs._render({
                id: e.target_id,
                inner: e.inner,
                html: e.data as string,
            });
        },
    };
}

const api = new FnSocketAPI("ws://localhost%s/%s/%s");
