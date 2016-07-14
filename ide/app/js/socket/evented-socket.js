import Socket from './socket';

App.Socket = Em.Object.extend(Em.Evented, {
  started: false,
  startedFirstTime: false,

  init: function() {
    const token = App.localStorage.get('token');
    const url = `${location.protocol.replace("http", "ws")}//${location.host}/socket?token=${token}`;
    const conn = this.conn = new Socket(url);

    conn.onopen = () => this.didOpen();
    conn.onclose = (ev) => this.didClose(ev);
    conn.onmessage = (msgName, msgId, data) => this.didReceiveMessage(msgName, msgId, data);
  },

  connect: function() {
    this.conn.open();
  },

  disconnect: function() {
    this.set('started', false);
    this.set('startedFirstTime', false);
    this.conn.close();
  },

  destroy: function() {
    this.conn = undefined;
  },

  didOpen: function() {
    this.set('started', true);
    if (!this.get('startedFirstTime'))
      this.set('startedFirstTime', true);
    console.log('connection open', this.conn.id);
  },

  didClose: function(ev) {
    console.log('connection close');
  },

  didReceiveMessage: function(msgName, msgId, data) {
    console.log('messageName:', msgName, 'msgId:', msgId, 'data:', data);
  },

  fetchFileList: function(path, cb) {
    const data = JSON.stringify({ path });
    this.conn.send("fs.ls", data, function(msgName, data, done) {
      let message = null;
      try {
        message = JSON.parse(data);
      } catch(err) {
        console.log('JSON parse error', err);
        return;
      }
      cb(message.error, message.files);
      done();
    });
  },

  fetchFileContent: function(path, cb) {
    const data = JSON.stringify({ path });
    this.conn.send("fs.read", data, function(msgName, data, done) {
      let message = null;
      try {
        message = JSON.parse(data);
      } catch(err) {
        console.log('JSON parse error', err);
        return;
      }
      cb(message.error, message.type, message.content);
      done();
    });
  },

  unwatchDir: function(path) {
  },

  unwatchFile: function(path) {
  },

  writeFile: function(path, content, cb) {
  },

  writeNewFile: function(path, content, cb) {
  },

  makeDirectory: function(path, cb) {
  },

  deleteFile: function(path, cb) {
  },

  deleteDir: function(path, cb) {
  },

  rename: function(path, newPath, cb) {
  },

  fileFind: function(term, root, cb) {
  },

  search: function(id, term, location, ignoreCase, useRegexp, filePattern, cb) {
  }
});
