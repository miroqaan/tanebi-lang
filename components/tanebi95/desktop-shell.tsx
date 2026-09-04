'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  BookOpenText,
  ChevronRight,
  FileCode2,
  FileText,
  Flame,
  FolderClosed,
  HardDrive,
  Info,
  MonitorCog,
  Power,
  Recycle,
  RotateCcw,
  TerminalSquare,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { loadTanebi } from '@/lib/tanebi-runtime';
import { TanebiStudio } from './tanebi-studio';
import { WindowFrame, type WindowModel } from './window-frame';

type WindowId = 'computer' | 'studio' | 'documents' | 'recycle' | 'about';

type SystemProfile = {
  name: string;
  motto: string;
  labels: Partial<Record<WindowId, string>>;
  boot: string[];
};

const defaultProfile: SystemProfile = {
  name: 'TANEBI 95',
  motto: '一行のコードが、世界に火を灯す。',
  labels: {
    computer: 'TANEBI 95',
    studio: 'TANEBI Studio',
    documents: 'ドキュメント',
    recycle: 'ごみ箱',
  },
  boot: [],
};

const icons = {
  computer: Flame,
  studio: Flame,
  documents: FolderClosed,
  recycle: Recycle,
  about: Info,
};

const initialWindows: Record<WindowId, WindowModel> = {
  computer: { id: 'computer', title: 'TANEBI 95', icon: Flame, x: 148, y: 72, width: 570, height: 370, z: 2, open: true, minimized: false, maximized: false },
  studio: { id: 'studio', title: 'TANEBI Studio', icon: Flame, x: 255, y: 105, width: 760, height: 500, z: 1, open: false, minimized: false, maximized: false },
  documents: { id: 'documents', title: 'ドキュメント', icon: FolderClosed, x: 210, y: 118, width: 560, height: 350, z: 1, open: false, minimized: false, maximized: false },
  recycle: { id: 'recycle', title: 'ごみ箱', icon: Recycle, x: 310, y: 155, width: 420, height: 260, z: 1, open: false, minimized: false, maximized: false },
  about: { id: 'about', title: 'TANEBI 95 について', icon: Info, x: 380, y: 125, width: 480, height: 330, z: 1, open: false, minimized: false, maximized: false },
};

function parseProfile(output: string): SystemProfile {
  const profile: SystemProfile = { ...defaultProfile, labels: { ...defaultProfile.labels }, boot: [] };
  for (const line of output.trim().split('\n')) {
    const [command, key, ...rest] = line.split('|');
    const value = rest.length ? rest.join('|') : key;
    if (command === 'SYSTEM' && key) profile.name = key;
    if (command === 'MOTTO' && key) profile.motto = key;
    if (command === 'ICON' && key && value && key in icons) profile.labels[key as WindowId] = value;
    if (command === 'BOOT' && key) profile.boot.push(key);
  }
  return profile;
}

