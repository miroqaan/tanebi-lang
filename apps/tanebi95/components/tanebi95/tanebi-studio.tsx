'use client';

import { Flame } from 'lucide-react';
import { Textarea } from '@/components/ui/textarea';

const starter = `fn main() {
    let 名前:str = "アカギ";
    let 種火:u64 = 0;
    while 種火 < 3 {
        print(名前);
        println("の世界が目を覚ます");
        種火 += 1;
    }
    println("種火は消えない");
}`;

export function TanebiStudio() {
  return (
    <div className="studio">
      <div className="studio-toolbar">
        <span className="runtime-status"><Flame /> ソース閲覧専用</span>
      </div>
      <div className="studio-panes">
        <div className="studio-pane">
          <label htmlFor="tanebi-source">PROGRAM.TANEBI</label>
          <Textarea id="tanebi-source" className="code-editor" value={starter} readOnly spellCheck={false} />
        </div>
        <section className="studio-pane" aria-label="コンパイラーの使い方">
          <span>コンパイルして実行</span>
          <pre className="console-output">{'ブラウザー版インタープリターは廃止されました。\nこの画面ではコードを実行しません。\n\nPC のターミナルで実行してください。\n\ntanebi check program.tanebi\ntanebi build program.tanebi\ntanebi run program.tanebi\n\nrun もネイティブ実行ファイルを生成してから起動します。'}</pre>
        </section>
      </div>
      <footer className="studio-status">Web UI アーカイブ · TANEBI コンパイラーのサンプル</footer>
    </div>
  );
}
