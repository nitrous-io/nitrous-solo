let lastUpdate = 0;

module.exports = function fileFind(conn, what, where, cb) {
  if ((+new Date() - lastUpdate) > 5000) {
    conn.send('sh.run', JSON.stringify({
      cmd: 'updatedb',
      run_as: 'root'
    }), (msgName, data, done) => {
      done();
      lastUpdate = +new Date();
      locate(conn, what, where, cb);
    });
    return;
  }

  locate(conn, what, where, cb);
}

function locate(conn, what, where, cb) {
  if (where.slice(-1) !== '/') {
    where += '/';
  }

  findCanonicalPath(conn, where, function(path) {
    conn.send('sh.run', JSON.stringify({
      cmd: "locate",
      run_as: 'root',
      args: [
        '--limit', '100',
        '--ignore-case',
        '--regex',
        `^${escapeRegexChars(path)}${convertFileFindTermToRegexString(what)}$`
      ]
    }), function(msgName, data, done) {
      let res;

      try {
        res = JSON.parse(data);
      } catch(err) {
        done();
        return;
      }

      let stdout, stderr;
      try {
        stdout = decodeUTF8(atob(res.stdout));
        stderr = decodeUTF8(atob(res.stderr));
      } catch(err) {
        done();
        return;
      }

      let list = stdout.split('\n');
      list.pop();

      cb(stderr, list.map(function(p) {
        return {
          filename: p.split('/').pop(),
          path: p
        };
      }));
      done();
    });
  });
}

function convertFileFindTermToRegexString(str) {
  return '.*' + str.split('.').map((s) => escapeRegexChars(s).replace(/\\\//g, '.*\\/.*')).join('[^\\/]*\\.[^\\/]*') + '.*';
}

function escapeRegexChars(s) {
  return s.replace(/(\/|\\|\[|\]|\(|\)|\{|\}|\.|\+|\-|\,|\?|\*|\||\^|\$)/g, '\\$1');
}

function decodeUTF8(s) {
  try {
    return decodeURIComponent(escape(s));
  } catch(err) {
    return s;
  }
}

function findCanonicalPath(conn, path, cb) {
  conn.send('sh.run', JSON.stringify({
    cmd: 'readlink',
    args: ['-f', path],
    run_as: 'root'
  }), function(msgName, data, done) {
    done();
    let resolved = path;
    try {
      const res = JSON.parse(data);
      if (res.exit_code === 0) {
        resolved = decodeUTF8(atob(res.stdout)).trim();
      }
    } catch(err) {}

    cb(resolved);
  });
}
