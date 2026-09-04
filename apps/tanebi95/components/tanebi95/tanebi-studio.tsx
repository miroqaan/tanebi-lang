'use client';

import { useEffect, useState } from 'react';
import { Flame, Play, RotateCcw } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { loadTanebi } from '@/lib/tanebi-runtime';

const starter = `# TANEBI WebAssembly
let 名前 = "アカギ"
let 種火 = 3

repeat 種火 {
    print 名前 + "の世界が目を覚ます"
}

if 種火 >= 3 {
    print "種火は消えない"
}`;

export function TanebiStudio() {
  const [source, setSource] = useState(starter);
  const [output, setOutput] = useState('TANEBI Runtime を読み込んでいます...');
  const [ready, setReady] = useState(false);

  useEffect(() => {
    loadTanebi()
      .then(() => {
        setReady(true);
        setOutput('TANEBI WebAssembly ready.\n実行ボタンを押してください。');
      })
      .catch((error: unknown) => setOutput(`起動エラー: ${error instanceof Error ? error.message : String(error)}`));
  }, []);

  async function run() {
    try {
      const execute = await loadTanebi();
      const result = execute(source);
      setOutput(result.error ? `TANEBI error: ${result.error}` : result.output || '(出力なし)');
    } catch (error) {
      setOutput(`起動エラー: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  return (
    <div className="studio">
      <div className="studio-toolbar">
        <Button className="win95-button primary-action" onClick={run} disabled={!ready}><Play /> 実行</Button>
        <Button className="win95-button" onClick={() => { setSource(starter); setOutput('サンプルを復元しました。'); }}><RotateCcw /> 復元</Button>
        <span className={`runtime-status ${ready ? 'online' : ''}`}><Flame /> {ready ? 'WASM ONLINE' : 'LOADING'}</span>
      </div>
      <div className="studio-panes">
        <div className="studio-pane">
          <label htmlFor="tanebi-source">PROGRAM.TANEBI</label>
          <Textarea id="tanebi-source" className="code-editor" value={source} onChange={(event) => setSource(event.target.value)} spellCheck={false} />
        </div>
        <section className="studio-pane" aria-label="TANEBI output">
          <span>OUTPUT</span>
          <pre className="console-output" aria-live="polite">{output}</pre>
        </section>
      </div>
      <footer className="studio-status">決定論モード · Unicode識別子 · Go/WebAssembly</footer>
    </div>
  );
}
