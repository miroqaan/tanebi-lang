'use client';

import { type PointerEvent, type ReactNode, useRef } from 'react';
import type { LucideIcon } from 'lucide-react';

export type WindowModel = {
  id: string;
  title: string;
  icon: LucideIcon;
  x: number;
  y: number;
  width: number;
  height: number;
  z: number;
  open: boolean;
  minimized: boolean;
  maximized: boolean;
};

type Props = {
  window: WindowModel;
  children: ReactNode;
  onFocus: () => void;
  onMove: (x: number, y: number) => void;
  onMinimize: () => void;
  onMaximize: () => void;
  onClose: () => void;
};

export function WindowFrame({ window, children, onFocus, onMove, onMinimize, onMaximize, onClose }: Props) {
  const dragStart = useRef<{ pointerX: number; pointerY: number; x: number; y: number } | null>(null);
  const Icon = window.icon;

  if (!window.open || window.minimized) return null;

  function beginDrag(event: PointerEvent<HTMLElement>) {
    if (window.maximized || (event.target as HTMLElement).closest('button')) return;
    onFocus();
    dragStart.current = { pointerX: event.clientX, pointerY: event.clientY, x: window.x, y: window.y };
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function drag(event: PointerEvent<HTMLElement>) {
    if (!dragStart.current) return;
    const nextX = Math.max(0, Math.min(globalThis.innerWidth - 180, dragStart.current.x + event.clientX - dragStart.current.pointerX));
    const nextY = Math.max(0, Math.min(globalThis.innerHeight - 90, dragStart.current.y + event.clientY - dragStart.current.pointerY));
    onMove(nextX, nextY);
  }

  function endDrag(event: PointerEvent<HTMLElement>) {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) event.currentTarget.releasePointerCapture(event.pointerId);
    dragStart.current = null;
  }

  const style = window.maximized
    ? { inset: '4px 4px 48px', zIndex: window.z }
    : { left: window.x, top: window.y, width: window.width, height: window.height, zIndex: window.z };

  return (
    <section className={`win95-window app-window${window.maximized ? ' maximized' : ''}`} style={style} onPointerDown={onFocus}>
      <header className="title-bar draggable" onPointerDown={beginDrag} onPointerMove={drag} onPointerUp={endDrag} onPointerCancel={endDrag} onDoubleClick={onMaximize}>
        <span><Icon aria-hidden="true" /> {window.title}</span>
        <span className="window-controls">
          <button type="button" onClick={onMinimize} aria-label={`${window.title}を最小化`}>_</button>
          <button type="button" onClick={onMaximize} aria-label={`${window.title}を最大化`}>{window.maximized ? '❐' : '□'}</button>
          <button type="button" onClick={onClose} aria-label={`${window.title}を閉じる`}>×</button>
        </span>
      </header>
      <div className="window-body">{children}</div>
    </section>
  );
}
