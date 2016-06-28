import './css/app.sass';
import Socket from './socket';

var conn = new Socket(`${location.protocol.replace("http", "ws")}//${location.host}/socket`);

conn.onopen = function(/*evt*/) {
  console.log('connection open', conn.id);
};

conn.onclose = function(/*evt*/) {
  console.log('connection closed');
};

conn.onmessage = function(evt) {
  console.log('message:', evt.data);
};

conn.open();
