//
// Electron
//
if (window.process && window.process.type === 'renderer') {
  window.electron = require('electron');
} else {
  window.electron = null;
}

//
// WASM
//
var Module = {};
window.initElectronFS = () => {
  if (electron) {
    // Mount assets directory for direct read / write access
    const repoPath = electron.remote.app
      .getAppPath()
      .replace(/\\/g, '/')
      .replace(/\/raylib-5k\/.*?$/, '/raylib-5k');
    window.assetsPath = repoPath + '/assets';
    try {
      FS.rename('assets', 'assets-orig');
    } catch (e) {}
    FS.mkdir('assets');
    FS.mount(NODEFS, { root: assetsPath }, 'assets');
  }
};
(() => {
  Module.canvas = document.getElementById('canvas');

  // Relative URL to '.wasm' file fails without this
  Module.locateFile = (filename) => window.location.href.replace(/[^/]*$/, '') + filename;

  // Fetch and launch core JS, trying to skip cache
  const s = document.createElement('script');
  s.async = true;
  s.type = 'text/javascript';
  s.src = '@PROJECT_NAME@.js?ts=' + new Date().getTime(); // CMake replaces filename
  document.getElementsByTagName('head')[0].appendChild(s);
})();

//
// Auto-reload
//
(() => {
  if (window.location.href.startsWith('http://localhost')) {
    const filenames = ['reload-trigger'];
    let lastUnfocusedTime = Date.now();
    let reloading = false;
    let interval;
    try {
      const getTimestampForFilename = async (filename) =>
        (await fetch(filename, { method: 'HEAD' })).headers.get('last-modified');
      const initialTimestamps = {};
      filenames.forEach(async (filename) => {
        initialTimestamps[filename] = await getTimestampForFilename(filename);
      });
      interval = setInterval(() => {
        const now = Date.now();
        if (!document.hasFocus() || now - lastUnfocusedTime < 3000) {
          filenames.forEach(async (filename) => {
            try {
              if (initialTimestamps[filename]) {
                const currentTimestamp = await getTimestampForFilename(filename);
                if (initialTimestamps[filename] !== currentTimestamp) {
                  if (!reloading) {
                    reloading = true;
                    Module._JS_saveEditSession();
                    console.log('reloading...');
                    setTimeout(() => window.location.reload(), 60);
                  }
                }
              }
            } catch (e) {
              clearInterval(interval);
            }
          });
        }
        if (!document.hasFocus()) {
          lastUnfocusedTime = now;
        }
      }, 280);
    } catch (e) {
      clearInterval(interval);
    }
  }
})();

//
// UI
//
(() => {
  IncrementalDOM.attributes.value = IncrementalDOM.applyProp;

  window.UI = {};
  const UI = window.UI;

  UI.nextToken = 0;

  const isInput = (elem) => elem.tagName === 'INPUT' || elem.tagName === 'TEXTAREA';

  UI.keyboardCaptureCount = 0;
  UI.eventCounts = new WeakMap();
  UI.setupEventListener = (target, type) => {
    target.addEventListener(type, (e) => {
      if (e.type === 'click' && e.detail === 0) {
        return;
      }
      const target = e.target;
      if (target.tagName === 'SUMMARY' && e.type === 'click') {
        e.preventDefault();
      }
      let counts = UI.eventCounts.get(target);
      if (counts === undefined) {
        counts = {};
        UI.eventCounts.set(target, counts);
        UI.noEvents = false;
      }
      const count = counts[e.type];
      if (count === undefined) {
        counts[e.type] = 1;
      } else {
        counts[e.type] = count + 1;
      }
    });
    if (isInput(target)) {
      target.addEventListener('focus', () => ++UI.keyboardCaptureCount);
      target.addEventListener('blur', () => setTimeout(() => --UI.keyboardCaptureCount, 100));
      const blurOnEnter = (e) => {
        const target = e.target;
        if (!e.shiftKey && e.key === 'Enter') {
          target.blur();
        }
      };
      target.addEventListener('keydown', blurOnEnter);
      target.addEventListener('keyup', blurOnEnter);
      if (target.type === 'number') {
        target.addEventListener('click', (e) => e.target.select());
      }
    }
  };

  const oldPreventDefault = Event.prototype.preventDefault;
  Event.prototype.preventDefault = function() {
    if (!(isInput(this.target) && (this.key === 'Backspace' || this.key === 'Tab'))) {
      oldPreventDefault.bind(this)();
    }
  };

  const preventDefaultKeys = (e) => {
    const target = e.target;
    if (!(target && isInput(target)) && e.code === 'Space') {
      e.preventDefault();
    }
    if (e.ctrlKey && (e.key === 'z' || e.key === 'y' || e.key === 's')) {
      e.preventDefault();
    }
  };
  window.addEventListener('keydown', preventDefaultKeys);
  window.addEventListener('keyup', preventDefaultKeys);
})();

//
// Play mode
//
(() => {
  if (new URLSearchParams(window.location.search).has('play')) {
    window.isPlayMode = true;
    ['top-container', 'bottom-container', 'side-container'].forEach((id) => {
      const el = document.getElementsByClassName(id)[0];
      if (el) {
        el.remove();
      }
    });
  }
})();