export function DesktopShell() {
  const [profile, setProfile] = useState(defaultProfile);
  const [windows, setWindows] = useState(initialWindows);
  const [nextZ, setNextZ] = useState(3);
  const [startOpen, setStartOpen] = useState(false);
  const [shutdown, setShutdown] = useState(false);
  const [booting, setBooting] = useState(true);
  const [bootMessage, setBootMessage] = useState('Loading Go host services...');
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 30_000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    let cancelled = false;
    async function boot() {
      try {
        setBootMessage('Igniting TANEBI WebAssembly...');
        const [execute, response] = await Promise.all([loadTanebi(), fetch('/system.tanebi')]);
        if (!response.ok) throw new Error(`system profile ${response.status}`);
        const result = execute(await response.text());
        if (result.error) throw new Error(result.error);
        const loaded = parseProfile(result.output);
        if (!cancelled) {
          setProfile(loaded);
          setWindows((current) => {
            const next = { ...current };
            for (const id of Object.keys(loaded.labels) as WindowId[]) {
              if (loaded.labels[id] && next[id]) next[id] = { ...next[id], title: loaded.labels[id] as string };
            }
            return next;
          });
          setBootMessage(loaded.boot.at(-1) ?? 'TANEBI user space: ignited');
          setTimeout(() => !cancelled && setBooting(false), 650);
        }
      } catch (error) {
        if (!cancelled) {
          setBootMessage(`Safe mode: ${error instanceof Error ? error.message : String(error)}`);
          setTimeout(() => !cancelled && setBooting(false), 900);
        }
      }
    }
    void boot();
    return () => { cancelled = true; };
  }, []);

  const openWindows = useMemo(() => Object.values(windows).filter((window) => window.open), [windows]);

  function updateWindow(id: WindowId, patch: Partial<WindowModel>) {
    setWindows((current) => ({ ...current, [id]: { ...current[id], ...patch } }));
  }

  function focusWindow(id: WindowId) {
    setWindows((current) => ({ ...current, [id]: { ...current[id], z: nextZ } }));
    setNextZ((value) => value + 1);
  }

  function openWindow(id: WindowId) {
    updateWindow(id, { open: true, minimized: false, z: nextZ });
    setNextZ((value) => value + 1);
    setStartOpen(false);
  }

  function taskClick(id: WindowId) {
    const item = windows[id];
    if (item.minimized) openWindow(id);
    else if (item.z === Math.max(...openWindows.filter((entry) => !entry.minimized).map((entry) => entry.z), 0)) updateWindow(id, { minimized: true });
    else focusWindow(id);
  }

  const desktopItems: Array<{ id: WindowId; label: string }> = [
    { id: 'computer', label: profile.labels.computer ?? 'TANEBI 95' },
    { id: 'studio', label: profile.labels.studio ?? 'TANEBI Studio' },
    { id: 'documents', label: profile.labels.documents ?? 'ドキュメント' },
    { id: 'recycle', label: profile.labels.recycle ?? 'ごみ箱' },
  ];

  if (shutdown) {
    return (
      <main className="shutdown-screen">
        <Flame aria-hidden="true" />
        <h1>TANEBI 95を終了しました。</h1>
        <p>種火は保存されています。</p>
        <Button className="win95-button" onClick={() => { setShutdown(false); setBooting(true); setTimeout(() => setBooting(false), 800); }}><RotateCcw /> 再起動</Button>
      </main>
    );
  }

  return (
    <main className="desktop" onPointerDown={() => setStartOpen(false)}>
      <div className="desktop-brand" aria-label={profile.name}>
        <Flame aria-hidden="true" /><span>TANEBI</span><strong>95</strong>
      </div>
      <p className="desktop-motto">{profile.motto}</p>

      <section className="desktop-icons" aria-label="デスクトップ アイコン">
        {desktopItems.map(({ id, label }) => {
          const Icon = icons[id];
          return (
            <button className="desktop-icon" key={id} type="button" onClick={(event) => { event.stopPropagation(); openWindow(id); }}>
              <span className="desktop-icon-image"><Icon aria-hidden="true" /></span>
              <span>{label}</span>
            </button>
          );
        })}
      </section>

      {(Object.keys(windows) as WindowId[]).map((id) => (
        <WindowFrame
          key={id}
          window={windows[id]}
          onFocus={() => focusWindow(id)}
          onMove={(x, y) => updateWindow(id, { x, y })}
          onMinimize={() => updateWindow(id, { minimized: true })}
          onMaximize={() => updateWindow(id, { maximized: !windows[id].maximized })}
          onClose={() => updateWindow(id, { open: false })}
        >
          <WindowContent id={id} profile={profile} openWindow={openWindow} />
        </WindowFrame>
      ))}

      {startOpen && (
        <section className="start-menu" onPointerDown={(event) => event.stopPropagation()}>
          <div className="start-rail"><span>TANEBI</span><strong>95</strong></div>
          <div className="start-items">
            <button type="button" onClick={() => openWindow('studio')}><Flame /><span>プログラム</span><ChevronRight /></button>
            <button type="button" onClick={() => openWindow('documents')}><FileText /><span>ドキュメント</span><ChevronRight /></button>
            <button type="button" onClick={() => openWindow('computer')}><MonitorCog /><span>システム</span><ChevronRight /></button>
            <button type="button" onClick={() => openWindow('about')}><Info /><span>TANEBI 95について</span></button>
            <hr />
            <button type="button" onClick={() => setShutdown(true)}><Power /><span>シャットダウン...</span></button>
          </div>
        </section>
      )}

      <footer className="taskbar" onPointerDown={(event) => event.stopPropagation()}>
        <Button className="start-button" onClick={() => setStartOpen((open) => !open)} aria-expanded={startOpen}><Flame /> スタート</Button>
        {openWindows.map((window) => {
          const Icon = window.icon;
          return <button className={`task-button${!window.minimized && window.z === Math.max(...openWindows.map((entry) => entry.z)) ? ' active' : ''}`} key={window.id} type="button" onClick={() => taskClick(window.id as WindowId)}><Icon /> {window.title}</button>;
        })}
        <div className="taskbar-spacer" />
        <div className="tray"><span className="tray-light" />種火 ON <time>{now.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })}</time></div>
      </footer>

      {booting && (
        <section className="boot-screen" aria-live="polite">
          <div className="boot-logo"><Flame /><span>TANEBI</span><strong>95</strong></div>
          <p>{bootMessage}</p>
          <div className="boot-progress"><span /></div>
        </section>
      )}
    </main>
  );
}

