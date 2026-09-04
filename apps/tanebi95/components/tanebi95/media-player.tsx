'use client';

import { useState } from 'react';
import { ExternalLink, Film, Play, Volume2 } from 'lucide-react';

import { Button } from '@/components/ui/button';

const videoUrl = 'https://www.youtube.com/watch?v=low-pfQAI0A&t=974s';
const embedUrl = 'https://www.youtube-nocookie.com/embed/low-pfQAI0A?start=974&autoplay=1&rel=0&playsinline=1';

export function MediaPlayer() {
  const [playing, setPlaying] = useState(false);

  return (
    <div className="media-app">
      <nav className="menu-bar"><span>ファイル(F)</span><span>再生(P)</span><span>表示(V)</span><span>ヘルプ(H)</span></nav>
      <div className="media-toolbar">
        <Button className="win95-button primary-action" type="button" onClick={() => setPlaying(true)}><Play /> 16:14から再生</Button>
        <a className="win95-link-button" href={videoUrl} target="_blank" rel="noreferrer"><ExternalLink /> YouTubeで開く</a>
        <span><Volume2 /> プレイヤー内で音量を調整できます</span>
      </div>
      <div className="media-screen">
        {playing ? (
          <iframe
            src={embedUrl}
            title="【日本語版】4人で潜り、3人で帰った｜100階層ターミナルTRPG自動遠征・30分"
            allow="autoplay; encrypted-media; picture-in-picture"
            allowFullScreen
          />
        ) : (
          <button className="media-cover" type="button" onClick={() => setPlaying(true)}>
            <Film aria-hidden="true" />
            <strong>アカギユニバース</strong>
            <span>【日本語版】4人で潜り、3人で帰った</span>
            <small>100階層ターミナルTRPG自動遠征 · 16:14から音声付き再生</small>
            <em><Play /> PLAY</em>
          </button>
        )}
      </div>
      <footer className="status-bar">YouTube Player · low-pfQAI0A · start=974s · sound enabled</footer>
    </div>
  );
}
