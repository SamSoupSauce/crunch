const { app, BrowserWindow, Menu, shell, ipcMain } = require('electron');
const path = require('path');

let mainWindow = null;

function createWindow() {
    const isMac = process.platform === 'darwin';

    mainWindow = new BrowserWindow({
        width: 1280,
        height: 840,
        minWidth: 960,
        minHeight: 620,
        backgroundColor: '#0d1117',
        show: false,
        title: 'Crunch — Rock Chess',
        icon: isMac
            ? path.join(__dirname, 'build', 'icon.icns')
            : path.join(__dirname, 'build', 'icon.ico'),
        titleBarStyle: isMac ? 'hiddenInset' : 'default',
        trafficLightPosition: isMac ? { x: 16, y: 16 } : undefined,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),
            contextIsolation: true,
            nodeIntegration: false,
            sandbox: true,
            spellcheck: false
        }
    });

    mainWindow.loadFile('index.html');

    mainWindow.once('ready-to-show', () => {
        mainWindow.show();
    });

    // Intercept target="_blank" and window.open calls to open in the system's default browser
    mainWindow.webContents.setWindowOpenHandler(({ url }) => {
        if (url.startsWith('http://') || url.startsWith('https://')) {
            shell.openExternal(url);
            return { action: 'deny' };
        }
        return { action: 'allow' };
    });

    // Prevent in-window navigation to external links
    mainWindow.webContents.on('will-navigate', (event, url) => {
        const fileUrlPrefix = 'file://';
        if (!url.startsWith(fileUrlPrefix)) {
            event.preventDefault();
            shell.openExternal(url);
        }
    });

    mainWindow.on('closed', () => {
        mainWindow = null;
    });

    buildAppMenu();
}

function sendMenuAction(action) {
    if (mainWindow && !mainWindow.isDestroyed()) {
        mainWindow.webContents.send('menu-action', action);
    }
}

function buildAppMenu() {
    const isMac = process.platform === 'darwin';

    const template = [
        ...(isMac ? [{
            label: app.name,
            submenu: [
                { role: 'about' },
                { type: 'separator' },
                { role: 'services' },
                { type: 'separator' },
                { role: 'hide' },
                { role: 'hideOthers' },
                { role: 'unhide' },
                { type: 'separator' },
                { role: 'quit' }
            ]
        }] : []),
        {
            label: 'Game',
            submenu: [
                {
                    label: 'Main Menu',
                    accelerator: 'CmdOrCtrl+M',
                    click: () => sendMenuAction('main-menu')
                },
                {
                    label: 'New Singleplayer Match',
                    accelerator: 'CmdOrCtrl+N',
                    click: () => sendMenuAction('new-singleplayer')
                },
                { type: 'separator' },
                {
                    label: 'Create Online Game',
                    click: () => sendMenuAction('create-room')
                },
                {
                    label: 'Join Online Game',
                    accelerator: 'CmdOrCtrl+J',
                    click: () => sendMenuAction('join-room')
                },
                { type: 'separator' },
                isMac ? { role: 'close' } : { role: 'quit' }
            ]
        },
        {
            label: 'Edit',
            submenu: [
                { role: 'undo' },
                { role: 'redo' },
                { type: 'separator' },
                { role: 'cut' },
                { role: 'copy' },
                { role: 'paste' },
                { role: 'selectAll' }
            ]
        },
        {
            label: 'View',
            submenu: [
                { role: 'reload' },
                { role: 'forceReload' },
                { role: 'toggleDevTools' },
                { type: 'separator' },
                { role: 'resetZoom' },
                { role: 'zoomIn' },
                { role: 'zoomOut' },
                { type: 'separator' },
                { role: 'togglefullscreen' }
            ]
        },
        {
            label: 'Window',
            submenu: [
                { role: 'minimize' },
                { role: 'zoom' },
                ...(isMac ? [
                    { type: 'separator' },
                    { role: 'front' },
                    { type: 'separator' },
                    { role: 'window' }
                ] : [
                    { role: 'close' }
                ])
            ]
        },
        {
            role: 'help',
            submenu: [
                {
                    label: 'Crunch Official Rules',
                    accelerator: 'CmdOrCtrl+H',
                    click: () => sendMenuAction('open-rules')
                },
                { type: 'separator' },
                {
                    label: 'GitHub Repository',
                    click: () => shell.openExternal('https://github.com/SamSoupSauce/crunch')
                },
                {
                    label: 'Report an Issue',
                    click: () => shell.openExternal('https://github.com/SamSoupSauce/crunch/issues')
                },
                {
                    label: 'Developer Website',
                    click: () => shell.openExternal('https://samuel-meyers.com')
                }
            ]
        }
    ];

    const menu = Menu.buildFromTemplate(template);
    Menu.setApplicationMenu(menu);
}

// IPC handlers
ipcMain.on('open-external', (_event, url) => {
    if (typeof url === 'string' && (url.startsWith('https://') || url.startsWith('http://'))) {
        shell.openExternal(url);
    }
});

// App lifecycle
app.whenReady().then(() => {
    createWindow();

    app.on('activate', () => {
        if (BrowserWindow.getAllWindows().length === 0) {
            createWindow();
        }
    });
});

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});