function WindowContent({ id, profile, openWindow }: { id: WindowId; profile: SystemProfile; openWindow: (id: WindowId) => void }) {
  if (id === 'studio') return <TanebiStudio />;
  if (id === 'computer') {
    return (
      <div className="window-layout">
        <nav className="menu-bar"><span>ファイル(F)</span><span>編集(E)</span><span>表示(V)</span><span>ヘルプ(H)</span></nav>
        <div className="address-bar"><span>アドレス</span><strong>マイ コンピュータ</strong></div>
        <div className="computer-grid">
          <button type="button"><HardDrive /><span><strong>System Host (C:)</strong><small>Go services · 95 MB</small></span></button>
          <button type="button" onDoubleClick={() => openWindow('studio')}><TerminalSquare /><span><strong>TANEBI Space (T:)</strong><small>User scripts · WASM</small></span></button>
          <button type="button" onDoubleClick={() => openWindow('documents')}><FileText /><span><strong>README.TXT</strong><small>Architecture notes</small></span></button>
        </div>
        <footer className="status-bar">3 個のオブジェクト · TANEBI user space online</footer>
      </div>
    );
  }
  if (id === 'documents') {
    return (
      <div className="window-layout">
        <nav className="menu-bar"><span>ファイル(F)</span><span>編集(E)</span><span>表示(V)</span></nav>
        <div className="file-list">
          <button type="button"><FileCode2 /><span>system.tanebi</span><small>システム構成</small></button>
          <button type="button"><BookOpenText /><span>ARCHITECTURE.TXT</span><small>Kernel / Host / User Space</small></button>
          <button type="button"><FileText /><span>AKAGI.TXT</span><small>一行のコードが、世界に火を灯す。</small></button>
        </div>
        <footer className="status-bar">TANEBIスクリプトがデスクトップ名とアイコンを定義しています。</footer>
      </div>
    );
  }
  if (id === 'recycle') {
    return <div className="empty-window"><Recycle /><p>ごみ箱は空です。</p><small>失敗した未来はNightglassへ移動しました。</small></div>;
  }
  return (
    <div className="about-window">
      <div className="about-logo"><Flame /><span>TANEBI</span><strong>95</strong></div>
      <h2>{profile.name}</h2>
      <p>{profile.motto}</p>
      <dl>
        <div><dt>System</dt><dd>Go host services</dd></div>
        <div><dt>User space</dt><dd>TANEBI WebAssembly</dd></div>
        <div><dt>Reference</dt><dd>GopherOS boundary architecture</dd></div>
      </dl>
      <p className="copyright">Akagi Universe · deterministic creative system</p>
    </div>
  );
}
