const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
    isElectron: true,
    platform: process.platform,
    versions: {
        electron: process.versions.electron,
        chrome: process.versions.chrome,
        node: process.versions.node
    },
    openExternal: (url) => {
        if (typeof url === 'string' && (url.startsWith('https://') || url.startsWith('http://') || url.startsWith('mailto:'))) {
            ipcRenderer.send('open-external', url);
        }
    },
    onMenuAction: (callback) => {
        if (typeof callback === 'function') {
            ipcRenderer.on('menu-action', (_event, action) => callback(action));
        }
    }
});
