import createDoomModule from './engine/chocolate-doom.js';

const canvas = document.querySelector('#doom-canvas');
const boot = document.querySelector('#boot');
const volume = Math.max(0, Math.min(100, Number(new URLSearchParams(location.search).get('volume') ?? 65)));
let engine;

function send(status, message = '') {
  parent.postMessage({ source: 'tanebi95-doom', status, message }, location.origin);
}

function ensureDir(fs, path) {
  try { fs.mkdir(path); } catch (error) {
    if (!String(error?.message ?? '').includes('File exists')) throw error;
  }
}

send('booting');

try {
  const packageParts = await Promise.all(
    [0, 1, 2, 3].map(async (index) => {
      const response = await fetch(`./engine/chocolate-doom.data.part${index}`);
      if (!response.ok) throw new Error(`Game data part ${index} could not be loaded.`);
      return new Uint8Array(await response.arrayBuffer());
    }),
  );
  const packageSize = packageParts.reduce((total, part) => total + part.length, 0);
  const packageBytes = new Uint8Array(packageSize);
  let packageOffset = 0;
  for (const part of packageParts) {
    packageBytes.set(part, packageOffset);
    packageOffset += part.length;
  }

  engine = await createDoomModule({
    canvas,
    keyboardListeningElement: canvas,
    locateFile: (name) => `./engine/${name}`,
    noInitialRun: true,
    getPreloadedPackage: (name, size) => (
      name.endsWith('chocolate-doom.data') && size === packageBytes.length ? packageBytes.buffer : null
    ),
    preRun: [
      (module) => {
        ensureDir(module.FS, '/config');
        ensureDir(module.FS, '/savegames');
        const gameVolume = Math.round((volume / 100) * 15);
        module.FS.writeFile('/config/default.cfg', [
          'fullscreen 0',
          'window_width 320',
          'window_height 200',
          'grabmouse 0',
          'use_mouse 1',
          `sfx_volume ${gameVolume}`,
          `music_volume ${gameVolume}`,
        ].join('\n') + '\n');
        module.FS.writeFile('/config/chocolate-doom.cfg', [
          'smooth_pixel_scaling 0',
          'force_software_renderer 1',
        ].join('\n') + '\n');
      },
    ],
    print: () => {},
    printErr: (message) => console.error(message),
    onAbort: (reason) => send('error', String(reason)),
  });

  try {
    engine.callMain([
      '-window',
      '-iwad', '/iwads/freedoom2.wad',
      '-savedir', '/savegames',
      '-config', '/config/default.cfg',
      '-extraconfig', '/config/chocolate-doom.cfg',
      '-warp', '1',
      '-skill', '3',
    ]);
  } catch (error) {
    const message = String(error ?? '');
    if (!message.includes('unwind') && !message.includes('SimulateInfiniteLoop')) throw error;
  }

  boot.hidden = true;
  canvas.focus();
  send('running');
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  boot.classList.add('error');
  boot.textContent = `起動エラー: ${message}`;
  send('error', message);
}

window.addEventListener('message', (event) => {
  if (event.origin !== location.origin || event.data?.source !== 'tanebi95-shell' || !engine) return;
  if (event.data.action === 'pause') {
    engine.pauseMainLoop?.();
    send('paused');
  }
  if (event.data.action === 'resume') {
    engine.resumeMainLoop?.();
    canvas.focus();
    send('running');
  }
  const key = {
    forward: { code: 'ArrowUp', key: 'ArrowUp', keyCode: 38 },
    fire: { code: 'ControlLeft', key: 'Control', keyCode: 17 },
  }[event.data.action];
  if (key) {
    canvas.focus();
    canvas.dispatchEvent(new KeyboardEvent('keydown', { ...key, bubbles: true }));
    setTimeout(() => canvas.dispatchEvent(new KeyboardEvent('keyup', { ...key, bubbles: true })), 180);
  }
});

canvas.addEventListener('click', () => canvas.focus());
