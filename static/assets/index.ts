// let functions: DispatchFunctions;
let conn_id: string | undefined = undefined;
let base_url: string | undefined = undefined;
let verbose = false;

type DispatchFunctions = {
    [key: string]: (data: Dispatch) => Dispatch | void;
};

type FnCustom = {
    function: string;
    data: Object;
};

type FnEventListener = {
    id: string;
    fn_id: string;
    on: string;
    action: string;
    method: string;
    form_data: string;
    data: Object;
};

type FnRender = {
    target_id: string;
    tag: string;
    inner: boolean;
    outer: boolean;
    html: string;
    event_listeners: FnEventListener[];
};

type FnError = {
    message: string;
};

type Dispatch = {
    key: string;
    function: "render" | "redirect" | "event" | "error" | "custom";
    id: string;
    action: string;
    handler_id: string;
    conn_id: string;
    label: string;
    event: FnEventListener;
    render: FnRender;
    error: FnError;
    custom: FnCustom;
};

class Socket {
    private ws: WebSocket | null = null;
    private addr: string | undefined = undefined;
    private key: string | undefined = undefined;

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
        let key = "";
        try {
            this.ws = new WebSocket(addr);
        } catch {
            throw new Error("ws: failed to connect to server...");
        }

        this.ws.onopen = function () {};
        this.ws.onclose = function () {};
        this.ws.onerror = function (e) {};

        this.ws.onmessage = function (event) {
            let d = JSON.parse(event.data) as Dispatch;
            if(d.key != key) {
                throw new Error("ws: invalid key...");
            }
            console.log(d);
            api.Process(this, d);
        };
    }
}

class API {
    private ws: WebSocket | null = null;
    constructor() {
        console.log("FunctionAPI: initialized...");
    }

    public Process(ws: WebSocket, d: Dispatch) {
        if (!this.ws) {
            this.ws = ws;
        }
        if(d.function == "custom") {
            let res = window[d.custom.function](d.custom.data);
            if (res) this.Dispatch(res);
            return;
        }
        let res = this.funs[d.function](d);
        if (res) this.Dispatch(res);
    }

    private Dispatch = (dispatch: Dispatch) => {
        if (!this.ws) {
            throw new Error("ws: not connected to server...");
        }
        this.ws.send(JSON.stringify(dispatch));
    };

    private funs: DispatchFunctions = {
        render: (dispatch: Dispatch) => {
            let elem: HTMLElement | null = null;
            let body: boolean = false;
            let { render } = dispatch;
            // Parse event listeners
            if (render.target_id != "") {
                elem = document.getElementById(render.target_id);
            } else {
                elem = document.body;
                body = true;
            }
            if (!elem) {
                return this.utils.Error(dispatch, "element not found");
            }
            if (render.inner) {
                elem.innerHTML = render.html;
            } else if (render.outer && !body) {
                elem.outerHTML = render.html;
            } else {
                console.log(elem);
                // create element from html
                const div = document.createElement("div");
                div.innerHTML = render.html;
                elem.appendChild(div);
            }
            if (!render.event_listeners) return;
        },
    };

    private utils = {
        // Element selectors
        parseFormData: (ev: Event, dispatch: Dispatch) => {
            const form = ev.target as HTMLFormElement;
            const formData = new FormData(form);
            dispatch.event.data = Object.fromEntries(formData.entries());
            return dispatch;
        },
        getElementById: (id: string): HTMLElement =>
            document.getElementById(id),
        getElementByTagName: (tag: string) =>
            document.getElementsByTagName(tag),
        getElementByClassName: (className: string) =>
            document.getElementsByClassName(className),
        getElementByAttribute: (attribute: string) =>
            document.querySelectorAll(`[${attribute}]`),
        // Replace elements
        replaceElementOuter: (elem: HTMLElement, html: string) => {
            elem.outerHTML = html;
        },
        replaceElementInner: (elem: HTMLElement, html: string) => {
            elem.innerHTML = html;
        },
        addEventListeners: (dispatch: Dispatch) => {
            // Event listeners
            dispatch.render.event_listeners.forEach(
                (listener: FnEventListener) => {
                    const elem = document.getElementById(listener.fn_id);
                    if (!elem) {
                        console.log("element not found: " + listener.fn_id);
                        return;
                    }
                    if (!elem.firstChild) {
                        console.log("element has no child: " + listener.fn_id);
                        return;
                    }
                    elem.firstChild.addEventListener(listener.on, (ev) => {
                        ev.preventDefault();
                        console.log("event: " + listener.on);
                        let dis: Dispatch = { ...dispatch, event: listener };
                        switch (listener.on) {
                            case "submit":
                                dis = this.utils.parseFormData(ev, dispatch);
                                break;
                            default:
                                break;
                        }
                        console.log("EVENT:");
                        console.log(dis);
                        this.Dispatch(dis);
                    });
                }
            );
        },
        Error: (dispatch: Dispatch, message: string): Dispatch => {
            dispatch.function = "error";
            dispatch.error.message =
                message + "; target_id: " + dispatch.render.target_id;
            return dispatch;
        },
    };
}

new Socket("ws://localhost%s%s");
const api = new API();
