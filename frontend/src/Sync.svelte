<script>
  import { onMount } from 'svelte';
  import { t as _ } from './lib/i18n.js';

  let log = [];
  let pending = [];

  function addLog(msg, cls) {
    log = [...log, { msg, cls, time: new Date().toLocaleTimeString() }];
    log = log;
  }

  async function push() {
    addLog(_('sync.push') + '...', 'info');
    try {
      let r = await window.go.main.App.Push();
      addLog(`${_('sync.pushed')}: ${r.applied} ${_('sync.changes')} (rev ${r.new_revision})`, r.applied > 0 ? 'success' : 'info');
      loadDiff();
    } catch(e) { addLog(_('sync.pushFailed') + ': ' + e, 'error'); }
  }

  async function pull() {
    addLog(_('sync.pull') + '...', 'info');
    try {
      let r = await window.go.main.App.Pull();
      addLog(`${_('sync.pulled')}: ${r.changes?.length || 0} ${_('sync.changes')}`, 'success');
    } catch(e) { addLog(_('sync.pullFailed'), 'error'); }
  }

  async function loadDiff() {
    try {
      pending = await window.go.main.App.GetDiff();
    } catch(e) {}
  }

  onMount(() => { loadDiff(); });
</script>

<div>
  <h2 style="margin-bottom:16px;font-size:20px">{_('sync.title')}</h2>

  <div class="toolbar">
    <button class="btn-primary" on:click={push}>{_('sync.push')}</button>
    <button class="btn-secondary" on:click={pull}>{_('sync.pull')}</button>
    <button class="btn-ghost" on:click={loadDiff}>{_('sync.refresh')}</button>
  </div>

  <div class="card" style="margin-bottom:16px">
    <div class="card-title">{_('sync.pendingChanges')}</div>
    {#if pending.length}
      {#each pending as p}
        <div class="change-row">
          <span class="status-tag">{p.status}</span>
          <span>{p.path}</span>
          <span style="color:#555;font-size:11px">({p.tool})</span>
        </div>
      {/each}
    {:else}
      <p style="color:#555;font-size:13px">{_('sync.noPending')}</p>
    {/if}
  </div>

  <div class="card">
    <div class="card-title">{_('sync.syncLog')}</div>
    <div class="log">
      {#each log as l}
        <div class="log-entry {l.cls}">[{l.time}] {l.msg}</div>
      {/each}
      {#if log.length === 0}
        <div style="color:#555;font-size:13px">{_('sync.ready')}</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .toolbar { display:flex; gap:8px; margin-bottom:16px; }
  .btn-primary { background:#7c5cfc; color:#fff; border:none; padding:8px 16px; border-radius:6px; cursor:pointer; font-size:13px; }
  .btn-primary:hover { background:#6b4de6; }
  .btn-secondary { background:#2a2a3a; color:#ccc; border:1px solid #3a3a4a; padding:8px 16px; border-radius:6px; cursor:pointer; font-size:13px; }
  .btn-ghost { background:none; color:#888; border:1px solid #2a2a3a; padding:8px 16px; border-radius:6px; cursor:pointer; font-size:13px; }
  .card { background:#1a1a28; border:1px solid #2a2a3a; border-radius:8px; padding:16px; }
  .card-title { font-size:12px; color:#888; text-transform:uppercase; letter-spacing:.5px; margin-bottom:8px; }
  .change-row { display:flex; gap:12px; padding:4px 0; font-size:13px; align-items:center; }
  .status-tag { background:#7c5cfc20; color:#7c5cfc; padding:1px 6px; border-radius:3px; font-size:10px; text-transform:uppercase; }
  .log { background:#0f0f14; border-radius:6px; padding:12px; font-family:monospace; font-size:12px; max-height:300px; overflow-y:auto; }
  .log-entry { margin-bottom:3px; color:#888; }
  .log-entry.success { color:#4ade80; }
  .log-entry.error { color:#f87171; }
  .log-entry.info { color:#7c5cfc; }
</style>