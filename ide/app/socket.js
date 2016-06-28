const WebSocket = window.WebSocket,
      HEARTBEAT_INTERVAL = 25000;

class Socket {
  constructor(url) {
    this.url = url;
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
      return this._ws.close(...arguments);
    }
    return;
  }

  send(data) {
    if (this._ws) {
      return this._ws.send(data);
    }
    return;
  }

  get readyState() {
    if (this._ws) {
      return this._ws.readyState;
    }
    return WebSocket.CLOSED;
  }

  _onopen(e) {
    this._heartbeatTimer = window.setInterval(() => {
      this.send('h');
    }, HEARTBEAT_INTERVAL);

    if (typeof this.onopen === 'function') {
      this.onopen(e);
    }
  }

  _onclose(e) {
    if (this._heartbeatTimer) {
      window.clearInterval(this._heartbeatTimer);
      this._heartbeatTimer = null;
    }

    if (typeof this.onclose === 'function') {
      this.onclose(e);
    }
  }

  _onmessage(e) {
    if (typeof this.onmessage === 'function') {
      this.onmessage(e);
    }
  }

  _onerror(e) {
    if (typeof this.onerror === 'function') {
      this.onerror(e);
    }
  }
}

export default Socket;
