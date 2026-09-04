'use client';

import { useEffect, useState } from 'react';
import { Clapperboard, Cpu, Flame, Gamepad2, Gauge, MemoryStick, ShieldCheck, TerminalSquare } from 'lucide-react';

type FeatureTarget = 'studio' | 'doom' | 'media';

export function SystemMonitor({ openWindow }: { openWindow: (id: FeatureTarget) => void }) {
  const [seconds, setSeconds] = useState(0);
  const [memory, setMemory] = useState<number | null>(null);

  useEffect(() => {
    const started = performance.now();
    const update = () => {
      setSeconds(Math.floor((performance.now() - started) / 1000));
      const performanceWithMemory = performance as Performance & { memory?: { usedJSHeapSize: number } };
      setMemory(performanceWithMemory.memory ? Math.round(performanceWithMemory.memory.usedJSHeapSize / 1024 / 1024) : null);
    };
    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, []);

  const minutes = Math.floor(seconds / 60).toString().padStart(2, '0');
  const remainder = (seconds % 60).toString().padStart(2, '0');

  return (
    <div className="monitor-app">
      <section className="monitor-metrics">
        <article><Gauge /><span>稼働時間</span><strong>{minutes}:{remainder}</strong></article>
        <article><MemoryStick /><span>JSヒープ</span><strong>{memory === null ? 'N/A' : `${memory} MB`}</strong></article>
        <article><Cpu /><span>Runtime</span><strong>WASM</strong></article>
        <article><ShieldCheck /><span>Mode</span><strong>SAFE</strong></article>
      </section>

      <section className="architecture-map" aria-label="TANEBI 95 architecture">
        <div><strong>React / vinext</strong><span>Window shell</span></div><b>→</b>
        <div><strong>Go WebAssembly</strong><span>TANEBI host</span></div><b>→</b>
        <div><strong>Lexer · Parser · AST</strong><span>User space</span></div>
      </section>

      <section className="feature-tour">
        <h2><Flame /> FEATURE TOUR</h2>
        <div>
          <button type="button" onClick={() => openWindow('studio')}><TerminalSquare /><strong>TANEBI Studio</strong><span>Unicodeコードを編集し、Go/WASMで実行</span></button>
          <button type="button" onClick={() => openWindow('doom')}><Gamepad2 /><strong>TANEBI 95 DOOM</strong><span>Chocolate Doom + FreedoomをWebAssemblyで実行</span></button>
          <button type="button" onClick={() => openWindow('media')}><Clapperboard /><strong>Media Player</strong><span>アカギユニバース動画を音声付き再生</span></button>
        </div>
      </section>

      <footer className="status-bar">All systems operational · deterministic core online</footer>
    </div>
  );
}
