class WatchManager {
  constructor() {
    this.reset();
  }

  reset() {
    this.watches = {}; // <ID, Path>
    this.debounceFuncs = {}; // <ID, functions>
  }

  add(id, path) {
    this.watches[id] = path;
  }

  remove(id) {
    delete this.watches[id];
  }

  debounce(id, wait, func) {
    if (!func) {
      return;
    } else if (wait <= 0) {
      func();
      return;
    }
    // ensures func is run only once within "wait" milliseconds
    const d = this.debounceFuncs[id];
    this.debounceFuncs[id] = func;
    if (!d) {
      window.setTimeout(() => {
        this.debounceFuncs[id]();
        delete this.debounceFuncs[id];
      }, wait);
    }
  }

  pathForID(id) {
    return this.watches[id];
  }

  idForPath(path) {
    for (let id in this.watches) {
      if (this.watches.hasOwnProperty(id)) {
        if (this.watches[id] === path) {
          return id;
        }
      }
    }
  }

  ifWatching(path, func) {
    if (this.idForPath(path)) {
      func();
    }
  }

  unlessWatching(path, func) {
    if (!this.idForPath(path)) {
      func();
    }
  }

  // returns ids of watched subpaths of the given path, as well as
  // the id of the given path itself
  idsForSubPaths(path) {
    const prefix = path === '/' ? '/' : path + '/',
          ids = [];

    for (let id in this.watches) {
      if (this.watches.hasOwnProperty(id)) {
        const p = this.watches[id];
        if (p.indexOf(prefix) === 0 || p === path) {
          ids.push(id);
        }
      }
    }
    return ids;
  }
}

export default new WatchManager();
