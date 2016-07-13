import util from './util';

const WebSocket = window.WebSocket,
      HEARTBEAT_INTERVAL = 25000;

let nextMsgId = 0;

class Socket {
  constructor(url) {
    this.url = url;
    this._ws = null;
    this._heartbeatTimer = null;
    this._intentionalClose = false;
    this._retries = 0;
    this._reconnectTimer = null;
    this._id = null;
    this._callbacks = {};
  }

  open() {
    if (this._ws && this._ws.readyState !== WebSocket.CLOSED) {
      return;
    }

    const ws = this._ws = new WebSocket(this.url);
    ws.onopen = (e) => this._onopen(e);
    ws.onclose = (e) => this._onclose(e);
    ws.onmessage = (e) => this._onmessage(e);
    ws.onerror = (e) => this._onerror(e);
  }

  close() {
    if (this._ws) {
      this._intentionalClose = true;
      return this._ws.close(...arguments);
    }
    return;
  }

  send(msgName, data, callback) {
    if (this._ws) {
      const msgId = nextMsgId++;
      if (typeof callback === "function") {
        this._callbacks[msgId] = callback;
      }
      this._ws.send(`m|${msgName}|${msgId}|${data || ''}`);
      return;
    }
    return;
  }

  get id() {
    return this._id;
  }

  get readyState() {
    if (this._ws) {
      const rawState = this._ws.readyState;
      if (rawState === WebSocket.OPEN && !this._id) {
        return WebSocket.CONNECTING;
      }
      return rawState;
    }
    return WebSocket.CLOSED;
  }

  _reconnect() {
    let delay = 5000;
    if (this._retries <= 1) {
      delay = 1000;
    } else if (this._retries <= 5) {
      delay = 2000;
    } else if (this._retries <= 10) {
      delay = 3000;
    }

    window.clearTimeout(this._reconnectTimer);
    this._reconnectTimer = window.setTimeout(() => {
      this.open();
    }, delay);

    this._retries++;
  }

  _onopen() {
    this._retries = 0;
    this._heartbeatTimer = window.setInterval(() => {
      if (this._ws) {
        return this._ws.send('h');
      }
    }, HEARTBEAT_INTERVAL);
  }

  _onclose(e) {
    window.clearInterval(this._heartbeatTimer);
    this._heartbeatTimer = null;
    this._id = null;
    const ws = this._ws;
    ws.onopen = ws.onclose = ws.onmessage = ws.onerror = null;
    this._ws = null;

    if (typeof this.onclose === 'function') {
      this.onclose(e);
    }

    if (!this._intentionalClose) {
      this._reconnect();
    }
    this._intentionalClose = false;
  }

  _onmessage(e) {
    const msg = e.data;
    if (msg) {
      var i = msg.indexOf("|");
      if (i !== -1) {
        const [hdr, body] = util.splitN(msg, '|', 2);

        if (hdr === 'reg') {
          this._id = body;

          if (typeof this.onopen === 'function') {
            this.onopen();
          }
          return;
        }

        if (hdr === 'm') {
          // message format: m|msgName|msgID?|data

          const [msgName, msgId, data] = util.splitN(body, '|', 3);

          if (typeof this.onmessage === 'function') {
            this.onmessage(msgName, msgId, data);
          }

          const callback = this._callbacks[msgId];
          if (typeof callback === 'function') {
            callback(msgName, data, () => {
              this._callbacks[msgId] = null;
            });
          }
        }
      }
    }
  }

  _onerror(e) {
    if (typeof this.onerror === 'function') {
      this.onerror(e);
    }
  }
}

export default Socket;
