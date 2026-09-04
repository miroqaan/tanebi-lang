import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'TANEBI 95',
  description: 'A Windows 95-inspired desktop shell powered by the TANEBI language.',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}

