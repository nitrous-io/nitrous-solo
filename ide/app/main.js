import './css/app.sass';
import Socket from './socket';

window.conn = new Socket(`${location.protocol.replace("http", "ws")}//${location.host}/socket`);

conn.onopen = function() {
  console.log('connection open', conn.id);

  conn.send("ping", null, function(msgName, data, done) {
    console.log(data);
    done(); // call done when no more responses are expected
  });

  conn.send("ping", "hello!", function(msgName, data, done) {
    console.log(data);
    done();
  });
};

conn.onclose = function(/*evt*/) {
  console.log('connection closed');
};

conn.onmessage = function(msgName, msgId, data) {
  console.log('messageName:', msgName, 'msgId:', msgId, 'data:', data);
};

conn.open();
