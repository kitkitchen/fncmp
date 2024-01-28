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

type FnEventListener = {
    id: string;
    fn_id: string;
    on: string;
    action: string;
    method: string;
    data: Object;
};

type Dispatch = {
    function: "render" | "redirect" | "event" | "error" | "_connect";
    conn_id: string;
    target_id: string;
    inner: boolean;
    action: string;
    method: string;
    event: FnEventListener;
    event_listeners: FnEventListener[];
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

        const retry = () => {
            setTimeout(() => {
                if (this.ws.CLOSED) {
                    location.reload();
                    retry();
                }
            }, 1000);
        };

        this.ws.onopen = function () {};
        this.ws.onclose = function () {
            retry();
        };
        this.ws.onerror = function (e) {
            retry();
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
        _render: (dispatch: Dispatch) => {
            // Parse event listeners
            const elem = document.getElementById(dispatch.target_id);
            if (!elem) {
                dispatch.function = "error";
                dispatch.message = "element not found: " + dispatch.target_id;
                this.ws.send(JSON.stringify(dispatch));
            }
            if (dispatch.inner) {
                elem.innerHTML = dispatch.data;
            } else {
                elem.outerHTML = dispatch.data;
            }
            console.log("ELEMENT:");
            console.log(elem);
            console.log("HTML:");
            console.log(dispatch.data);

            dispatch.event_listeners.forEach((fc: FnEventListener) => {
                const elem = document.getElementById(fc.fn_id);
                if (!elem) {
                    return;
                }
                elem.addEventListener(fc.on, (ev) => {
                    ev.preventDefault();
                    console.log("event: " + fc.on);
                    let data: any;
                    if (fc.on == "submit") {
                        const form = ev.target as HTMLFormElement;
                        const formData = new FormData(form);
                        data = Object.fromEntries(formData.entries());
                        console.log(data);
                    }

                    let event: Dispatch = {
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
                    console.log("EVENT:")
                    console.log(event);
                    // Send event to server
                    this.ws.send(JSON.stringify(event));
                });
            });
        },
        redirect: (e: Dispatch) => {
            let req: Dispatch = {
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
        render: (data: Dispatch) => {
            console.log(data);
            // call websocket for render instructions
            this.funs._render(data);
        },
    };
}

const api = new FnSocketAPI("ws://localhost%s/%s/%s");
