import './css/app.sass';

var conn = new WebSocket(`${location.protocol.replace("http", "ws")}//${location.host}/socket`);
conn.onclose = function(/*evt*/) {
  console.log('connection closed');
};
conn.onmessage = function(evt) {
  console.log('message:', evt.data);
};
