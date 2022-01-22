const { app, BrowserWindow, Menu } = require('electron');
const http = require('http');
const { default: installExtension } = require('electron-devtools-installer');

Menu.setApplicationMenu(
  Menu.buildFromTemplate([
    {
      label: 'raylib 5k editor',
      submenu: [{ role: 'quit' }],
    },
    {
      label: 'Edit',
      submenu: [
        { role: 'cut', label: 'Cut text' },
        { role: 'copy', label: 'Copy text' },
        { role: 'paste', label: 'Paste text' },
        { role: 'selectAll', label: 'Select all text' },
      ],
    },
    {
      label: 'View',
      submenu: [
        { role: 'reload' },
        { role: 'forceReload' },
        { role: 'toggleDevTools' },
        { type: 'separator' },
        { role: 'togglefullscreen' },
        { type: 'separator' },
        { role: 'zoomIn', label: 'Larger UI elements' },
        { role: 'zoomOut', label: 'Smaller UI elements' },
        { role: 'resetZoom', label: 'Reset UI elements size' },
      ],
    },
  ])
);

app.whenReady().then(() => {
  const win = new BrowserWindow({
    webPreferences: {
      webviewTag: true,
      nodeIntegration: true,
      nodeIntegrationInSubFrames: false,
      nativeWindowOpen: true,
      contextIsolation: false,
      enableRemoteModule: true,
    },
    backgroundColor: '#000000',
  });
  win.setMenuBarVisibility(false);
  win.maximize();

  const loadFile = () => {
    const repoPath = app
      .getAppPath()
      .replace(/\\/g, '/')
      .replace(/\/raylib-5k\/.*?$/, '/raylib-5k');
    win.loadFile(repoPath + '/app/web-release-fast/index.html');
  };
  const loadLocalhost = () => {
    win.webContents.openDevTools();
    win.loadURL('http://localhost:9002/index.html');
  };
  const req = http.request(
    { method: 'HEAD', host: 'localhost', port: 9002, path: '/reload-trigger' },
    (res) => {
      if (res.statusCode === 200) {
        loadLocalhost();
      } else {
        loadFile();
      }
    }
  );
  req.on('error', () => {
    loadFile();
  });
  req.end();

  installExtension('pdcpmagijalfljmkmjngeonclgbbannb');
});
