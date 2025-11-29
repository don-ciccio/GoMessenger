import '../styles/globals.css';
import type { AppProps } from 'next/app';
import Head from 'next/head';

export default function App({ Component, pageProps }: AppProps) {
  return (
    <>
      <Head>
        <title>✨ AI Support Assistant | Next-Gen Customer Support</title>
        <meta name="description" content="AI-powered customer support assistant with advanced RAG technology, Groq LLM, and stunning UI" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <meta name="theme-color" content="#667eea" />
        <meta property="og:title" content="AI Support Assistant" />
        <meta property="og:description" content="Next-generation AI-powered customer support" />
        <meta property="og:type" content="website" />
        
        {/* SVG Favicon - Embedded directly */}
        <link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><defs><linearGradient id='grad' x1='0%' y1='0%' x2='100%' y2='100%'><stop offset='0%' style='stop-color:%23667eea;stop-opacity:1' /><stop offset='100%' style='stop-color:%23f093fb;stop-opacity:1' /></linearGradient></defs><circle cx='50' cy='50' r='45' fill='url(%23grad)'/><path d='M 30 40 Q 35 35 40 40 L 45 45 L 50 40 L 55 45 L 60 40 Q 65 35 70 40' stroke='white' stroke-width='4' fill='none' stroke-linecap='round'/><path d='M 35 60 Q 50 70 65 60' stroke='white' stroke-width='4' fill='none' stroke-linecap='round'/><circle cx='35' cy='38' r='3' fill='white'/><circle cx='65' cy='38' r='3' fill='white'/></svg>" />
        
        {/* Apple Touch Icon */}
        <link rel="apple-touch-icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><defs><linearGradient id='grad' x1='0%' y1='0%' x2='100%' y2='100%'><stop offset='0%' style='stop-color:%23667eea;stop-opacity:1' /><stop offset='100%' style='stop-color:%23f093fb;stop-opacity:1' /></linearGradient></defs><rect width='100' height='100' rx='20' fill='url(%23grad)'/><text x='50' y='70' font-size='60' text-anchor='middle' fill='white'>✨</text></svg>" />
      </Head>
      <Component {...pageProps} />
    </>
  );
}

