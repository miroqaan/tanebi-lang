export type TanebiResult = {
  output: string;
  error: string;
};

type GoRuntime = {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
};

declare global {
  interface Window {
    Go?: new () => GoRuntime;
    tanebiRun?: (source: string) => TanebiResult;
  }
}

let runtimePromise: Promise<(source: string) => TanebiResult> | undefined;

function loadScript(source: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(`script[src="${source}"]`);
    if (existing) {
      if (window.Go) resolve();
      else existing.addEventListener('load', () => resolve(), { once: true });
      return;
    }
    const script = document.createElement('script');
    script.src = source;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`failed to load ${source}`));
    document.head.appendChild(script);
  });
}

async function instantiateGoWasm(go: GoRuntime): Promise<WebAssembly.Instance> {
  const response = await fetch('/tanebi.wasm');
  if (!response.ok) throw new Error(`failed to load TANEBI runtime (${response.status})`);

  try {
    const result = await WebAssembly.instantiateStreaming(response.clone(), go.importObject);
    return result.instance;
  } catch {
    const bytes = await response.arrayBuffer();
    const result = await WebAssembly.instantiate(bytes, go.importObject);
    return result;
  }
}

export function loadTanebi(): Promise<(source: string) => TanebiResult> {
  runtimePromise ??= (async () => {
    await loadScript('/wasm_exec.js');
    if (!window.Go) throw new Error('Go WebAssembly bridge is unavailable');
    const go = new window.Go();
    const instance = await instantiateGoWasm(go);
    void go.run(instance);

    for (let attempt = 0; attempt < 100 && !window.tanebiRun; attempt += 1) {
      await new Promise((resolve) => setTimeout(resolve, 10));
    }
    if (!window.tanebiRun) throw new Error('TANEBI runtime did not start');
    return window.tanebiRun;
  })();
  return runtimePromise;
}
