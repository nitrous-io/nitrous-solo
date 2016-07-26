import Socket from './socket';

App.Socket = Em.Object.extend(Em.Evented, {
  started: false,
  startedFirstTime: false,

  init() {
    const token = App.localStorage.get('token');
    const url = `${location.protocol.replace("http", "ws")}//${location.host}/socket?token=${token}`;
    const conn = this.conn = new Socket(url);

    conn.onopen = () => this.didOpen();
    conn.onclose = (ev) => this.didClose(ev);
    conn.onmessage = (msgName, msgId, data) => this.didReceiveMessage(msgName, msgId, data);
  },

  connect() {
    this.conn.open();
  },

  disconnect() {
    this.conn.close();
  },

  destroy() {
    this.conn = undefined;
  },

  didOpen() {
    this.set('started', true);
    if (!this.get('startedFirstTime')) {
      this.set('startedFirstTime', true);
    }
    this.trigger('open');
  },

  didClose(ev) {
    this.set('started', false);
    this.trigger('close');
  },

  didReceiveMessage(msgName, msgId, data) {
  },

  fetchFileList(path, cb) {
    const data = JSON.stringify({ path });
    this.conn.send("fs.ls", data, function(msgName, data, done) {
      let message = null;
      try {
        message = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }
      cb(message.error, message.files);
      done();
    });
  },

  fetchFileContent(path, cb) {
    const data = JSON.stringify({ path });
    this.conn.send("fs.read", data, function(msgName, data, done) {
      let message = null;
      try {
        message = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }
      if (message.type === 'text') {
        let content = message.content;
        try {
          content = atob(content);
        } catch(err) {
          done();
          return;
        }
        message.content = decodeUTF8(content);
      }
      cb(message.error, message.type, message.content);
      done();
    });
  },

  watch(path, onStart, onChange, onStop) {
    this.conn.send('fs.watch', JSON.stringify({ path }), function(msgName, data, done) {
      let message;
      try {
        message = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }
      switch (message.evt) {
        case 'started':
          onStart(message.id);
          break;

        case 'stopped':
          done();
          onStop(message.id);
          break;

        default:
          onChange(message.id, message.evt, message.path);
      }
    });
  },

  unwatch(id) {
    this.conn.send('fs.unwatch', JSON.stringify({ id }));
  },

  watchFile() {},

  unwatchFile() {},

  writeFile(path, overwrite, content, cb) {
    this.conn.send('fs.write', JSON.stringify({
      path,
      overwrite,
      content: btoa(encodeUTF8(content))
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      cb(res.error, path);
      done();
    });
  },

  makeDirectory(path, cb) {
    this.conn.send('sh.run', JSON.stringify({
      cmd: 'mkdir',
      args: [
        '-m', '0755',
        '-p', path
      ]
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      cb(res.error);
      done();
    });
  },

  deleteFile(path, cb) {
    this.conn.send('sh.run', JSON.stringify({
      cmd: "rm",
      args: [path]
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      cb(res.error);
      done();
    });
  },

  deleteDir(path, cb) {
    // prevent accidental deletion of sensitive paths
    if (!path || path === "/" || path.match(/^\/(home|Users)(\/[^\/]+)?\/?$/) || path.match(/^\/root\/?$/)) {
      cb(`Cowardly refusing to delete "${path}".`)
      return;
    }

    this.conn.send('sh.run', JSON.stringify({
      cmd: "rm",
      args: ["-r", path]
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      cb(res.error);
      done();
    });
  },

  rename(path, newPath, cb) {
    this.conn.send('sh.run', JSON.stringify({
      cmd: "mv",
      args: [path, newPath]
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return
      }

      cb(res.error);
      done();
    });
  },

  fileFind(term, root, cb) {
  },

  search(term, path, ignoreCase, useRegexp, filePattern, cb) {
    this.conn.send('fs.findInFiles', JSON.stringify({
      term,
      path,
      ignore_case: ignoreCase,
      use_regexp: useRegexp,
      file_pattern: filePattern
    }), function(magName, data, done) {
      let res;
      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      try {
        res.results = (res.results || []).map(function(result, i) {
          return Object.assign({}, result, {
            matches: result.matches.map(function(match, j) {
              match.content = decodeUTF8(atob(match.content));
              match.highlighted_content = decodeUTF8(atob(match.highlighted_content));
              return match;
            })
          });
        });
      } catch(err) {
        done();
        return
      }

      cb(res);
      done();
    });
  },

  // start an interactive terminal
  shellStart(cmd, args, env, wd, runAs, cols, rows, onAttach, onData, onError, onExit) {
    const data = JSON.stringify({
      cmd, args, env, wd, run_as: runAs, cols, rows
    });
    this.conn.send('sh.start', data, function(msgName, data, done) {
      let j;
      try {
        j = JSON.parse(data)
      } catch(err) {
        onError(err.toString());
        done();
        return;
      }

      if (Array.isArray(j)) {
        const kind = j[0],
              data = j[1];

        if (typeof data === 'string') {
          if (kind === 'd') {
            let d;
            try {
              d = atob(data);
            } catch(err) {}
            if (d) {
              onData(d);
            }
          } else if (kind === 'exit') {
            onExit(data);
            done();
            return;
          }
        }
      }

      if (j.error) {
        onError(j.errorDescription || j.error)
        done();
        return;
      }

      if (j.subscription_id && j.shell_session_id) {
        onAttach(j.subscription_id, j.shell_session_id);
        // attached and expecting more replies - should not call done()
        return;
      }
    });
  },

  shellAttach(shellSessionID, cols, rows, onAttach, onData, onError, onExit) {
    const data = JSON.stringify({
      shell_session_id: shellSessionID,
      cols, rows
    });
    this.conn.send('sh.attach', data, function(msgName, data, done) {
      let j;
      try {
        j = JSON.parse(data)
      } catch(err) {
        onError(err.toString());
        done();
        return;
      }

      if (Array.isArray(j)) {
        const kind = j[0],
              data = j[1];

        if (typeof data === 'string') {
          if (kind === 'd') {
            let d;
            try {
              d = atob(data);
            } catch(err) {}
            if (d) {
              onData(d);
            }
          } else if (kind === 'exit') {
            onExit(data);
            done();
            return;
          }
        }
      }

      if (j.error) {
        onError(j.errorDescription || j.error)
        done();
        return;
      }

      if (j.subscription_id) {
        onAttach(j.subscription_id, j.scrollback);
        // attached and expecting more replies - should not call done()
        return;
      }
    });
  },

  // input for interactive terminal
  shellInput(shellSessionID, data) {
    if (!shellSessionID || !data) {
      return;
    }
    this.conn.send('sh.i', JSON.stringify([
      shellSessionID,
      btoa(encodeUTF8(data))
    ]));
  },

  // resize for interactive terminal
  shellResize(shellSessionID, cols, rows) {
    if (!shellSessionID || cols == null || rows == null) {
      return;
    }

    this.conn.send('sh.r', JSON.stringify([
      shellSessionID,
      cols,
      rows
    ]));
  },

  shellKill(shellSessionID) {
    if (!shellSessionID) {
      return;
    }

    this.conn.send('sh.kill', JSON.stringify({
      shell_session_id: shellSessionID
    }));
  },

  shellList(cb) {
    this.conn.send('sh.list', null, function(msgName, data, done) {
      let list;
      try {
        list = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }
      cb(list);
      done();
    });
  },
});

function encodeUTF8(s) {
  try {
    return unescape(encodeURIComponent(s));
  } catch(err) {
    return s;
  }
}

function decodeUTF8(s) {
  try {
    return decodeURIComponent(escape(s));
  } catch(err) {
    return s;
  }
}
