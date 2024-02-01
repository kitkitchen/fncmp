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
    append: boolean;
    html: string;
    event_listeners: FnEventListener[];
};

type FnRedirect = {
    url: string;
};

type FnError = {
    message: string;
};

type Dispatch = {
    function: "render" | "redirect" | "event" | "error" | "custom";
    id: string;
    key: string;
    conn_id: string;
    handler_id: string;
    action: string;
    label: string;
    event: FnEventListener;
    render: FnRender;
    redirect: FnRedirect;
    custom: FnCustom;
    error: FnError;
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
        // let key = "";
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
            // if(d.key != key) {
            //     throw new Error("ws: invalid key...");
            // }
            console.log(d);
            api.Process(this, d);
        };
    }
}

class API {
    private ws: WebSocket | null = null;
    constructor() {
        console.log("API: initialized...");
    }

    // Process is the entry point for all api calls via the websocket
    public Process(ws: WebSocket, d: Dispatch) {
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
                if (!d.render.event_listeners) return;
                this.Dispatch(this.utils.addEventListeners(d));
                return;
            default:
                this.Error(d, "invalid function: " + d.function);
                break;
        }
    }

    private Dispatch = (data: Dispatch | void) => {
        if (!data) return;
        if (!this.ws) {
            throw new Error("ws: not connected to server...");
        }
        this.ws.send(JSON.stringify(data));
    };

    private funs: DispatchFunctions = {
        render: (d: Dispatch) => {
            let elem: Element | null = null;

            // Select element to render to
            if (d.render.tag != "") {
                elem = this.utils.getElementsByTagName(d.render.tag)[0];
                if (!elem) {
                    return this.Error(
                        d,
                        "element with tag not found: " + d.render.tag
                    );
                }
            } else if (d.render.target_id != "") {
                elem = this.utils.getElementById(d.render.target_id);
                if (!elem) {
                    return this.Error(
                        d,
                        "element with target_id not found: " +
                            d.render.target_id
                    );
                }
            } else {
                return this.Error(d, "no target or tag specified");
            }
            if (!elem) {
                return this.Error(d, "element not found");
            }

            // Render element
            if (d.render.inner) {
                this.utils.replaceElementInner(
                    elem as HTMLElement,
                    d.render.html
                );
                return;
            }
            if (d.render.outer) {
                this.utils.replaceElementOuter(
                    elem as HTMLElement,
                    d.render.html
                );
                return;
            }
            if (d.render.append) {
                this.utils.appendElement(elem as HTMLElement, d.render.html);
                return;
            }
        },
    };

    private utils = {
        // Element selectors
        parseFormData: (ev: Event, d: Dispatch) => {
            const form = ev.target as HTMLFormElement;
            const formData = new FormData(form);
            d.event.data = Object.fromEntries(formData.entries());
            return d;
        },
        getElementById: (id: string): HTMLElement =>
            document.getElementById(id),
        getElementsByTagName: (tag: string) =>
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
        appendElement: (elem: HTMLElement, html: string) => {
            elem.innerHTML += html;
        },
        trackTouch: (elem: HTMLElement) => {
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
        addEventListeners: (d: Dispatch) => {
            // Event listeners
            d.render.event_listeners.forEach((listener: FnEventListener) => {
                let elem = document.getElementById(listener.fn_id);
                if (!elem) {
                    this.Error(d, "element not found");
                    return;
                }
                if (elem.firstChild) {
                    elem = elem.firstChild as HTMLElement;
                }
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

    private Error = (d: Dispatch, message: string) => {
        d.function = "error";
        d.error.message = message;
        this.Dispatch(d);
    };
}

new Socket("ws://localhost%s%s");
const api = new API();
