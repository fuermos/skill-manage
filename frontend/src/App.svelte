<script>
  import { onMount, onDestroy } from 'svelte';
  import Dashboard from './Dashboard.svelte';
  import Skills from './Skills.svelte';
  import Sync from './Sync.svelte';
  import Settings from './Settings.svelte';
  import { t, getLocale, setLocale } from './lib/i18n.js';

  let page = 'dashboard';
  let status = { revision: 0, connected: false };
  let lang = getLocale();
  let interval;

  function toggleLang() {
    setLocale(lang === 'zh' ? 'en' : 'zh');
    lang = getLocale();
  }

  async function checkStatus() {
    try {
      let s = await window.go.main.App.GetStatus();
      status = { revision: s.revision, connected: true };
    } catch(e) {
      status = { revision: 0, connected: false };
    }
  }

  onMount(() => {
    checkStatus();
    interval = setInterval(checkStatus, 30000);
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });
</script>

<div class="app">
  <aside class="sidebar">
    <h1 class="logo">{t('app.title')}</h1>
    <p class="subtitle">{t('app.subtitle')}</p>
    <nav>
      <button class:active={page==='dashboard'} on:click={() => page='dashboard'}>{t('nav.dashboard')}</button>
      <button class:active={page==='skills'} on:click={() => page='skills'}>{t('nav.skills')}</button>
      <button class:active={page==='sync'} on:click={() => page='sync'}>{t('nav.sync')}</button>
      <button class:active={page==='settings'} on:click={() => page='settings'}>{t('nav.settings')}</button>
    </nav>
    <div class="bottom">
      <button class="lang-btn" on:click={toggleLang}>{lang === 'zh' ? 'EN' : '中'}</button>
      <div class="status">
        <span class:dot-green={status.connected} class:dot-red={!status.connected}></span>
        {status.connected ? `${t('status.rev')} ${status.revision}` : t('status.offline')}
      </div>
    </div>
  </aside>
  <main>
    {#if page === 'dashboard'}<Dashboard />{/if}
    {#if page === 'skills'}<Skills />{/if}
    {#if page === 'sync'}<Sync />{/if}
    {#if page === 'settings'}<Settings />{/if}
  </main>
</div>

<style>
  :global(*) { margin:0; padding:0; box-sizing:border-box; }
  :global(body) { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', sans-serif; background:#0f0f14; color:#e0e0e0; }
  .app { display:flex; height:100vh; }
  .sidebar { width:200px; background:#16161f; border-right:1px solid #2a2a3a; display:flex; flex-direction:column; padding:16px 0; flex-shrink:0; }
  .logo { font-size:18px; color:#7c5cfc; padding:0 16px; }
  .subtitle { font-size:11px; color:#555; padding:2px 16px 20px; }
  nav { flex:1; display:flex; flex-direction:column; gap:2px; }
  nav button { background:none; border:none; color:#888; padding:10px 16px; text-align:left; cursor:pointer; font-size:13px; }
  nav button:hover { background:#1e1e2e; color:#ccc; }
  nav button.active { background:#7c5cfc20; color:#7c5cfc; border-left:3px solid #7c5cfc; padding-left:13px; }
  .bottom { padding:12px 16px; border-top:1px solid #2a2a3a; margin-top:8px; }
  .lang-btn { background:#2a2a3a; color:#888; border:none; padding:4px 10px; border-radius:4px; cursor:pointer; font-size:11px; margin-bottom:8px; }
  .lang-btn:hover { background:#3a3a4a; color:#ccc; }
  .status { font-size:12px; color:#555; display:flex; align-items:center; gap:8px; }
  main { flex:1; padding:24px; overflow-y:auto; }
  .dot-green { width:8px; height:8px; border-radius:50%; background:#4ade80; display:inline-block; }
  .dot-red { width:8px; height:8px; border-radius:50%; background:#f87171; display:inline-block; }
</style>