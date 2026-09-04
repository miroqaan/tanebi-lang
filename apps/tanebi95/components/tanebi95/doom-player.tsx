'use client';

import { useEffect, useState } from 'react';
import { ChevronsUp, Crosshair, Gamepad2, Pause, Play, RotateCcw, Square, Volume2 } from 'lucide-react';

import { Button } from '@/components/ui/button';

type DoomMessage = {
  source?: string;
  status?: 'booting' | 'running' | 'paused' | 'error';
  message?: string;
};

export function DoomPlayer() {
  const [status, setStatus] = useState('READY');
  const [started, setStarted] = useState(false);
  const [paused, setPaused] = useState(false);
  const [volume, setVolume] = useState(() => (
    typeof window !== 'undefined' && new URLSearchParams(window.location.search).has('muteGame') ? 0 : 65
  ));
  const [frameSrc, setFrameSrc] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    function receive(event: MessageEvent<DoomMessage>) {
      if (event.origin !== window.location.origin || event.data?.source !== 'tanebi95-doom') return;
      if (event.data.status === 'booting') setStatus('CHOCOLATE DOOM STARTING...');
      if (event.data.status === 'running') setStatus('FREEDOOM PHASE 2');
      if (event.data.status === 'paused') setStatus('PAUSED');
      if (event.data.status === 'error') {
        setStatus('ERROR');
        setError(event.data.message ?? 'DOOM engine failed to start.');
      }
    }

    window.addEventListener('message', receive);
    return () => window.removeEventListener('message', receive);
  }, []);

  function start() {
    setError('');
    setStatus('LOADING 30.7 MB WASM ENGINE...');
    setStarted(true);
    setPaused(false);
    setFrameSrc(`/doom/player.html?volume=${volume}&session=${Date.now()}`);
  }

  function stop() {
    setFrameSrc('');
    setStarted(false);
    setPaused(false);
    setStatus('READY');
  }

  function togglePause() {
    const next = !paused;
    const frame = document.querySelector<HTMLIFrameElement>('iframe[data-doom-frame="true"]');
    frame?.contentWindow?.postMessage({ source: 'tanebi95-shell', action: next ? 'pause' : 'resume' }, window.location.origin);
    setPaused(next);
    setStatus(next ? 'PAUSED' : 'FREEDOOM PHASE 2');
  }

  function sendGameAction(action: 'forward' | 'fire') {
    const frame = document.querySelector<HTMLIFrameElement>('iframe[data-doom-frame="true"]');
    frame?.contentWindow?.postMessage({ source: 'tanebi95-shell', action }, window.location.origin);
  }

  return (
    <div className="doom-app">
      <div className="doom-toolbar">
        <Button className="win95-button primary-action" type="button" onClick={start}>
          {started ? <RotateCcw /> : <Play />} {started ? '再起動' : 'DOOMを起動'}
        </Button>
        <Button className="win95-button" type="button" onClick={togglePause} disabled={!started || status === 'ERROR'}>
          {paused ? <Play /> : <Pause />} {paused ? '再開' : '一時停止'}
        </Button>
        <Button className="win95-button" type="button" onClick={stop} disabled={!started}>
          <Square /> 停止
        </Button>
        <Button className="win95-button" type="button" onClick={() => sendGameAction('forward')} disabled={!started || paused}>
          <ChevronsUp /> 前進
        </Button>
        <Button className="win95-button" type="button" onClick={() => sendGameAction('fire')} disabled={!started || paused}>
          <Crosshair /> 攻撃
        </Button>
        <label className="doom-volume">
          <Volume2 aria-hidden="true" />
          <input aria-label="DOOM 音量" type="range" min="0" max="100" value={volume} disabled={started} onChange={(event) => setVolume(Number(event.target.value))} />
          <span>{volume}%</span>
        </label>
      </div>

      <div className="doom-screen-wrap">
        {frameSrc && (
          <iframe
            key={frameSrc}
            allow="autoplay; fullscreen; gamepad"
            aria-label="Freedoom game"
            className="doom-screen doom-frame"
            data-doom-frame="true"
            src={frameSrc}
            title="TANEBI 95 DOOM — Freedoom"
          />
        )}
        {!started && (
          <div className="doom-splash">
            <Gamepad2 aria-hidden="true" />
            <h2>TANEBI 95 DOOM</h2>
            <p>Chocolate DoomのWebAssemblyエンジンでFreedoom Phase 2を起動します。</p>
            <p className="doom-keys">矢印: 移動 · Ctrl: 攻撃 · Space: 使用 · Shift: 走る</p>
          </div>
        )}
        {error && <p className="doom-error">{error}</p>}
      </div>

      <footer className="doom-status">
        <span>{status}</span>
        <span>Chocolate Doom · Freedoom 0.13.0 · WebAssembly · 音声対応</span>
      </footer>
    </div>
  );
}
