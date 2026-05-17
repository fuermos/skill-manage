<script>
  import { onMount } from 'svelte';
  import { t as _ } from './lib/i18n.js';

  let server = '';
  let token = '';
  let tools = [];
  let testResult = '';

  async function load() {
    try {
      let cfg = await window.go.main.App.GetConfig();
      server = cfg.server || '';
      token = cfg.token || '';
      tools = await window.go.main.App.GetTools();
    } catch(e) {}
  }

  async function save() {
    try {
      await window.go.main.App.SaveConfig(server, token);
      testResult = _('settings.saved');
    } catch(e) { testResult = _('settings.failed') + ': ' + e; }
  }

  onMount(() => { load(); });
</script>

<div>
  <h2 style="margin-bottom:20px;font-size:20px">{_('settings.title')}</h2>

  <div class="card" style="margin-bottom:16px">
    <div class="card-title">{_('settings.serverConnection')}</div>
    <div class="form-group">
      <label>{_('settings.serverUrl')}</label>
      <input type="text" bind:value={server} placeholder="http://192.168.1.100:8080" />
    </div>
    <div class="form-group">
      <label>{_('settings.authToken')}</label>
      <input type="password" bind:value={token} placeholder="API token" />
    </div>
    <button class="btn-primary" on:click={save}>{_('settings.saveTest')}</button>
    {#if testResult}
      <span style="margin-left:12px;font-size:13px;color:#4ade80">{testResult}</span>
    {/if}
  </div>

  <div class="card">
    <div class="card-title">{_('settings.configuredTools')}</div>
    {#each tools as tool}
      <div class="tool-row">
        <span>{tool.display || tool.name}</span>
        <span class:tag-ok={tool.installed} class:tag-err={!tool.installed}>
          {tool.installed ? _('settings.installed') : _('settings.notInstalled')}
        </span>
        <span style="color:#555;font-size:11px">{tool.enabled ? _('settings.enabled') : _('settings.disabled')}</span>
      </div>
    {/each}
  </div>
</div>

<style>
  .card { background:#1a1a28; border:1px solid #2a2a3a; border-radius:8px; padding:16px; }
  .card-title { font-size:12px; color:#888; text-transform:uppercase; letter-spacing:.5px; margin-bottom:12px; }
  .form-group { margin-bottom:12px; }
  .form-group label { display:block; font-size:12px; color:#888; margin-bottom:4px; }
  .form-group input { width:100%; max-width:400px; background:#0f0f14; border:1px solid #2a2a3a; color:#ccc; padding:8px 12px; border-radius:6px; font-size:13px; }
  .form-group input:focus { border-color:#7c5cfc; outline:none; }
  .btn-primary { background:#7c5cfc; color:#fff; border:none; padding:8px 16px; border-radius:6px; cursor:pointer; font-size:13px; margin-top:8px; }
  .tool-row { display:flex; justify-content:space-between; padding:8px 0; border-bottom:1px solid #1e1e2e; font-size:13px; align-items:center; }
  .tag-ok { color:#4ade80; font-size:11px; }
  .tag-err { color:#f87171; font-size:11px; }
</style>